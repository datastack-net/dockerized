package main

import (
	"encoding/json"
	"github.com/hashicorp/go-version"
	"io"
	"net/http"
)

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/loader"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/docker/distribution/reference"
	"github.com/docker/hub-tool/pkg/hub"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
)

var Version string

var options = []string{
	"--shell",
	"--build",
	"-h",
	"--help",
	"-v",
	"--verbose",
	"--version",
}

func main() {
	normalizeEnvironment()

	dockerizedOptions, commandName, commandVersion, commandArgs := parseArguments()

	var optionHelp = contains(dockerizedOptions, "--help") || contains(dockerizedOptions, "-h")
	var optionVerbose = contains(dockerizedOptions, "--verbose") || contains(dockerizedOptions, "-v")
	var optionShell = contains(dockerizedOptions, "--shell")
	var optionBuild = contains(dockerizedOptions, "--build")
	var optionVersion = contains(dockerizedOptions, "--version")

	if optionVersion {
		fmt.Printf("dockerized %s\n", Version)
		os.Exit(0)
	}

	dockerizedDockerComposeFilePath := os.Getenv("COMPOSE_FILE")
	if dockerizedDockerComposeFilePath == "" {
		dockerizedRoot := getDockerizedRoot()
		dockerizedDockerComposeFilePath = filepath.Join(dockerizedRoot, "docker-compose.yml")
	}

	if optionVerbose {
		fmt.Println("Dockerized docker-compose file: ", dockerizedDockerComposeFilePath)
	}

	if commandName == "" || optionHelp {
		err := help(dockerizedDockerComposeFilePath)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		if optionHelp {
			os.Exit(0)
		} else {
			os.Exit(1)
		}
	}

	hostCwd, _ := os.Getwd()
	var err = loadEnvFiles(hostCwd, optionVerbose)
	if err != nil {
		panic(err)
	}

	if commandVersion != "" {
		if commandVersion == "?" {
			err = PrintCommandVersions(dockerizedDockerComposeFilePath, commandName, optionVerbose)
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			} else {
				os.Exit(0)
			}
		} else {
			setCommandVersion(dockerizedDockerComposeFilePath, commandName, optionVerbose, commandVersion)
		}
	}

	project, err := getProject(dockerizedDockerComposeFilePath)
	if err != nil {
		panic(err)
	}

	hostName, _ := os.Hostname()
	hostCwdDirName := filepath.Base(hostCwd)
	containerCwd := "/host"
	if hostCwdDirName != "\\" {
		containerCwd += "/" + hostCwdDirName
	}

	runOptions := api.RunOptions{
		Service: commandName,
		Environment: []string{
			"HOST_HOSTNAME=" + hostName,
		},
		Command:    commandArgs,
		AutoRemove: true,
		Tty:        true,
		WorkingDir: containerCwd,
	}

	volumes := []types.ServiceVolumeConfig{
		{
			Type:   "bind",
			Source: hostCwd,
			Target: containerCwd,
		}}

	if optionBuild {
		if optionVerbose {
			fmt.Printf("Building container image for %s...\n", commandName)
		}
		err := dockerComposeBuild(dockerizedDockerComposeFilePath, api.BuildOptions{
			Services: []string{commandName},
		})

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
	}

	if optionShell {
		if optionVerbose {
			fmt.Printf("Opening shell in container for %s...\n", commandName)

			if len(commandArgs) > 0 {
				fmt.Printf("Passing arguments to shell: %s\n", commandArgs)
			}
		}

		var ps1 = fmt.Sprintf(
			"%s %s:\\w \\$ ",
			color.BlueString("dockerized %s", commandName),
			color.New(color.FgHiBlue).Add(color.Bold).Sprintf("\\u@\\h"),
		)
		var welcomeMessage = "Welcome to dockerized shell. Type 'exit' or press Ctrl+D to exit.\n"
		welcomeMessage += "Mounted volumes:\n"

		for _, volume := range volumes {
			welcomeMessage += fmt.Sprintf("  %s -> %s\n", volume.Source, volume.Target)
		}
		service, err := project.GetService(commandName)
		if err == nil {
			for _, volume := range service.Volumes {
				welcomeMessage += fmt.Sprintf("  %s -> %s\n", volume.Source, volume.Target)
			}
		}
		welcomeMessage = strings.ReplaceAll(welcomeMessage, "\\", "\\\\")

		shells := []string{
			"bash",
			"zsh",
			"sh",
		}
		var shellDetectionCommands []string
		for _, shell := range shells {
			shellDetectionCommands = append(shellDetectionCommands, "command -v "+shell)
		}
		for _, shell := range shells {
			shellDetectionCommands = append(shellDetectionCommands, "which "+shell)
		}

		var cmdPrintWelcome = fmt.Sprintf("echo '%s'", color.YellowString(welcomeMessage))
		var cmdLaunchShell = fmt.Sprintf("$(%s)", strings.Join(shellDetectionCommands, " || "))

		runOptions.Environment = append(runOptions.Environment, "PS1="+ps1)
		runOptions.Entrypoint = []string{"/bin/sh"}

		if len(commandArgs) > 0 {
			runOptions.Command = []string{"-c", fmt.Sprintf("%s; %s \"%s\"", cmdPrintWelcome, cmdLaunchShell, strings.Join(commandArgs, "\" \""))}
		} else {
			runOptions.Command = []string{"-c", fmt.Sprintf("%s; %s", cmdPrintWelcome, cmdLaunchShell)}
		}
	}

	if !contains(project.ServiceNames(), commandName) {
		image := "r.j3ss.co/" + commandName
		if optionVerbose {
			fmt.Printf("Service %s not found in %s. Fallback to: %s.\n", commandName, dockerizedDockerComposeFilePath, image)
			fmt.Printf("  This command, if it exists, will not support version switching.\n")
			fmt.Printf("  See: https://github.com/jessfraz/dockerfiles\n")
		}
		err := dockerRun(image, runOptions, volumes)
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	err = dockerComposeRun(project, runOptions, volumes)
	if err != nil {
		if optionVerbose {
			fmt.Println(err)
		}
		os.Exit(1)
	}
}

