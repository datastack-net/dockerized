# Adding commands

To add your favorite command, simply add an entry to the [docker-compose.yml](docker-compose.yml) file.

For example, this is how `go` is added:

```yaml
services:
  go:
    image: golang:latest
    entrypoint: [ "go" ]
```

- `image: golang:latest` specifies the docker image to use. You can find these on [Docker Hub](https://hub.docker.com/).
- `entrypoint` is the command to run when the service starts.

That's it. Now you can run `dockerized go`.

## Configurable version

Let's make sure users can choose the version of the command.

Check the current version of the command. Often with `<command> --version` or `<command> version`.

```shell
dockerized go version
> go version go1.17.8 linux/amd64
```

Replace the version tag `latest` with `${GO_VERSION}`:

```yaml
  go:
    image: "golang:${GO_VERSION}"
    entrypoint: [ "go" ]
```

Set the global default version in [.env](.env):

```dotenv
GO_VERSION=1.17.8
```

- Don't use `latest`, as there's no guarantee that newer versions will always work.

## Contribute your changes

```bash
# Create a branch:
dockerized git checkout -b fork

# Commit your changes:
dockerized git commit -am "Add go"

# Authenticate to github:
dockerized gh auth login
```

- Choose SSH, and Browser authentication
- See [gh](apps/gh/README.md) for more information.

```bash
# Create a PR:
dockerized gh pr create
```

## Dockerized Development

When running from an IDE like Goland, or using the `go run` command, dockerized will have trouble finding the included `.env` file in the root of the project.
To fix this, you can set up an environment variable `DOCKERIZED_ROOT` that points to the root of the project.

> 🤔 If anyone knows how to 'embed' static files in a go project

```bash
export DOCKERIZED_ROOT=/path/to/dockerized
go run main.go --help
```