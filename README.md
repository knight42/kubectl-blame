kubectl-blame: git-like blame for kubectl
======

Annotate each line in the given resource's YAML with information from the managedFields
to show who last modified the field.

As long as the field `.metadata.manageFields` of the resource is set properly, this command
is able to display the manager of each field.

[![asciicast](https://asciinema.org/a/375008.svg)](https://asciinema.org/a/375008)

## Installing

| Distribution                           | Command / Link                                                         |
|----------------------------------------|------------------------------------------------------------------------|
| [Krew](https://krew.sigs.k8s.io/)      | `kubectl krew install blame`                                           |
| Pre-built binaries for macOS, Linux    | [GitHub releases](https://github.com/knight42/kubectl-blame/releases)  |

## Demos

### 1. Customize Time Format

[![asciicast](https://asciinema.org/a/375691.svg)](https://asciinema.org/a/375691)