func getNpmPackageVersions(packageName string) ([]string, error) {
	var registryUrl = "https://registry.npmjs.org/" + packageName
	request, err := http.NewRequest(http.MethodGet, registryUrl, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Add("Accept", "application/vnd.npm.install-v1+json")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(response.Body)

	// parse json
	var registryResponse map[string]interface{}
	err = json.NewDecoder(response.Body).Decode(&registryResponse)
	if err != nil {
		return nil, err
	}
	// read versions
	var versions = registryResponse["versions"].(map[string]interface{})
	var versionKeys = make([]string, 0, len(versions))
	for k := range versions {
		versionKeys = append(versionKeys, k)
	}
	sort.Strings(versionKeys)
	return versionKeys, nil
}

func PrintCommandVersions(dockerizedDockerComposeFilePath string, commandName string, verbose bool) error {
	project, err := getProject(dockerizedDockerComposeFilePath)
	if err != nil {
		return err
	}

	service, err := project.GetService(commandName)
	if err != nil {
		return err
	}

	var semanticVersions []string
	var rawVersions []string

	isNpmPackage := len(service.Entrypoint) > 0 && service.Entrypoint[0] == "npx"
	if isNpmPackage {
		var packagePattern = regexp.MustCompile(`--package=([^@]+)@([^\s]+)`)
		packageArgument := service.Entrypoint[1]
		packageMatch := packagePattern.FindStringSubmatch(packageArgument)
		var packageName = packageMatch[1]
		rawVersions, err = getNpmPackageVersions(packageName)
		if err != nil {
			return err
		}
	} else {
		// isDockerHubImage
		if service.Build != nil {
			fmt.Printf("Cannot determine versions for command %s because it has a build step.\n", commandName)
			os.Exit(1)
		}

		ref, err := reference.ParseDockerRef(service.Image)
		if err != nil {
			return err
		}

		refDomain := reference.Domain(ref)

		if refDomain != "docker.io" {
			fmt.Printf("Listing versions for commands is currently only supported for docker.io images.\n")
			os.Exit(1)
		}

		refPath := reference.Path(ref)

		hubClient, err := hub.NewClient(hub.WithAllElements())

		if err != nil {
			return err
		}
		tags, _, err := hubClient.GetTags(refPath)
		if err != nil {
			return err
		}

		for _, tag := range tags {
			var tagParts = strings.Split(tag.Name, ":")
			var tagVersion = tagParts[1]
			rawVersions = append(rawVersions, tagVersion)
		}
	}
	sort.Strings(rawVersions)
	rawVersions = unique(rawVersions)
	semanticVersions, err = getSemanticVersions(rawVersions)
	sortVersions(semanticVersions)
	semanticVersions = unique(semanticVersions)

	if verbose {
		fmt.Printf("\n")
		fmt.Printf("Raw versions:\n")
		for _, rawVersion := range rawVersions {
			fmt.Printf("%s\n", rawVersion)
		}
		fmt.Printf("\n")
	}

	if len(semanticVersions) == 0 {
		fmt.Printf("No parseable versions found for command %s.\n", commandName)
		if len(rawVersions) > 0 {
			fmt.Printf("Found: %s\n", strings.Join(rawVersions, ", "))
		}
		os.Exit(1)
	}

	var versionGroups = make(map[string][]string)
	var versionGroupKeys []string
	for _, semanticVersion := range semanticVersions {
		var v, err = version.NewVersion(semanticVersion)
		if err != nil {
			return err
		}

		var versionGroup string
		var segments = v.Segments()
		if len(segments) >= 2 {
			versionGroup = fmt.Sprintf("%d.%d", segments[0], segments[1])
		} else {
			versionGroup = fmt.Sprintf("%d.0", segments[0])
		}
		versionGroups[versionGroup] = append(versionGroups[versionGroup], semanticVersion)
		versionGroupKeys = append(versionGroupKeys, versionGroup)
	}
	versionGroupKeys = unique(versionGroupKeys)
	sortVersions(versionGroupKeys)

	for _, versionGroup := range versionGroupKeys {
		var versions = versionGroups[versionGroup]
		fmt.Printf("%s\n", strings.Join(versions, ", "))
	}
	return nil
}

func getSemanticVersions(rawVersions []string) ([]string, error) {
	var semanticVersions []string
	for _, rawVersion := range rawVersions {
		var semanticVersion = regexp.MustCompile(`^v?(\d+(\.\d+)*$)`).FindStringSubmatch(rawVersion)
		if semanticVersion != nil {
			semanticVersions = append(semanticVersions, semanticVersion[1])
		}
	}
	return semanticVersions, nil
}

func sortVersions(versions []string) {
	sort.Slice(versions, func(i, j int) bool {
		v1, e1 := version.NewVersion(versions[i])
		v2, e2 := version.NewVersion(versions[j])
		return e1 == nil && e2 == nil && v1.LessThan(v2)
	})
}

func setCommandVersion(dockerizedDockerComposeFilePath string, commandName string, optionVerbose bool, commandVersion string) {
	rawProject, err := getRawProject(dockerizedDockerComposeFilePath)
	if err != nil {
		panic(err)
	}

	rawService, err := rawProject.GetService(commandName)

	var versionVariableExpected = strings.ReplaceAll(strings.ToUpper(commandName), "-", "_") + "_VERSION"
	var variablesUsed []string
	for _, variable := range ExtractVariables(rawService) {
		variablesUsed = append(variablesUsed, variable)
	}

	for _, entryPointArgument := range rawService.Entrypoint {
		for _, entryPointVariable := range ExtractVariablesFromString(entryPointArgument) {
			variablesUsed = append(variablesUsed, entryPointVariable)
		}
	}

	if len(variablesUsed) == 0 {
		fmt.Printf("Error: Version selection for %s is currently not supported.\n", commandName)
		os.Exit(1)
	}

	var versionVariablesUsed []string
	for _, variable := range variablesUsed {
		if strings.HasSuffix(variable, "_VERSION") {
			versionVariablesUsed = append(versionVariablesUsed, variable)
		}
	}
	versionKey := versionVariableExpected

	if !contains(variablesUsed, versionVariableExpected) {
		if len(versionVariablesUsed) == 1 {
			fmt.Printf("Error: To specify the version of %s, please set %s.\n",
				commandName,
				versionVariablesUsed[0],
			)
			os.Exit(1)
		} else if len(versionVariablesUsed) > 1 {
			fmt.Println("Multiple version variables found:")
			for _, versionVariable := range versionVariablesUsed {
				fmt.Println("  " + versionVariable)
			}
			os.Exit(1)
		}
	}

	if optionVerbose {
		fmt.Printf("Setting %s to %s...\n", versionKey, commandVersion)
	}
	err = os.Setenv(versionKey, commandVersion)
	if err != nil {
		panic(err)
	}
}

func loadEnvFiles(hostCwd string, optionVerbose bool) error {
	homeDir, _ := os.UserHomeDir()
	userGlobalDockerizedEnvFile := filepath.Join(homeDir, dockerizedEnvFileName)
	localDockerizedEnvFile, err := findLocalEnvFile(hostCwd)

	var envFiles []string

	if _, err := os.Stat(userGlobalDockerizedEnvFile); err == nil {
		envFiles = append(envFiles, userGlobalDockerizedEnvFile)
	}
	if err == nil && !contains(envFiles, localDockerizedEnvFile) {
		envFiles = append(envFiles, localDockerizedEnvFile)
	}

	if optionVerbose {
		// Print it in order of priority (lowest to highest)
		for _, envFile := range envFiles {
			fmt.Println("Loading: ", envFile)
		}
	}
	// Load in reverse. GoDotEnv does not override vars, this allows runtime env-vars to override the env files.
	for i := len(envFiles) - 1; i >= 0; i-- {
		err := godotenv.Load(envFiles[i])
		if err != nil {
			return err
		}
	}
	return nil
}

func dockerComposeRunAdHocService(service types.ServiceConfig, runOptions api.RunOptions) error {
	if service.Environment == nil {
		service.Environment = map[string]*string{}
	}
	return dockerComposeRun(&types.Project{
		Name: "dockerized",
		Services: []types.ServiceConfig{
			service,
		},
		WorkingDir: getDockerizedRoot(),
	}, runOptions, []types.ServiceVolumeConfig{})
}

func dockerRun(image string, runOptions api.RunOptions, volumes []types.ServiceVolumeConfig) error {
	// Couldn't get 'docker run' to work, so instead define a Docker Compose Service and run that.
	// This coincidentally allows re-using the same code for both 'docker run' and 'docker-compose run'
	// - ContainerCreate is simple, but the logic to attach to it is very complex, and not exposed by the Docker SDK.
	// - Using [container.NewRunCommand] didn't work due to dependency compatibility issues.
	return dockerComposeRunAdHocService(types.ServiceConfig{
		Name:    runOptions.Service,
		Image:   image,
		Volumes: volumes,
	}, runOptions)
}

var dockerizedEnvFileName = "dockerized.env"

func getDockerizedRoot() string {
	executable, err := os.Executable()
	if err != nil {
		panic("Cannot detect dockerized root directory: " + err.Error())
	}
	return filepath.Dir(filepath.Dir(executable))
}

func contains(s []string, str string) bool {
	for _, v := range s {
		if v == str {
			return true
		}
	}

	return false
}

func findLocalEnvFile(path string) (string, error) {
	envFilePath := ""
	for i := 0; i < 10; i++ {
		envFilePath = filepath.Join(path, dockerizedEnvFileName)
		if _, err := os.Stat(envFilePath); err == nil {
			return envFilePath, nil
		}
		path = filepath.Dir(path)
	}
	return "", fmt.Errorf("no local %s found", dockerizedEnvFileName)
}
func normalizeEnvironment() {
	homeDir, _ := os.UserHomeDir()
	if os.Getenv("HOME") == "" {
		_ = os.Setenv("HOME", homeDir)
	}
}

func newSigContext() (context.Context, func()) {
	ctx, cancel := context.WithCancel(context.Background())
	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-s
		cancel()
	}()
	return ctx, cancel
}

