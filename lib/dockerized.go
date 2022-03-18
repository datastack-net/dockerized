package main

import (
	"context"
	"fmt"
	"github.com/compose-spec/compose-go/cli"
	"github.com/compose-spec/compose-go/types"
	"github.com/docker/cli/cli/command"
	"github.com/docker/cli/cli/flags"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/docker/compose/v2/pkg/compose"
	"github.com/fatih/color"
	"github.com/joho/godotenv"
	"os"
	"os/signal"
	"path/filepath"
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

	dockerizedOptions, commandName, commandArgs := parseArguments()

	var optionHelp = contains(dockerizedOptions, "--help") || contains(dockerizedOptions, "-h")
	var optionVerbose = contains(dockerizedOptions, "--verbose") || contains(dockerizedOptions, "-v")
	var optionShell = contains(dockerizedOptions, "--shell")
	var optionBuild = contains(dockerizedOptions, "--build")
	var optionVersion = contains(dockerizedOptions, "--version")

	if optionVersion {
		fmt.Printf("dockerized %s\n", Version)
		os.Exit(0)
	}

	dockerizedRoot := getDockerizedRoot()
	dockerizedDockerComposeFilePath := os.Getenv("COMPOSE_FILE")
	if dockerizedDockerComposeFilePath != "" {
		if optionVerbose {
			fmt.Println("COMPOSE_FILE: ", dockerizedDockerComposeFilePath)
		}
	} else {
		dockerizedDockerComposeFilePath = filepath.Join(dockerizedRoot, "docker-compose.yml")
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

	project, err := getProject(dockerizedDockerComposeFilePath)
	if err != nil {
		panic(err)
	}

	hostName, _ := os.Hostname()
	hostCwd, _ := os.Getwd()
	hostCwdDirName := filepath.Base(hostCwd)
	containerCwd := "/host/" + hostCwdDirName

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
				fmt.Printf("Ignoring arguments: %s\n", commandArgs[0])
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

		var preferredShells = "bash zsh sh"
		var cmdPrintWelcome = fmt.Sprintf("echo -e '%s'", color.YellowString(welcomeMessage))
		var cmdLaunchShell = fmt.Sprintf("$(which %[1]s 2> /dev/null || command -v %[1]s | head -n 1)", preferredShells)

		runOptions.Environment = append(runOptions.Environment, "PS1="+ps1)
		runOptions.Entrypoint = []string{"/bin/sh"}
		runOptions.Command = []string{"-c", fmt.Sprintf("%s; %s", cmdPrintWelcome, cmdLaunchShell)}
	}

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
			panic(err)
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

func dockerComposeRunAdHocService(service types.ServiceConfig, runOptions api.RunOptions) error {
	if service.Environment == nil {
		service.Environment = map[string]*string{}
	}
	return dockerComposeRun(&types.Project{
		Name: "dockerized",
		Services: []types.ServiceConfig{
			service,
		},
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
	return filepath.Dir(filepath.Dir(os.Args[0]))
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

	backend, err := getBackend()
	if err != nil {
		return err
	}
	ctx, _ := newSigContext()
	return backend.Build(ctx, project, buildOptions)
}

func dockerComposeRun(project *types.Project, runOptions api.RunOptions, volumes []types.ServiceVolumeConfig) error {
	ctx, _ := newSigContext()

	serviceName := runOptions.Service

	service, err := project.GetService(serviceName)
	if service.CustomLabels == nil {
		service.CustomLabels = map[string]string{}
	}
	service.Volumes = append(service.Volumes, volumes...)

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

	fmt.Println("Usage: dockerized [options] <command> [arguments]")
	fmt.Println("")
	fmt.Println("Commands:")

	services := project.ServiceNames()
	sort.Strings(services)
	for _, service := range services {
		if service[0] == '_' {
			continue
		}
		fmt.Printf("  %s\n", service)
	}

	fmt.Println("")

	fmt.Println("Options:")
	fmt.Println("      --build    Rebuild the container before running it.")
	fmt.Println("      --shell    Start a shell inside the command container. Similar to `docker run --entrypoint=sh`.")
	fmt.Println("  -v, --verbose  Log what dockerized is doing.")
	fmt.Println("  -h, --help     Show this help.")

	fmt.Println()
	fmt.Println("Arguments:")
	fmt.Println("  All arguments after <command> are passed to the command itself.")

	return nil
}

func parseArguments() ([]string, string, []string) {
	commandName := ""
	var commandArgs []string
	var dockerizedOptions []string
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
	return dockerizedOptions, commandName, commandArgs
}
