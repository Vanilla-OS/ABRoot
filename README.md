<div align="center">
  <img src="abroot-logo.svg" height="120">
  <h1 align="center">ABRoot v2</h1>
  <p align="center">ABRoot is utility which provides full immutability and
		atomicity to a Linux system, by transacting between
		two root filesystems. Updates are performed using OCI
		images, to ensure that the system is always in a
		consistent state.</p>
</div>

> **NOTE**: ABRoot v2 is currently in development. The current
> version is v1, which is available in the `v1` branch.

## Help

```md
ABRoot provides full immutability and atomicity by performing transactions between 2 root partitions (A<->B).

Usage:
  abroot [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  kargs       Manage kernel parameters
  pkg         Manage packages
  rollback    Return the system to a previous state
  status      Display status
  upgrade     Update the boot partition

Flags:
  -h, --help      help for abroot
  -v, --verbose   Show more detailed output
      --version   version for abroot

Use "abroot [command] --help" for more information about a command.
```
