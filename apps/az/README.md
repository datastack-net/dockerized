# Dockerized az (Azure CLI)

```shell
dockerized az
```

## Examples

```shell
dockerized az login          # Authenticate with Azure
dockerized az resource list  # List resources
```

## Volumes

- `~/.dockerized/apps/az` &rarr; `/root/.azure`
    - This directory will contain your azure profile.
    - You can access the directory from your host machine.