func getRawProject(dockerComposeFilePath string) (*types.Project, error) {
	options, err := cli.NewProjectOptions([]string{
		dockerComposeFilePath,
	},
		cli.WithInterpolation(false),
		cli.WithLoadOptions(func(l *loader.Options) {
			l.SkipValidation = true
			l.SkipConsistencyCheck = true
			l.SkipNormalization = true
		}),
	)

	if err != nil {
		return nil, nil
	}

	return cli.ProjectFromOptions(options)
}

func getProject(dockerComposeFilePath string) (*types.Project, error) {
	options, err := cli.NewProjectOptions([]string{
		dockerComposeFilePath,
	},
		cli.WithDotEnv,
		cli.WithOsEnv,
		cli.WithConfigFileEnv,
	) //, cli.WithDefaultConfigPath

	if err != nil {
		return nil, nil
	}

	return cli.ProjectFromOptions(options)
}

func dockerComposeUpNetworkOnly(backend *api.ServiceProxy, ctx context.Context, project *types.Project) error {
	project.Services = []types.ServiceConfig{}
	upOptions := api.UpOptions{
		Create: api.CreateOptions{
			Services:      []string{},
			RemoveOrphans: true,
			Recreate:      "always",
		},
	}
	err := backend.Up(ctx, project, upOptions)

	// docker compose up will return error if there is no service to start, but the network will have been created.
	expectedErrorMessage := "no container found for project \"" + project.Name + "\": not found"
	if err == nil || api.IsNotFoundError(err) && err.Error() == expectedErrorMessage {
		return nil
	}
	return err
}

