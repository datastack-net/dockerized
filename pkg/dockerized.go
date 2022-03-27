package dockerized

import (
	"encoding/json"
	"github.com/compose-spec/compose-go/dotenv"
	"github.com/datastack-net/dockerized/pkg/util"
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
	"os"
	"os/signal"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"syscall"
)

// Determine which docker-compose file to use. Assumes .env files are already loaded.
func GetComposeFilePaths(dockerizedRoot string) []string {
	var composeFilePaths []string
	composeFilePath := os.Getenv("COMPOSE_FILE")
	if composeFilePath == "" {
		composeFilePaths = append(composeFilePaths, filepath.Join(dockerizedRoot, "docker-compose.yml"))
	} else {
		composePathSeparator := os.Getenv("COMPOSE_PATH_SEPARATOR")
		if composePathSeparator == "" {
			composePathSeparator = ";"
		}
		composeFilePaths = strings.Split(composeFilePath, composePathSeparator)
	}
	return composeFilePaths
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

func PrintCommandVersions(composeFilePaths []string, commandName string, verbose bool) error {
	project, err := GetProject(composeFilePaths)
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
		var packagePattern = regexp.MustCompile(`--package=(@?[^@]+)@([^\s]+)`)
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

func SetCommandVersion(composeFilePaths []string, commandName string, optionVerbose bool, commandVersion string) {
	rawProject, err := getRawProject(composeFilePaths)
	if err != nil {
		panic(err)
	}

	rawService, err := rawProject.GetService(commandName)

	var versionVariableExpected = strings.ReplaceAll(strings.ToUpper(commandName), "-", "_") + "_VERSION"
	var variablesUsed []string
	for _, variable := range ExtractVariables(rawService) {
		variablesUsed = append(variablesUsed, variable)
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

	if !util.Contains(variablesUsed, versionVariableExpected) {
		if len(versionVariablesUsed) == 1 {
			fmt.Printf("Error: To specify the version of %s, please set %s.\n",
				commandName,
				versionVariablesUsed[0],
			)
			os.Exit(1)
		} else if len(versionVariablesUsed) > 1 {
			fmt.Printf("Error: To specify the version of %s, please set one of:\n", commandName)
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

func LoadEnvFiles(hostCwd string, optionVerbose bool) error {
	var envFiles []string

	// Default
	dockerizedEnvFile := GetDockerizedRoot() + "/.env"
	envFiles = append(envFiles, dockerizedEnvFile)

	// Global overrides
	homeDir, _ := os.UserHomeDir()
	userGlobalDockerizedEnvFile := filepath.Join(homeDir, dockerizedEnvFileName)
	if _, err := os.Stat(userGlobalDockerizedEnvFile); err == nil {
		envFiles = append(envFiles, userGlobalDockerizedEnvFile)
	}

	// Local overrides
	if localDockerizedEnvFile, err := findLocalEnvFile(hostCwd); err == nil {
		envFiles = append(envFiles, localDockerizedEnvFile)
	}

	envFiles = unique(envFiles)

	if optionVerbose {
		for _, envFile := range envFiles {
			fmt.Printf("Loading: '%s'\n", envFile)
		}
	}

	dockerizedEnvMap, err := dotenv.ReadWithLookup(func(key string) (string, bool) {
		if os.Getenv(key) != "" {
			return os.Getenv(key), true
		} else {
			return "", false
		}
	}, dockerizedEnvFile)
	if err != nil {
		return err
	}

	envMap, err := dotenv.ReadWithLookup(func(key string) (string, bool) {
		if dockerizedEnvMap[key] != "" {
			return dockerizedEnvMap[key], true
		}
		var envValue = os.Getenv(key)
		if envValue != "" {
			return envValue, true
		} else {
			return "", false
		}
	}, envFiles...)
	if err != nil {
		return err
	}

	currentEnv := map[string]bool{}
	rawEnv := os.Environ()
	for _, rawEnvLine := range rawEnv {
		key := strings.Split(rawEnvLine, "=")[0]
		currentEnv[key] = true
	}

	for key, value := range envMap {
		if !currentEnv[key] {
			_ = os.Setenv(key, value)
		}
	}

	return nil
}

func dockerComposeRunAdHocService(service types.ServiceConfig, runOptions api.RunOptions) error {
	if service.Environment == nil {
		service.Environment = map[string]*string{}
	}
	return DockerComposeRun(&types.Project{
		Name: "dockerized",
		Services: []types.ServiceConfig{
			service,
		},
		WorkingDir: GetDockerizedRoot(),
	}, runOptions, []types.ServiceVolumeConfig{})
}

func DockerRun(image string, runOptions api.RunOptions, volumes []types.ServiceVolumeConfig) error {
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

func GetDockerizedRoot() string {
	if os.Getenv("DOCKERIZED_ROOT") != "" {
		return os.Getenv("DOCKERIZED_ROOT")
	}
	executable, err := os.Executable()
	if err != nil {
		panic("Cannot detect dockerized root directory: " + err.Error())
	}
	return filepath.Dir(filepath.Dir(executable))
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
func NormalizeEnvironment(dockerizedRoot string) {
	_ = os.Setenv("DOCKERIZED_ROOT", dockerizedRoot)
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

func getRawProject(composeFilePaths []string) (*types.Project, error) {
	options, err := cli.NewProjectOptions(composeFilePaths,
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

func GetProject(composeFilePaths []string) (*types.Project, error) {
	options, err := cli.NewProjectOptions([]string{},
		cli.WithOsEnv,
		cli.WithConfigFileEnv,
	)

	if err != nil {
		return nil, err
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

func DockerComposeBuild(composeFilePaths []string, buildOptions api.BuildOptions) error {
	project, err := GetProject(composeFilePaths)
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

func DockerComposeRun(project *types.Project, runOptions api.RunOptions, volumes []types.ServiceVolumeConfig) error {
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
		for _, argValue := range rawService.Build.Args {
			for _, argValueVariable := range ExtractVariablesFromString(*argValue) {
				usedVariables = append(usedVariables, argValueVariable)
			}
		}
	}
	for _, imageVariable := range ExtractVariablesFromString(rawService.Image) {
		usedVariables = append(usedVariables, imageVariable)
	}

	for _, entryPointArgument := range rawService.Entrypoint {
		for _, entryPointVariable := range ExtractVariablesFromString(entryPointArgument) {
			usedVariables = append(usedVariables, entryPointVariable)
		}
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
