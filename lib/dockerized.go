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
	"github.com/joho/godotenv"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"
)

var options = []string{
	"--shell",
	"--build",
	"-h",
	"--help",
	"-v",
	"--verbose",
}

func main() {
	normalizeEnvironment()

	dockerizedOptions, commandName, commandArgs := parseArguments()

	var optionHelp = contains(dockerizedOptions, "--help") || contains(dockerizedOptions, "-h")
	var optionVerbose = contains(dockerizedOptions, "--verbose") || contains(dockerizedOptions, "-v")

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

	hostName, _ := os.Hostname()
	hostCwd, _ := os.Getwd()
	hostCwdDirName := filepath.Base(hostCwd)
	containerCwd := "/host/" + hostCwdDirName

	runOptions := api.RunOptions{
		Name:    "dockerized",
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

	if contains(dockerizedOptions, "--shell") {
		runOptions.Entrypoint = []string{"/bin/sh"}
		runOptions.Command = []string{"-c", "$(which bash sh zsh | head -n 1)"}
	} else if contains(dockerizedOptions, "--build") {
		err := dockerCompose([]string{
			"-f", dockerizedDockerComposeFilePath,
			"build",
			commandName,
		})
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
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

	err = dockerComposeRun(dockerizedDockerComposeFilePath, runOptions, volumes)
	if err != nil {
		panic(err)
	}
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
	expectedErrorMessage := "no container found for project \"dockerized\": not found"
	if err == nil || api.IsNotFoundError(err) && err.Error() == expectedErrorMessage {
		return nil
	}
	return err
}

func getBackend() (*api.ServiceProxy, error) {
	dockerCli, err := command.NewDockerCli()
	if err != nil {
		return nil, err
	}
	dockerCliOpts := flags.NewClientOptions()
	err = dockerCli.Initialize(dockerCliOpts)
	if err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}

	var backend = api.NewServiceProxy()
	composeService := compose.NewComposeService(dockerCli)
	backend.WithService(composeService)
	return backend, nil
}

func dockerComposeRun(dockerComposeFilePath string, runOptions api.RunOptions, volumes []types.ServiceVolumeConfig) error {
	backend, err := getBackend()
	if err != nil {
		return err
	}

	project, err := getProject(dockerComposeFilePath)
	if err != nil {
		return err
	}

	ctx, _ := newSigContext()

	serviceName := runOptions.Service

	service, err := project.GetService(serviceName)
	if service.CustomLabels == nil {
		service.CustomLabels = map[string]string{}
	}
	service.Hostname = "dockerized-" + serviceName
	service.Volumes = append(service.Volumes, volumes...)

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

	fmt.Println("Usage: dockerized [options] <command> [args]")
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
	fmt.Println("  --help, -h Show this help")
	fmt.Println("  --shell    Start a shell inside the command container. Similar to `docker run --entrypoint=sh`.")
	fmt.Println("  --build    Rebuild the container before running it.")

	fmt.Println()
	fmt.Println("Args:")
	fmt.Println("  <command> [args]  Arguments are passed to the command within the container.")

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

func dockerCompose(composeRunArgs []string) error {
	cmd := exec.Command("docker-compose", composeRunArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Run()
	if err != nil {
		return err
	}
	return nil
}
