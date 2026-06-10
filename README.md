[![CI](https://github.com/heathcliff26/gh-utility/actions/workflows/ci.yaml/badge.svg?event=push)](https://github.com/heathcliff26/gh-utility/actions/workflows/ci.yaml)
[![Coverage Status](https://coveralls.io/repos/github/heathcliff26/gh-utility/badge.svg)](https://coveralls.io/github/heathcliff26/gh-utility)
[![Editorconfig Check](https://github.com/heathcliff26/gh-utility/actions/workflows/editorconfig-check.yaml/badge.svg?event=push)](https://github.com/heathcliff26/gh-utility/actions/workflows/editorconfig-check.yaml)
[![Coverage report](https://github.com/heathcliff26/gh-utility/actions/workflows/go-testcover-report.yaml/badge.svg)](https://github.com/heathcliff26/gh-utility/actions/workflows/go-testcover-report.yaml)

# gh-utility

CLI tool to interact with the GitHub API as an app.

## Table of Contents

- [gh-utility](#gh-utility)
  - [Table of Contents](#table-of-contents)
  - [Usage](#usage)
    - [Tekton tasks](#tekton-tasks)
  - [Container Images](#container-images)
    - [Image location](#image-location)
    - [Tags](#tags)

## Usage

Example usage with container:
```bash
podman run -it ghcr.io/heathcliff26/gh-utility:latest --help
```
Output of help:
```bash
$ gh-utility help
gh-utility to interact with the GitHub API as an app

Usage:
  gh-utility [flags]
  gh-utility [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  token       Create a new temporary access token for the app
  version     Print version information and exit

Flags:
  -h, --help   help for gh-utility

Use "gh-utility [command] --help" for more information about a command.
```

### Tekton tasks

Examples for using this image inside tekton tasks can be found in the [ci repository](https://github.com/heathcliff26/ci/tree/main/tekton/pipelines/shared-tasks).

## Container Images

### Image location

| Container Registry                                                                      | Image                       |
| --------------------------------------------------------------------------------------- | --------------------------- |
| [Github Container](https://github.com/users/heathcliff26/packages/container/package/gh-utility) | `ghcr.io/heathcliff26/gh-utility`   |
| [Docker Hub](https://hub.docker.com/r/heathcliff26/gh-utility)                                  | `docker.io/heathcliff26/gh-utility` |
| [Quay.io](https://quay.io/heathcliff26/gh-utility)                                              | `quay.io/heathcliff26/gh-utility`   |

### Tags

There are different flavors of the image:

| Tag(s)      | Description                                                 |
| ----------- | ----------------------------------------------------------- |
| **latest**  | Last released version of the image                          |
| **rolling** | Rolling update of the image, always build from main branch. |
| **vX.Y.Z**  | Released version of the image                               |
