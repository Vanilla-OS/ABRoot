<div align="center">
  <img src="abroot-logo.svg" height="120">
  <h1 align="center">ABRoot</h1>
  <p align="center">ABRoot is a utility that allow fully atomic transactions between 2 root partitions (A⟺B).</p>
</div>

> **Note**: This is a work in progress. It is not ready for production use.

This program is meant to be used with [apx](https://github.com/vanilla-os/apx), 
an apt replacement for VanillaOS.

## Help

```bash
ABRoot provides full immutability and atomicity by performing transactions between 2 root partitions (A<->B)

Usage:
  abroot [command]

Available Commands:
  completion   Generate the autocompletion script for the specified shell
  diff         Show modifications from latest transaction.
  exec         Execute a command in a transactional shell in the future root and switch to it on next boot
  get          Outputs the present or future root partition state (A or B)
  help         Help about any command
  kargs        Manage kernel parameters.
  rollback     Return the system to a previous state.
  shell        Enter a transactional shell

Flags:
  -h, --help      help for abroot
  -v, --verbose   show more detailed output
      --version   version for abroot

Use "abroot [command] --help" for more information about a command.
```

## Documentation

The official **documentation and manpage** for `abroot` are available at <https://documentation.vanillaos.org/docs/ABRoot>.

## Porting ABRoot to your Distribution

Learn how to port ABRoot to your distribution at <https://documentation.vanillaos.org/docs/ABRoot/porting>.

## Generating man pages for translations

- Copy the `en.yml` file under the `locales` directory, rename it to your language code then translate the strings.
- Once the translation is complete, perform `go build` and execute this command `LANG=<language_code> ./abroot man > man/<language_code>/abroot.1`. If the man page gets generated without errors, open a PR here.
