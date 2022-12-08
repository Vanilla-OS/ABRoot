<div align="center">
  <img src="abroot-logo.svg" height="120">
  <h1 align="center">ABRoot</h1>
  <p align="center">Provides full immutability and atomicity by transacting between 2 root partitions (A&lt;->B), it also allows on-demand transactions via a transactional shell.</p>
</div>

> **Note**: This is a work in progress. It is not ready for production use.

The intention of this project is to replace Almost in the first RC of Vanilla OS.

### Read here
This program is meant to be used with [apx](https://github.com/vanilla-os/apx), 
an apt replacement for VanillaOS.

### Help
```
Usage: 
	abroot [options] [command]

Options:
	--help/-h		show this message
	--verbose/-v		show more verbosity
	--version/-V		show version

Commands:
	_update-boot		update the boot partition (only for advanced users)
	get			get the present or future root partition
	shell			enter a transactional shell in the future root partition and switch root on the next boot
	exec			execute a command in a transactional shell in the future root partition and switch to it on the next boot
```
