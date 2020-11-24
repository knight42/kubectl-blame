kubectl-blame: git-like blame for kubectl
======

Annotate each line in the given resource's YAML with information from the managedFields
to show who last modified the field.

As long as the field `.metadata.manageFields` of the resource is set properly, this command
is able to display the manager of each field.

[![asciicast](https://asciinema.org/a/375008.svg)](https://asciinema.org/a/375008)

## Installing

```bash
```
