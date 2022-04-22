package main

import (
	"fmt"
	dockerized "github.com/datastack-net/dockerized/pkg"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"k8s.io/apimachinery/pkg/util/rand"
	"os"
	"strconv"
	"strings"
	"testing"
)

var initialEnv = os.Environ()

type Context struct {
	homePath  string
	before    []func()
	after     []func()
	envBefore []string
	cwdBefore string
}

func TestHelp(t *testing.T) {
	output := testDockerized(t, []string{"--help"})
	assert.Contains(t, output, "Usage:")
}

func TestEntrypoint(t *testing.T) {
	var projectDir = dockerized.GetDockerizedRoot() + "/test/test_entrypoint"
	defer context().
		WithDir(projectDir).
		WithCwd(projectDir).
		WithFile("foo.txt", "foo").
		WithFile("bar.txt", "bar").
		Restore()

	output := testDockerized(t, []string{"--entrypoint", "ls", "go"})
	assert.Contains(t, output, "foo.txt")
	assert.Contains(t, output, "bar.txt")
}

func TestOverrideVersionWithEnvVar(t *testing.T) {
	defer context().WithEnv("PROTOC_VERSION", "3.6.0").Restore()
	var output = testDockerized(t, []string{"protoc", "--version"})
	assert.Contains(t, output, "libprotoc 3.6.0")
}

func TestLocalEnvFileOverridesGlobalEnvFile(t *testing.T) {
	var projectPath = dockerized.GetDockerizedRoot() + "/test/project_override_global"
	defer context().
		WithTempHome().
		WithHomeEnvFile("PROTOC_VERSION=3.6.0").
		WithDir(projectPath).
		WithCwd(projectPath).
		WithFile(projectPath+"/dockerized.env", "PROTOC_VERSION=3.8.0").
		Restore()
	var output = testDockerized(t, []string{"-v", "protoc", "--version"})
	assert.Contains(t, output, "libprotoc 3.8.0")
}

func TestRuntimeEnvOverridesLocalEnvFile(t *testing.T) {
	var projectPath = dockerized.GetDockerizedRoot() + "/test/project_override_global"
	defer context().
		WithTempHome().
		WithDir(projectPath).
		WithCwd(projectPath).
		WithFile(projectPath+"/dockerized.env", "PROTOC_VERSION=3.8.0").
		WithEnv("PROTOC_VERSION", "3.16.1").
		Restore()
	var output = testDockerized(t, []string{"protoc", "--version"})
	assert.Contains(t, output, "libprotoc 3.16.1")
}

func TestCustomGlobalComposeFileAdditionalService(t *testing.T) {
	defer context().
		WithTempHome().
		WithHomeEnvFile(`COMPOSE_FILE="${COMPOSE_FILE};${HOME}/docker-compose.yml"`).
		WithHomeFile("docker-compose.yml", `
version: "3"
services:
  test:
    image: alpine
`).
		Restore()
	var output = testDockerized(t, []string{"test", "uname"})
	assert.Contains(t, output, "Linux")
}

func TestUserCanGloballyCustomizeDockerizedCommands(t *testing.T) {
	defer context().
		WithTempHome().
		WithHomeEnvFile(`COMPOSE_FILE="${COMPOSE_FILE};${HOME}/docker-compose.yml"`).
		WithHomeFile("docker-compose.yml", `
version: "3"
services:
  alpine:
    environment:
      CUSTOM: "CUSTOM_123456"
`).
		Restore()
	var output = testDockerized(t, []string{"alpine", "env"})
	assert.Contains(t, output, "CUSTOM_123456")
}

func TestUserCanLocallyCustomizeDockerizedCommands(t *testing.T) {
	projectPath := dockerized.GetDockerizedRoot() + "/test/project_with_customized_service"
	projectSubPath := projectPath + "/sub"

	defer context().
		WithTempHome().
		WithDir(projectPath).
		WithDir(projectSubPath).
		WithCwd(projectSubPath).
		WithFile(projectPath+"/dockerized.env", `COMPOSE_FILE="${COMPOSE_FILE};${DOCKERIZED_PROJECT_ROOT}/docker-compose.yml"`).
		WithFile(projectPath+"/docker-compose.yml", `
version: "3"
services:
  alpine:
    environment:
      CUSTOM: "CUSTOM_123456"
`).
		Restore()
	var output = testDockerized(t, []string{"-v", "alpine", "env"})
	assert.Contains(t, output, "CUSTOM_123456")
}

