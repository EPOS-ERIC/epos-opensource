# EPOS Open Source CLI

A command-line tool for deploying the EPOS Platform locally using Docker Compose or Kubernetes.

## Features

- Deploy new environments
- Populate environments with metadata
- Update or delete an existing environment
- Manage multiple named deployments

## Requirements

- Docker and Docker Compose for Docker-based setups
- `kubectl` and access to a Kubernetes cluster for Kubernetes deployments

## Installation

### Pre-built binaries

1. Download the archive for your platform from the [releases](https://github.com/epos-eu/epos-opensource/releases).
2. Make the binary executable and place it in your `$PATH`:
   ```shell
   chmod +x epos-opensource
   mv epos-opensource /usr/local/bin/
   ```
3. Check the installation:
   ```shell
   epos-opensource --version
   ```

### Build from source

- Make sure to have the Go 1.24.4 (or more recent) toolchain when building from source

```shell
make build
```

## Usage

List all commands with:

```shell
epos-opensource --help
```

### Docker example

```shell
epos-opensource docker deploy myenv
```

### Kubernetes example

```shell
epos-opensource kubernetes deploy myenv
```

## Maintenance

Container images referenced by the tool are updated regularly.

## Contributing

Look for a project's contribution instructions. If there are any, follow them.

- Create a personal fork of the project on Github/GitLab.
- Clone the fork on your local machine. Your remote repo on Github/GitLab is called `origin`.
- Add the original repository as a remote called `upstream`.
- If you created your fork a while ago be sure to pull upstream changes into your local repository.
- Create a new branch to work on! Branch from `develop` if it exists, else from `master` or `main`.
- Implement/fix your feature, comment your code.
- Follow the code style of the project, including indentation.
- If the project has tests run them!
- Write or adapt tests as needed.
- Add or change the documentation as needed.
- Squash your commits into a single commit with git's [interactive rebase](https://help.github.com/articles/interactive-rebase). Create a new branch if necessary.
- Push your branch to your fork on Github/GitLab, the remote `origin`.
- From your fork open a pull request in the correct branch. Target the project's `develop` branch if there is one, else go for `master` or `main`!
- ...
- If the maintainer requests further changes just push them to your branch. The PR will be updated automatically.
- Once the pull request is approved and merged you can pull the changes from `upstream` to your local repo and delete
  your extra branch(es).

And last but not least: Always write your commit messages in the present tense. Your commit message should describe what the commit, when applied, does to the code, not what you did to the code.
