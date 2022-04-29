# Dockerized mkdocs

```shell
dockerized mkdocs
```

## Getting started

> See: https://www.mkdocs.org/getting-started/

**Installation**

You can skip this step.

**Creating a new Project**

```shell
dockerized mkdocs new my-project
```

**Serving**

The `mkdocs serve` command needs some adjustment to work with dockerized:

```shell
dockerized -p 8000 mkdocs serve --dev-addr=0.0.0.0:8000
```

- `-p 8000` &mdash; Tells dockerized to forward port 8000 from the host to the container.
- `--dev-addr=0.0.0.0:8000` &mdash; Tells mkdocs to listen on all interfaces on port 8000. 

You can now access the live site at http://localhost:8000.

## Mkdocs plugins

> See:
>  - https://www.mkdocs.org/dev-guide/plugins/ 
>  - https://github.com/mkdocs/mkdocs/wiki/MkDocs-Plugins

By default, the following plugins are installed:

- `mkdocs-material`
- `mkdocs-material-extensions`

To install more plugins:

- Add their pip package name to the `MKDOCS_PACKAGES` environment variable.
    
    ```shell
    # dockerized.env
    MKDOCS_PACKAGES="${MKDOCS_PACKAGES} mdx-truly-sane-lists"
    ```
   
    Rebuild mkdocs:

    ```shell
    dockerized --build mkdocs
    ```
- Follow the other instructions in the plugin's documentation.


