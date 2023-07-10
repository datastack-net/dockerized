package help

import (
	"fmt"
	dockerized "github.com/datastack-net/dockerized/pkg"
	"sort"
)

func Help(composeFilePaths []string) error {
	project, err := dockerized.GetProject(composeFilePaths)
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
	fmt.Println("      --build       Rebuild the container before running it.")
	fmt.Println("      --pull        Pull the latest version of the container before building it (with --build).")
	fmt.Println("      --no-cache    Do not use cache when building the container (with --build).")
	fmt.Println("      --shell       Start a shell inside the command container. Similar to `docker run --entrypoint=sh`.")
	fmt.Println("      --entrypoint <entrypoint>")
	fmt.Println("                    Override the default entrypoint of the command container.")
	fmt.Println("  -p <port>         Exposes given port to host, e.g. -p 8080")
	fmt.Println("  -p <port>:<port>  Maps host port to container port, e.g. -p 80:8080")
	fmt.Println("  -v, --verbose     Log what dockerized is doing.")
	fmt.Println("  -h, --help        Show this help.")
	fmt.Println()

	fmt.Println("Version:")
	fmt.Println("  :<version>        The version of the command to run, e.g. 1, 1.8, 1.8.1.")
	fmt.Println("  :?                List all available versions. E.g. `dockerized go:?`")
	fmt.Println("  :                 Same as ':?' .")
	fmt.Println()

	fmt.Println("Arguments:")
	fmt.Println("  All arguments after <command> are passed to the command itself.")

	return nil
}
