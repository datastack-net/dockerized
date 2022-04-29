# Dockerized CLI Reference

```shell
dockerized [options] <command>[:version] [arguments]
```

When running dockerized from source, extra compilation options are available. See [Compilation options](#compilation-options).

## Options

- `--build` &mdash; Rebuild the container before running it.
- `--shell` &mdash; Start a shell inside the command container. Similar to `docker run --entrypoint=sh`.
- `--entrypoint <entrypoint>`   Override the default entrypoint of the command container.
- `-p <port>` &mdash; Exposes given port to host, e.g. `-p 8080`.
- `-p <port>:<port>` &mdash; Maps host port to container port, e.g. `-p 80:8080`.
- `-v`, `--verbose` &mdash; Log what dockerized is doing.
- `-h`, `--help` &mdash; Show this help.

## Version

- `:<version>` &mdash; The version of the command to run, e.g. `1`, `1.8`, `1.8.1`.
- `:?`, `:` &mdash; List all available versions. E.g. `dockerized go:?`

## Arguments

- All arguments after `<command>` are passed to the command itself.

## Compilation options

When running dockerized from source, there's an extra compilation option available.
It should be the first argument.

```shell
dockerized --compile[=docker|host]
```

- `--compile`, `--compile=docker` &mdash; Compile dockerized using docker.
- `--compile=host` &mdash; Compile dockerized on the host, requires `go` 1.17+ to be installed.

Example:

```shell
# Re-compile dockerized with go and run python.
dockerized --compile=host python
```