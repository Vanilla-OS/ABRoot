<div align="center">
  <img src="abroot-logo.svg" height="120">
  <h1 align="center">ABRoot</h1>
  <p align="center">Provides full immutability and atomicity by transacting between 2 root partitions (A&lt;->B), it also allows on-demand transactions via a transactional shell.</p>
</div>

> **Note**: This is a work in progress. It is not ready for production use.

The intention of this project is to replace Almost in the first RC of Vanilla OS.

### Read here
This program is meant to be used with PackageKit and [apx](https://github.com/vanilla-os/apx), 
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
	enter			set the filesystem as ro or rw until reboot
	config			show the current configuration
	check			check whether the filesystem is read-only or read-write
```
