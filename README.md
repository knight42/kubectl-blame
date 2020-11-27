kubectl-blame: git-like blame for kubectl
======

Annotate each line in the given resource's YAML with information from the managedFields
to show who last modified the field.

As long as the field `.metadata.manageFields` of the resource is set properly, this command
is able to display the manager of each field.

[![asciicast](https://asciinema.org/a/375008.svg)](https://asciinema.org/a/375008)

## Installing

```bash
VERSION=0.0.2
curl -o kubectl-blame.tar.gz -Lf https://github.com/knight42/kubectl-blame/releases/download/v${VERSION}/kubectl-blame-v${VERSION}-$(go env GOOS)-amd64.tar.gz
tar xf kubectl-blame.tar.gz
cp kubectl-blame-v${VERSION}-$(go env GOOS)-amd64/kubectl-blame $GOPATH/bin/
```

## Demos

### 1. Customize Time Format

[![asciicast](https://asciinema.org/a/375691.svg)](https://asciinema.org/a/375691)