func getDockerCli() (*command.DockerCli, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}
	dockerCliOpts := flags.NewClientOptions()
	err = dockerCli.Initialize(dockerCliOpts)
	if err != nil {
		return nil, err
	}
	return dockerCli, nil
}

func getBackend() (*api.ServiceProxy, error) {
	dockerCli, err := getDockerCli()
	if err != nil {
		return nil, err
	}

	var backend = api.NewServiceProxy()
	composeService := compose.NewComposeService(dockerCli)
	backend.WithService(composeService)
	return backend, nil
}

func dockerComposeBuild(dockerComposeFilePath string, buildOptions api.BuildOptions) error {
	project, err := getProject(dockerComposeFilePath)
	if err != nil {
		return err
	}
	err = os.Chdir(project.WorkingDir)
	if err != nil {
		return err
	}

	backend, err := getBackend()
	if err != nil {
		return err
	}
	ctx, _ := newSigContext()
	return backend.Build(ctx, project, buildOptions)
}

func dockerComposeRun(project *types.Project, runOptions api.RunOptions, volumes []types.ServiceVolumeConfig) error {
	err := os.Chdir(project.WorkingDir)
	if err != nil {
		return err
	}
	ctx, _ := newSigContext()

	serviceName := runOptions.Service

	service, err := project.GetService(serviceName)
	if service.CustomLabels == nil {
		service.CustomLabels = map[string]string{}
	}

	stopGracePeriod := types.Duration(1)
	service.Volumes = append(service.Volumes, volumes...)
	service.StopGracePeriod = &stopGracePeriod
	service.StdinOpen = true

	backend, err := getBackend()
	if err != nil {
		return err
	}

	err = dockerComposeUpNetworkOnly(backend, ctx, project)
	if err != nil {
		return err
	}

	project.Services = []types.ServiceConfig{service}

	exitCode, err := backend.RunOneOffContainer(ctx, project, runOptions)
	if err != nil {
		return err
	}
	if exitCode != 0 {
		return fmt.Errorf("docker-compose exited with code %d", exitCode)
	}
	return nil
}

