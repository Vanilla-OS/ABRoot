<div align="center">
  <img src="abroot-logo.svg" height="120">
  <h1 align="center">ABRoot</h1>
  <p align="center">ABRoot is a utility that allow fully atomic transactions between 2 root partitions (A⟺B).</p>
</div>

> **Note**: This is a work in progress. It is not ready for production use.

This program is meant to be used with [apx](https://github.com/vanilla-os/apx), 
an apt replacement for VanillaOS.

### Help

```bash
abroot [options] [command]

Options:
	--help/-h		show this message
	--verbose/-v		show more verbosity
	--version/-V		show version

Commands:
	get			outputs the present or future root partition state
	shell			enter a transactional shell in the future root partition and switch root on the next boot
	exec			execute a command in a transactional shell in the future root partition and switch to it on the next boot
```

## Docs

The official **documentation and manpage** for `abroot` are available at https://documentation.vanillaos.org/docs/ABRoot/.
