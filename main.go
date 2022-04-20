package main

import (
	"fmt"
	"github.com/compose-spec/compose-go/types"
	dockerized "github.com/datastack-net/dockerized/pkg"
	"github.com/datastack-net/dockerized/pkg/help"
	"github.com/datastack-net/dockerized/pkg/labels"
	util "github.com/datastack-net/dockerized/pkg/util"
	"github.com/docker/compose/v2/pkg/api"
	"os"
	"path/filepath"
	"strings"
)

var Version string

var contains = util.Contains
var hasKey = util.HasKey

func main() {
	err, exitCode := RunCli(os.Args[1:])
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	os.Exit(exitCode)
}

func RunCli(args []string) (err error, exitCode int) {
	dockerizedOptions, commandName, commandVersion, commandArgs := parseArguments(args)

	var optionHelp = hasKey(dockerizedOptions, "--help") || hasKey(dockerizedOptions, "-h")
	var optionVerbose = hasKey(dockerizedOptions, "--verbose") || hasKey(dockerizedOptions, "-v")
	var optionShell = hasKey(dockerizedOptions, "--shell")
	var optionBuild = hasKey(dockerizedOptions, "--build")
	var optionPull = hasKey(dockerizedOptions, "--pull")
	var optionVersion = hasKey(dockerizedOptions, "--version")
	var optionDigest = hasKey(dockerizedOptions, "--digest")
	var optionPort = hasKey(dockerizedOptions, "-p")
	var optionEntrypoint = hasKey(dockerizedOptions, "--entrypoint")
	var optionCommands = hasKey(dockerizedOptions, "--commands")

	dockerizedRoot := dockerized.GetDockerizedRoot()
	dockerized.NormalizeEnvironment(dockerizedRoot)

	if optionVerbose {
		fmt.Printf("Dockerized root: %s\n", dockerizedRoot)
	}

	if optionVersion {
		fmt.Printf("dockerized %s\n", Version)
		return nil, 0
	}

	hostCwd, _ := os.Getwd()
	err = dockerized.LoadEnvFiles(hostCwd, optionVerbose)
	if err != nil {
		return err, 1
	}

	composeFilePaths := dockerized.GetComposeFilePaths(dockerizedRoot)

	if optionVerbose {
		fmt.Printf("Compose files: %s\n", strings.Join(composeFilePaths, ", "))
	}

	if commandName == "" || optionHelp {
		err := help.Help()
		if err != nil {
			return err, 1
		}
		if optionHelp {
			return nil, 0
		} else {
			return nil, 1
		}
	}

	if commandVersion != "" {
		if commandVersion == "?" {
			err = dockerized.PrintCommandVersions(commandName, optionVerbose)
			if err != nil {
				return err, 1
			} else {
				return nil, 0
			}
		} else {
			dockerized.SetCommandVersion(composeFilePaths, commandName, optionVerbose, commandVersion)
		}
	}

	project, err := dockerized.GetProject()
	if err != nil {
		return err, 1
	}

	hostName, _ := os.Hostname()
	hostCwdDirName := filepath.Base(hostCwd)
	containerCwd := "/host"
	if hostCwdDirName != "\\" {
		containerCwd += "/" + hostCwdDirName
	}

	if optionCommands {
		for _, service := range project.Services {
			fmt.Printf("%s\n", service.Name)
		}
		return nil, 0
	}

	runOptions := api.RunOptions{
		Service: commandName,
		Environment: []string{
			"HOST_HOSTNAME=" + hostName,
		},
		Command:    commandArgs,
		AutoRemove: true,
		Tty:        false,
		WorkingDir: containerCwd,
	}

	var serviceOptions []func(config *types.ServiceConfig) error

	if optionPort {
		var port = dockerizedOptions["-p"]
		if port == "" {
			return fmt.Errorf("port option requires a port number"), 1
		}
		if optionVerbose {
			fmt.Printf("Mapping port: %s\n", port)
		}
		serviceOptions = append(serviceOptions, func(config *types.ServiceConfig) error {
			if !strings.ContainsRune(port, ':') {
				port = port + ":" + port
			}
			portConfig, err := types.ParsePortConfig(port)
			if err != nil {
				return err
			}
			config.Ports = portConfig
			return nil
		})
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
		err := dockerized.DockerComposeBuild(api.BuildOptions{
			Services: []string{commandName},
		})

		if err != nil {
			return err, 1
		}
	} else if optionPull {
		err = dockerized.Pull(commandName)
		if err != nil {
			return err, 1
		}
		if !optionDigest {
			return nil, 0
		}
	}

	if optionDigest {
		digest, err := dockerized.GetDigest(commandName)

		if err != nil {
			return err, 0
		}
		fmt.Println(digest)
		return nil, 0
	}

	if optionShell && optionEntrypoint {
		return fmt.Errorf("--shell and --entrypoint are mutually exclusive"), 1
	}

	if optionShell {
		if optionVerbose {
			fmt.Printf("Setting up shell in container for %s...\n", commandName)
		}

		//var ps1 = fmt.Sprintf(
		//	"%s %s:\\w \\$ ",
		//	color.BlueString("dockerized %s", commandName),
		//	color.New(color.FgHiBlue).Add(color.Bold).Sprintf("\\u@\\h"),
		//)
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

		var shell = "sh"
		if service.Labels[labels.Shell] != "" {
			shell = service.Labels[labels.Shell]
		}

		runOptions.Entrypoint = []string{shell}
		runOptions.Command = commandArgs
	}

	if optionEntrypoint {
		var entrypoint = dockerizedOptions["--entrypoint"]
		if optionVerbose {
			fmt.Printf("Setting entrypoint to %s\n", entrypoint)
		}
		runOptions.Entrypoint = strings.Split(entrypoint, " ")
	}

	if optionVerbose {
		fmt.Printf("Entrypoint: %s\n", runOptions.Entrypoint)
		fmt.Printf("Command:    %s\n", runOptions.Command)
	}

	if !contains(project.ServiceNames(), commandName) {
		image := "r.j3ss.co/" + commandName
		if optionVerbose {
			fmt.Printf("Service %s not found in compose file(s). Fallback to: %s.\n", commandName, image)
			fmt.Printf("  This command, if it exists, will not support version switching.\n")
			fmt.Printf("  See: https://github.com/jessfraz/dockerfiles\n")
		}
		return dockerized.DockerRun(image, runOptions, volumes)
	}

	return dockerized.DockerComposeRun(project, runOptions, volumes, optionVerbose, serviceOptions...)
}

func parseArguments(args []string) (map[string]string, string, string, []string) {
	var options = []string{
		"--build",
		"-h",
		"--help",
		"-p",
		"--pull",
		"--digest",
		"--commands",
		"--shell",
		"--entrypoint",
		"-v",
		"--verbose",
		"--version",
	}

	var optionsWithParameters = []string{
		"-p",
		"--entrypoint",
	}

	commandName := ""
	var commandArgs []string
	var dockerizedOptions []string
	var commandVersion string

	var optionMap = make(map[string]string)
	var optionBefore = ""

	for _, arg := range args {
		if arg[0] == '-' && commandName == "" {
			if util.Contains(options, arg) {
				var option = arg
				dockerizedOptions = append(dockerizedOptions, option)
				optionBefore = option
				optionMap[option] = ""
			} else {
				fmt.Println("Unknown option:", arg)
				os.Exit(1)
			}
		} else {
			if contains(optionsWithParameters, optionBefore) {
				optionMap[optionBefore] = arg
			} else if commandName == "" {
				commandName = arg
			} else {
				commandArgs = append(commandArgs, arg)
			}
			optionBefore = ""
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
	return optionMap, commandName, commandVersion, commandArgs
}
