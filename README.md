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


- Create a `dockerized.env` file in your project for local configuration

**Global**

Coming soon.

## Supported commands

- http
- jq
- node
- npm
- npx
- protoc
- tsc
- s3cmd
- swagger-codegen
- vue
- wget
- yarn
