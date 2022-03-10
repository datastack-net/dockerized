# Dockerized
Run specific versions of popular commandline tools within docker, so you don't have to install them.

## Installation

- Clone this repo anywhere: `git clone git@github.com:datastack-net/dockerized.git`
- Add the `bin` directory to your path

## Usage

Run any supported command, but within Docker.

```bash
dockerized node --version             # v16.13.0
dockerized vue create new-project     # create a project with vue cli
dockerized tsc --init                 # initialize typescript for the current directory
dockerized npm install                # install packages.json
```

## Use Cases

- Quickly try out command line tools without the effort of downloading and installing them.
- Installing multiple versions of node/python/typescript.
- You need unix commands on Windows.
- You don't want to pollute your system with a lot of tools you don't use.
- Easily update your tools.
- Ensure everyone on the team is using the same version of commandline tools.

## Design Goals

- All commands work out of the box.
- Dockerized commands behave the same as their native counterparts.
  - Files in the current directory are accessible using relative paths.
- Cross-platform: Works on Linux, MacOS, and Windows (CMD, Powershell, Git Bash).
- Suitable for ad-hoc usage (i.e. you quickly need to run a command, that is not on your system).
- Configurability: for use within a project or CI/CD pipeline.

## Specify command version

Each command has a `<COMMAND>_VERSION` environment variable which you can override.

**Ad-hoc (Unix)**
```bash
NODE_VERSION=15.0.0 dockerized node --version
15.0.0: Pulling from library/node
0400ac8f7460: Downloading [=============================================>     ]  40.93MB/45.37MB
# ...
v15.0.0
```

**Per directory**

You can specify the versions per directory. This allows you to "lock" your tools to specific versions for your project.

Create a `dockerized.env` file in your project root, containing your versions:

```env
NODE_VERSION=15.0.0
```

From anywhere within the project directory, you will get node 15 when calling `dockerized node`.

**Windows**

- Create a `dockerized.env` file in your project for local configuration.

**Global**

- Create a `dockerized.env` file in your home directory for global configuration.

## Supported commands

- [dotnet](apps/dotnet/Readme.md)
- http
- jq
- node
- npm
- npx
- protoc
- pip
- python
- python2
- s3cmd
- tsc
- tree
- swagger-codegen
- vue
- wget
- yarn
