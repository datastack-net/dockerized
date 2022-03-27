package main

import (
	"fmt"
	dockerized "github.com/datastack-net/dockerized/pkg"
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
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

func TestOverrideVersionWithEnvVar(t *testing.T) {
	defer context().WithEnv("PROTOC_VERSION", "3.6.0").Restore()
	var output = testDockerized(t, []string{"protoc", "--version"})
	assert.Contains(t, output, "libprotoc 3.6.0")
}

func TestCustomGlobalComposeFileAdditionalService(t *testing.T) {
	homePath := dockerized.GetDockerizedRoot() + "/test/additional_service"

	println(strings.Join(os.Environ(), " / "))

	defer context().
		WithHome(homePath).
		WithHomeEnvFile(`COMPOSE_FILE="${COMPOSE_FILE};${HOME}/docker-compose.yml"`).
		WithFile(homePath+"/docker-compose.yml", `
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
	homePath := dockerized.GetDockerizedRoot() + "/test/customized_service"

	defer context().
		WithHome(homePath).
		WithHomeEnvFile(`COMPOSE_FILE="${COMPOSE_FILE};${HOME}/docker-compose.yml"`).
		WithFile(homePath+"/docker-compose.yml", `
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

	defer context().
		WithCwd(projectPath).
		WithHomeEnvFile(`COMPOSE_FILE="${COMPOSE_FILE};docker-compose.yml"`).
		WithFile(projectPath+"/docker-compose.yml", `
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

func (c *Context) WithEnv(key string, value string) *Context {
	_ = os.Setenv(key, value)
	//c.after = append(c.after, func() {
	//	_ = os.Unsetenv(key)
	//})
	return c
}

func (c *Context) WithHome(path string) *Context {
	c.homePath = path
	_ = os.MkdirAll(path, os.ModePerm)
	c.WithEnv("HOME", path)
	c.WithEnv("USERPROFILE", path)
	return c
}

func (c *Context) WithCwd(path string) *Context {
	c.cwdBefore, _ = os.Getwd()
	os.Chdir(path)
	c.after = append(c.after, func() {
		os.Chdir(c.cwdBefore)
	})
	return c
}

func (c *Context) WithHomeEnvFile(content string) *Context {
	return c.WithFile(c.homePath+"/dockerized.env", content)
}

func (c *Context) WithFile(path string, content string) *Context {
	_ = os.WriteFile(path, []byte(content), 0644)
	c.after = append(c.after, func() {
		_ = os.Remove(path)
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
	assert.Nil(t, err, fmt.Sprintf("error: %s", err))
	assert.Equal(t, 0, exitCode)
	return output
}
