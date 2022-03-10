# Dockerized doctl

```shell
dockerized doctl
```

## Examples

```shell
dockerized doctl auth init      # Set authentication token
dockerized doctl projects list  # List projects
```

## Volumes

- `~/.dockerized/apps/doctl` &rarr; `/root`
  - This directory will contain `.config/config.yaml`, containing your authentication token.
  - You can access the directory from your host machine.
