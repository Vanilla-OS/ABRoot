<div align="center">
  <img src="abroot-logo.svg" height="120">
  <h1 align="center">ABRoot</h1>
  
[![Build status][github-actions-image]][github-actions-url]
[![Translation Status][weblate-image]][weblate-url]
[![build result][build-image]][build-url]
  
[github-actions-url]: https://github.com/Vanilla-OS/ABRoot/actions/workflows/go.yml
[github-actions-image]: https://github.com/Vanilla-OS/ABRoot/actions/workflows/go.yml/badge.svg
[weblate-url]: https://hosted.weblate.org/engage/vanilla-os
[weblate-image]: https://hosted.weblate.org/widgets/vanilla-os/-/abroot/svg-badge.svg
[weblate-status-image]: https://hosted.weblate.org/widgets/vanilla-os/-/abroot/multi-auto.svg
[build-image]: https://build.opensuse.org/projects/home:fabricators:orchid/packages/abroot/badge.svg?type=default
[build-url]: https://build.opensuse.org/package/show/home:fabricators:orchid/abroot

  <p align="center">ABRoot is a utility that allow fully atomic transactions between 2 root partitions (A⟺B).</p>

<i>This program is meant to be used with [apx](https://github.com/vanilla-os/apx), 
an apt replacement for VanillaOS.</i>
</div>

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

## Translations

Contribute translations for the manpage and help page in [Weblate](https://hosted.weblate.org/projects/vanilla-os/abroot).

[![Translation Status][weblate-status-image]][weblate-url]

### Generating man pages for translations

Once the translation is complete in Weblate and the changes committed, clone the repository using `git` and perform `go build`, create a directory using the `mkdir man/<language_code>` command, and execute this command `LANG=<language_code> ./abroot man > man/<language_code>/abroot.1`. Open a PR for the generated manpage here.
