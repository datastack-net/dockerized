package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

var dockerizedCommands = []string{
	//"--shell",
}

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

func main() {
	args := os.Args[1:]
	for _, arg := range args {
		fmt.Println("- " + arg)
	}

	command := ""
	var commandArgs []string
	var dockerizedOptions []string
	for _, arg := range args {
		if arg[0] == '-' && command == "" {
			dockerizedOptions = append(dockerizedOptions, arg)
		} else {
			if command == "" {
				command = arg
			} else {
				commandArgs = append(commandArgs, arg)
			}
		}
	}

	if command == "" {
		fmt.Println("No command found")
		os.Exit(1)
	}

	// if main argument is a command
	if contains(dockerizedCommands, command) {
		switch command {
		case "--shell":

		}
	} else {
		// get current working directory
		cwd, err := os.Getwd()
		if err != nil {
			fmt.Println(err)
			panic(err)
		}

		dockerizedRoot := getDockerizedRoot()
		dockerizedDockerComposeFilePath := filepath.Join(dockerizedRoot, "docker-compose.yml")

		composeRunArgs := []string{
			"-f", dockerizedDockerComposeFilePath,
			"run", "--rm",
			"-v", cwd + ":" + "/host",
			"-w", "/host",
		}
		if contains(dockerizedOptions, "--shell") {
			composeRunArgs = append(composeRunArgs, "--entrypoint=sh")
			commandArgs = []string{
				"-c",
				"$(which bash sh zsh | head -n 1)",
			}
		} else if contains(dockerizedOptions, "--build") {
			dockerCompose([]string{
				"-f", dockerizedDockerComposeFilePath,
				"build",
				command,
			})
		}
		composeRunArgs = append(composeRunArgs, command)
		composeRunArgs = append(composeRunArgs, commandArgs...)

		fmt.Printf("composeRunArgs: %v", composeRunArgs)

		dockerCompose(composeRunArgs)
	}
}

func dockerCompose(composeRunArgs []string) {
	cmd := exec.Command("docker-compose", composeRunArgs...)
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "FOOBAR=some_value")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}

func versions(serviceName string) {

}
