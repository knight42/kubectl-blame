kubectl-blame: git-like blame for kubectl
======

![GitHub](https://img.shields.io/github/license/knight42/kubectl-blame)
![](https://github.com/knight42/kubectl-blame/actions/workflows/build.yml/badge.svg)
![GitHub last commit](https://img.shields.io/github/last-commit/knight42/kubectl-blame)

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

## Usage

```
# Blame pod 'foo' in default namespace
kubectl blame pods foo

# Blame deployment 'foo' and 'bar' in 'ns1' namespace
kubectl blame -n ns1 deploy foo bar

# Blame deployment 'bar' in 'ns1' namespace and hide the update time
kubectl blame -n ns1 --time none deploy bar

# Blame resources in file 'pod.yaml'(will access remote server)
kubectl blame -f pod.yaml

# Blame deployment saved in local file 'deployment.yaml'(will NOT access remote server)
kubectl blame -i deployment.yaml
# Or
cat deployment.yaml | kubectl blame -i -
```

### Flags

| flag               | default    | description                                                              |
|--------------------|------------|--------------------------------------------------------------------------|
| `--time`           | `relative` | Time format. One of: `full`, `relative`, `none`.                         |
| `--filename`, `-f` |            | Filename identifying the resource to get from a server.                  |
| `--input`, `-i`    |            | Read object from the give file. When the file is -, read standard input. |