func TestUserCanIncludeGlobalAndProjectComposeFile(t *testing.T) {
	projectPath := dockerized.GetDockerizedRoot() + "/test/project" + strconv.Itoa(rand.Int())

	context := context().
		WithTempHome().
		WithHomeEnvFile(`COMPOSE_FILE="${COMPOSE_FILE};${HOME}/docker-compose.dockerized.yml"`).
		WithHomeFile("docker-compose.dockerized.yml", `
version: "3"
services:
  home_cmd:
    image: "alpine"
    entrypoint: [ "echo", "HOME123" ]
`).
		WithDir(projectPath).
		WithCwd(projectPath).
		WithFile(projectPath+"/dockerized.env", `COMPOSE_FILE="${COMPOSE_FILE};${DOCKERIZED_PROJECT_ROOT}/docker-compose.dockerized.yml"`).
		WithFile(projectPath+"/docker-compose.dockerized.yml", `
version: "3"
services:
  project_cmd:
    image: "alpine"
    entrypoint: [ "echo", "PROJECT123" ]
`)
	defer context.Restore()
	var outputProjectCmd = testDockerized(t, []string{"-v", "project_cmd"})
	assert.Contains(t, outputProjectCmd, "PROJECT123")

	var outputHomeCmd = testDockerized(t, []string{"-v", "home_cmd"})
	assert.Contains(t, outputHomeCmd, "HOME123")

}

func (c *Context) WithEnv(key string, value string) *Context {
	_ = os.Setenv(key, value)
	return c
}

func (c *Context) WithHome(path string) *Context {
	c.homePath = path
	c.WithDir(path)
	c.WithEnv("HOME", path)
	c.WithEnv("USERPROFILE", path)
	return c
}

func (c *Context) WithTempHome() *Context {
	var homePath = dockerized.GetDockerizedRoot() + "/test/home" + strconv.Itoa(rand.Int())
	c.WithHome(homePath)
	return c
}

func (c *Context) WithCwd(path string) *Context {
	c.cwdBefore, _ = os.Getwd()
	err := os.Chdir(path)
	if err != nil {
		panic(err)
	}
	c.after = append(c.after, func() {
		os.Chdir(c.cwdBefore)
	})
	return c
}

func (c *Context) WithHomeEnvFile(content string) *Context {
	return c.WithHomeFile("dockerized.env", content)
}

func (c *Context) WithFile(path string, content string) *Context {
	_ = os.WriteFile(path, []byte(content), 0644)
	c.after = append(c.after, func() {
		_ = os.Remove(path)
	})
	return c
}

func (c *Context) WithHomeFile(path string, content string) *Context {
	return c.WithFile(c.homePath+"/"+path, content)
}

func (c *Context) WithDir(path string) *Context {
	_ = os.MkdirAll(path, os.ModePerm)
	c.after = append(c.after, func() {
		_ = os.RemoveAll(path)
	})
	return c
}

func (c *Context) Execute(callback func()) {
	for _, before := range c.before {
		before()
	}
	defer func() {
		for _, after := range c.after {
			after()
		}
	}()
	callback()
}

func restoreInitialEnv() {
	os.Clearenv()
	for _, envEntry := range initialEnv {
		keyValue := strings.Split(envEntry, "=")
		os.Setenv(keyValue[0], keyValue[1])
	}
}

func (c *Context) Restore() {
	defer restoreInitialEnv()
	for _, after := range c.after {
		//goland:noinspection GoDeferInLoop
		defer after()
	}
}

func context() *Context {
	restoreInitialEnv()
	return &Context{
		envBefore: os.Environ(),
	}
}

func TestOverrideVersionWithGlobalEnvFile(t *testing.T) {
	defer context().
		WithHome(dockerized.GetDockerizedRoot() + "/test/home").
		WithHomeEnvFile("PROTOC_VERSION=3.8.0").
		Restore()

	var output = testDockerized(t, []string{"protoc", "--version"})

	assert.Contains(t, output, "3.8.0")
}

func capture(callback func()) string {
	read, write, _ := os.Pipe()
	os.Stdout = write

	callback()

	os.Stdout.Close()
	bytes, _ := ioutil.ReadAll(read)
	var output = string(bytes)
	return output
}

func testDockerized(t *testing.T, args []string) string {
	var err error
	var exitCode int
	var output = capture(func() {
		err, exitCode = RunCli(args)
	})
	println(output)
	assert.Nil(t, err, fmt.Sprintf("error: %s", err))
	assert.Equal(t, 0, exitCode)
	return output
}
