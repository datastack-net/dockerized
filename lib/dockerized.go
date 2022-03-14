package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
)

var options = []string{
	"--shell",
	"--build",
	"-h",
	"--help",
	"-v",
	"--verbose",
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

func main() {
	command := ""
	var commandArgs []string
	var dockerizedOptions []string
	for _, arg := range os.Args[1:] {
		if arg[0] == '-' && command == "" {
			if contains(options, arg) {
				dockerizedOptions = append(dockerizedOptions, arg)
			} else {
				fmt.Println("Unknown option:", arg)
				os.Exit(1)
			}
		} else {
			if command == "" {
				command = arg
			} else {
				commandArgs = append(commandArgs, arg)
			}
		}
	}

	dockerizedRoot := getDockerizedRoot()
	dockerizedDockerComposeFilePath := filepath.Join(dockerizedRoot, "docker-compose.yml")

	var optionHelp = contains(dockerizedOptions, "--help") || contains(dockerizedOptions, "-h")
	var optionVerbose = contains(options, "--verbose") || contains(options, "-v")

	if command == "" || optionHelp {
		help(dockerizedDockerComposeFilePath)
		os.Exit(1)
	}

	hostCwd, _ := os.Getwd()
	hostCwdDirName := filepath.Base(hostCwd)
	composeRunArgs := []string{
		"-f", dockerizedDockerComposeFilePath,
		"run", "--rm",
		"-v", hostCwd + ":" + "/host/" + hostCwdDirName,
		"-w", "/host/" + hostCwdDirName,
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

	// fmt.Printf("composeRunArgs: %v\n", composeRunArgs)

	homeDir, _ := os.UserHomeDir()

	// find the first .env file above the current directory

	userGlobalDockerizedEnvFile := filepath.Join(homeDir, dockerizedEnvFileName)
	localDockerizedEnvFile, err := findLocalEnvFile(hostCwd)

	envFiles := []string{}

	if _, err := os.Stat(userGlobalDockerizedEnvFile); err == nil {
		envFiles = append(envFiles, userGlobalDockerizedEnvFile)
	}
	if err == nil && !contains(envFiles, localDockerizedEnvFile) {
		envFiles = append(envFiles, localDockerizedEnvFile)
		//if optionVerbose {
		//	fmt.Println("Loading dockerized.env from", localDockerizedEnvFile)
		//}
		//godotenv.Load(localDockerizedEnvFile)
	}

	if optionVerbose {
		// Print it in order of priority (lowest to highest)
		for _, envFile := range envFiles {
			fmt.Println("Loading: ", envFile)
		}
	}
	// Load in reverse. GoDotEnv does not override vars, this allows runtime env-vars to override the env files.
	for i := len(envFiles) - 1; i >= 0; i-- {
		godotenv.Load(envFiles[i])
	}
}

func help(dockerComposeFilePath string) {
	fmt.Println("Usage: dockerized [options] <command> [args]")
	fmt.Println("")
	fmt.Println("Commands:")
	//fmt.Println("  ")
	cmd := exec.Command("docker-compose", "-f", dockerComposeFilePath, "config", "--service")
	output, err := cmd.CombinedOutput()
	if err != nil {
		panic(err)
	}

	services := strings.Split(string(output), "\n")
	sort.Strings(services)

	for _, line := range services {
		if line != "" && line[0] != '_' {
			fmt.Printf("  %s\n", line)
		}
	}

	fmt.Println("")

	fmt.Println("Options:")
	fmt.Println("  --help, -h Show this help")
	fmt.Println("")
	fmt.Println("  --shell    Start a shell inside the command container. Similar to `docker run --entrypoint=sh`.")
	fmt.Println("  --build    Rebuild the container before running it.")
}

func dockerCompose(composeRunArgs []string) {
	cmd := exec.Command("docker-compose", composeRunArgs...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Run()
}
