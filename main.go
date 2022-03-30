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

func main() {
	err, exitCode := RunCli(os.Args[1:])
	if err != nil {
		fmt.Printf("%s\n", err)
	}
	os.Exit(exitCode)
}

func RunCli(args []string) (err error, exitCode int) {
	dockerizedOptions, commandName, commandVersion, commandArgs := parseArguments(args)

	var optionHelp = contains(dockerizedOptions, "--help") || contains(dockerizedOptions, "-h")
	var optionVerbose = contains(dockerizedOptions, "--verbose") || contains(dockerizedOptions, "-v")
	var optionShell = contains(dockerizedOptions, "--shell")
	var optionBuild = contains(dockerizedOptions, "--build")
	var optionVersion = contains(dockerizedOptions, "--version")

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
		err := help.Help(composeFilePaths)
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
			err = dockerized.PrintCommandVersions(composeFilePaths, commandName, optionVerbose)
			if err != nil {
				return err, 1
			} else {
				return nil, 0
			}
		} else {
			dockerized.SetCommandVersion(composeFilePaths, commandName, optionVerbose, commandVersion)
		}
	}

	project, err := dockerized.GetProject(composeFilePaths)
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
		err := dockerized.DockerComposeBuild(composeFilePaths, api.BuildOptions{
			Services: []string{commandName},
		})

		if err != nil {
			return err, 1
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
			fmt.Printf("Service %s not found in compose file(s). Fallback to: %s.\n", commandName, image)
			fmt.Printf("  This command, if it exists, will not support version switching.\n")
			fmt.Printf("  See: https://github.com/jessfraz/dockerfiles\n")
		}
		return dockerized.DockerRun(image, runOptions, volumes)
	}

	return dockerized.DockerComposeRun(project, runOptions, volumes)
}

func parseArguments(args []string) ([]string, string, string, []string) {
	var options = []string{
		"--shell",
		"--build",
		"-h",
		"--help",
		"-v",
		"--verbose",
		"--version",
	}

	commandName := ""
	var commandArgs []string
	var dockerizedOptions []string
	var commandVersion string
	for _, arg := range args {
		if arg[0] == '-' && commandName == "" {
			if util.Contains(options, arg) {
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
