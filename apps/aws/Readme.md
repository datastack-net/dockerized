# dockerized aws (aws-cli)

```shell
dockerized aws
```

## Setup

- Add your aws-credentials to a `dockerized.env` file, in the root of your project or in your HOME directory.
  ```ini
  AWS_ACCESS_KEY_ID=****
  AWS_SECRET_ACCESS_KEY=****
  ```

## Usage

```shell
dockerized aws s3 ls
dockerized aws s3 --endpoint-url=https://ams3.digitaloceanspaces.com ls
```
