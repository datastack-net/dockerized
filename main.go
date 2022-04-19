package main

import (
	"fmt"
	"github.com/compose-spec/compose-go/types"
	dockerized "github.com/datastack-net/dockerized/pkg"
	"github.com/datastack-net/dockerized/pkg/help"
	util "github.com/datastack-net/dockerized/pkg/util"
	"github.com/docker/compose/v2/pkg/api"
	"github.com/fatih/color"
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
			fmt.Printf("Service %s not found in compose file(s). Fallback to: %s.\n", commandName, image)
			fmt.Printf("  This command, if it exists, will not support version switching.\n")
			fmt.Printf("  See: https://github.com/jessfraz/dockerfiles\n")
		}
		return dockerized.DockerRun(image, runOptions, volumes)
	}

	return dockerized.DockerComposeRun(project, runOptions, volumes, serviceOptions...)
}

func parseArguments(args []string) (map[string]string, string, string, []string) {
	var options = []string{
		"--build",
		"-h",
		"--help",
		"-p",
		"--pull",
		"--digest",
		"--shell",
		"-v",
		"--verbose",
		"--version",
	}

	var optionsWithParameters = []string{
		"-p",
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