func help(dockerComposeFilePath string) error {
	project, err := getProject(dockerComposeFilePath)
	if err != nil {
		return err
	}

	fmt.Println("Usage: dockerized [options] <command>[:version] [arguments]")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  dockerized go")
	fmt.Println("  dockerized go:1.8 build")
	fmt.Println("  dockerized --shell go")
	fmt.Println("  dockerized go:?")
	fmt.Println("")

	fmt.Println("Commands:")
	services := project.ServiceNames()
	sort.Strings(services)
	for _, service := range services {
		fmt.Printf("  %s\n", service)
	}
	fmt.Println()

	fmt.Println("Options:")
	fmt.Println("      --build    Rebuild the container before running it.")
	fmt.Println("      --shell    Start a shell inside the command container. Similar to `docker run --entrypoint=sh`.")
	fmt.Println("  -v, --verbose  Log what dockerized is doing.")
	fmt.Println("  -h, --help     Show this help.")
	fmt.Println()

	fmt.Println("Version:")
	fmt.Println("  :<version>      The version of the command to run, e.g. 1, 1.8, 1.8.1.")
	fmt.Println("  :?              List all available versions. E.g. `dockerized go:?`")
	fmt.Println("  :               Same as ':?' .")
	fmt.Println()

	fmt.Println("Arguments:")
	fmt.Println("  All arguments after <command> are passed to the command itself.")

	return nil
}

