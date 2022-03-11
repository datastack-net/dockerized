# Adding commands

To add your favorite command, simply add an entry to the [docker-compose.yml](docker-compose.yml) file.

For example, this is how `go` is added:

```yaml
services:
  go:
    <<: *default
    image: golang:latest
    entrypoint: [ "go" ]
```

- `<<: *default` is a YAML reference that copies some default settings for the service. Just include it.
- `image: golang:latest` specifies the docker image to use. You can find these on [Docker Hub](https://hub.docker.com/).
- `entrypoint` is the command to run when the service starts.

That's it. Now you can run `dockerize go`.

## Configurable version

Let's make sure users can choose the version of the command.

Check the current version of the command. Often with `<command> --version` or `<command> version`.

```shell
dockerize go version
> go version go1.17.8 linux/amd64
```

Replace the version tag `latest` with `${GO_VERSION}`:

```yaml
  go:
    <<: *default
    image: "golang:${GO_VERSION}"
    entrypoint: [ "go" ]
```

Set the global default version in [dockerized.env](dockerized.env):

```dotenv
GO_VERSION=1.17.8
```

- Don't use `latest`, as there's no guarantee that newer versions will always work.

## Contribute your changes

```bash
# Create a branch:
dockerize git checkout -b fork

# Commit your changes:
dockerize git commit -am "Add go"

# Authenticate to github:
dockerize gh auth login
```

- Choose SSH, and Browser authentication
- See [gh](apps/gh/Readme.md) for more information.

```bash
# Create a PR:
dockerize gh pr create
```
