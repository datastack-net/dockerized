# Dockerized s3cmd

```shell
dockerized s3cmd
```


## Setup

- Add your aws-credentials to a `dockerized.env` file, in the root of your project or in your HOME directory.
  ```ini
  AWS_ACCESS_KEY_ID=****
  AWS_SECRET_ACCESS_KEY=****
  ```
- Configure s3cmd
  ```shell
  dockerized s3cmd --configure
  ```
- Configuration will be saved to `~/.dockerized/apps/s3cmd/.s3cfg`

## Usage

```shell
dockerized s3cmd --help
dockerized s3cmd ls
```
