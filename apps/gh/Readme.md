# Dockerized gh (github client)

```shell
dockerized gh
```

## Authentication

Dockerized gh supports SSH out of the box, using its own SSH key.

```shell
dockerized gh auth login
```
- Protocol: SSH
- SSH key: `/root/.ssh/dockerized.pub`  
  > This key is generated for you, you can view it at `~/.dockerized/apps/gh/.ssh/dockerized.pub`
- Authenticate using Web Browser
  > As `gh` cannot access your host machine's browser, please open the urls by (Ctrl-) clicking them or copy-pasting them into your browser.
