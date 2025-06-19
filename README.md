# EPOS Open Source

## Introduction

epos-opensource is part of the EPOS Open Source project for local installation using either Docker or Kubernetes.
It consists of a simple cli tool to deploy instances of the EPOS Platform.

Use the `epos-opensource` binary to spin up local environment on Linux, Mac OS X or Windows.

## Prerequisites

- For Kubernetes installation of the environment `kubectl` must be installed and accessible on path. A context must be set where the namespace with the EPOS Platform will be deployed. For further information follow the [official guidelines](https://kubernetes.io/docs/home/)
- For Docker environments, Docker must be installed and `docker compose` must be accessible on path. For further information follow the [official guidelines](https://docs.docker.com/get-docker/)

## Installation

1. **Download the Binary:**  
    Grab the binary for your platform from the [releases](https://github.com/epos-eu/epos-opensource/releases) section.
   Give permissions on `` file and move on binary folder from a Terminal (in Linux/MacOS):
2. **Make it executable:**
   ```shell
   chmod +x epos-opensource
   ```
3. **Test it:**
   ```shell
   ./epos-opensource --version
   ```

## Usage

For a complete list of available commands and their options:

```sh
epos-opensource --help
```

For help with a specific command:

```sh
epos-opensource [command] --help
```

## Maintenance

We regularly update images used in this stack.

## Contributing

If you want to contribute to a project and make it better, your help is very welcome. Contributing is also a great way to learn more about social coding on Github, new technologies and and their ecosystems and how to make constructive, helpful bug reports, feature requests and the noblest of all contributions: a good, clean pull request.

### How to make a clean pull request

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
- …
- If the maintainer requests further changes just push them to your branch. The PR will be updated automatically.
- Once the pull request is approved and merged you can pull the changes from `upstream` to your local repo and delete
  your extra branch(es).

And last but not least: Always write your commit messages in the present tense. Your commit message should describe what the commit, when applied, does to the code – not what you did to the code.

```

```
