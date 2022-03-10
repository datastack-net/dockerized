# Dockerized npm

```shell
dockerized npm
```

## Examples

```shell
dockerized npm init    # Initialize a new project
dockerized npm install # Install package.json
```

## Global installs

Global installations are supported and are stored within a docker volume, which is shared with `npx`, `node` and `npm`.

```shell
dockerized npx --package=@vue/cli vue
```
