# EPOS Open Source CLI

A command-line tool for deploying the EPOS Platform locally using Docker Compose or Kubernetes.

## Features

- Deploy new environments
- Populate environments with metadata
- Update or delete existing environments
- Manage multiple named deployments

## Requirements

- Docker and Docker Compose (for Docker-based setups)
- `kubectl` and access to a Kubernetes cluster (for Kubernetes-based setups)
- Go 1.16 or later (for installation via `go install`)
- Go 1.24.4 or later (for building from source)

## Installation

### Installation script

> [!IMPORTANT]
> ALWAYS REVIEW SCRIPTS BEFORE EXECUTING ON YOUR SYSTEM!

```bash
curl -fsSL https://raw.githubusercontent.com/epos-eu/epos-opensource/main/install.sh | bash
```

### Using Go

Install the latest version:

```shell
go install github.com/epos-eu/epos-opensource@latest
```

Ensure `$GOPATH/bin` (or `$HOME/go/bin`) is in your `$PATH`.

### Pre-built binaries

1. Download the appropriate archive from the [releases](https://github.com/epos-eu/epos-opensource/releases).
2. Make the binary executable and move it to your `$PATH`:

   ```shell
   chmod +x epos-opensource-{your version}
   mv epos-opensource-{your version} /usr/local/bin/epos-opensource
   ```

3. Verify the installation:

   ```shell
   epos-opensource --version
   ```

### Build from source

Build using `make` with Go 1.24.4 or later:

```shell
make build
```

## Usage

List available commands:

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

Referenced container images are updated regularly.

## Contributions Welcome

We welcome all contributions, including bug reports, feature ideas, documentation, or code changes.

If you have questions or are unsure how to get started, feel free to [open an issue](https://github.com/epos-eu/epos-opensource/issues). We are happy to assist!

## Contributing

1. Fork the repository and clone it locally.
2. Create a new branch from `main` or `develop`.
3. Make your changes and follow the existing code style.
4. Add or update tests and documentation as needed.
5. Commit your changes using clear, present-tense messages.
6. Push your branch and open a pull request.

Before submitting:

- Pull the latest changes from upstream.
- Squash commits if needed.
- Ensure all tests pass.

After submitting:

- Address any feedback on your pull request.
- Update your branch until it is approved and merged.

Thank you for your contribution!