func parseArguments() ([]string, string, string, []string) {
	commandName := ""
	var commandArgs []string
	var dockerizedOptions []string
	var commandVersion string
	for _, arg := range os.Args[1:] {
		if arg[0] == '-' && commandName == "" {
			if contains(options, arg) {
				dockerizedOptions = append(dockerizedOptions, arg)
			} else {
				fmt.Println("Unknown option:", arg)
				os.Exit(1)
			}
		} else {
			if commandName == "" {
				commandName = arg
			} else {
				commandArgs = append(commandArgs, arg)
			}
		}
	}
	if strings.ContainsRune(commandName, ':') {
		commandSplit := strings.Split(commandName, ":")
		commandName = commandSplit[0]
		commandVersion = commandSplit[1]
		if commandVersion == "" {
			commandVersion = "?"
		}
	}
	return dockerizedOptions, commandName, commandVersion, commandArgs
}

func unique(s []string) []string {
	keys := make(map[string]bool)
	var list []string
	for _, entry := range s {
		if _, value := keys[entry]; !value {
			keys[entry] = true
			list = append(list, entry)
		}
	}
	return list
}

func ExtractVariables(rawService types.ServiceConfig) []string {
	var usedVariables []string
	for envKey := range rawService.Environment {
		usedVariables = append(usedVariables, envKey)
	}
	if rawService.Build != nil {
		for argKey := range rawService.Build.Args {
			usedVariables = append(usedVariables, argKey)
		}
	}
	for _, imageVariable := range ExtractVariablesFromString(rawService.Image) {
		usedVariables = append(usedVariables, imageVariable)
	}

	usedVariables = unique(usedVariables)
	sort.Strings(usedVariables)

	return usedVariables
}

func ExtractVariablesFromString(value string) []string {
	var usedVariables []string
	pattern := regexp.MustCompile(`\${([^}]+)}`)
	for _, match := range pattern.FindAllStringSubmatch(value, -1) {
		usedVariables = append(usedVariables, match[1])
	}
	return usedVariables
}
