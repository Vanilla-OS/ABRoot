# Setup Notes

## Root Structure

- /
- /abimage.abr
- /bin -> usr/bin
- /boot
- /dev
- /etc
- /FsGuard
- /home -> var/home
- /lib -> usr/lib
- /lib64 -> usr/lib64
- /media -> run/media
- /mnt -> var/mnt
- /opt
- /part-future
- /proc
- /root -> var/root
- /run
- /sbin -> usr/sbin
- /srv -> var/srv
- /sys
- /sysconf
- /tmp
- /usr
- /usr/local -> var/usrlocal
- /var

## abimage.abr Example

```json
{
    "digest":"sha256:e9f2773ab60fabf4e456e4549d86e345b38754f1f8397183ce4dc28d52bab66e",
    "timestamp":"2023-04-23T15:13:19.066903531+02:00",
    "image":"registry.vanillaos.org/vanillaos/desktop:main"
}
```

## Boot Structure

- /
- /grub
- /grub/grub.cfg
- /grub/grub.cfg.future

> Check the `/samples/grub` folder for examples.

Essentially, we need 2 copy of the `/samples/grub/bootPart.grub.cfg` file,
one for the current root and one for the future root. What changes is the
order of the menu entries, the present is always the first entry. So we
have 1 file with A (current) and B (previous) and another file with 
B (current) and A (previous).

We can use `set default=0` too but this way the result should be more
understandable for the user.

After a successful update, `grub.cfg` and `grub.cfg.future` are swapped.

## Init Structure

Each root has a folder in init partition with the following structure:

- /vos-a/abroot.cfg
- /vos-a/config-<version>
- /vos-a/initrd.img-<version>
- /vos-a/System.map-<version>
- /vos-a/vmlinuz-<version>

> Check the `/samples/grub` folder for examples.

The `abroot.cfg` file is the same as the `rootPart.abroot.cfg` file but
with the `root=` parameter set to the correct UUID.

Note that this file is being loaded as a configuration using the `configfile`
command. Do not include a `menuentry`, otherwise it will result in
a submenu. (Good for advanced cases?).

# Development Notes

## Root Lifecycle

The root lifecycle is composed of 2 stages: 

- **Current**: The root which is currently being used by the system.
- **Future**: The root which will be used by the system after an update.

The "current" root should never be modified, in any way. The "future" root is
the root which is being modified by the update process and any other process
which may require root modification (e.g. kargs, fstab..).

## System Update / Root Re-generation

When developing new features which does not envolve the update process, it is 
important to consider the possibility of regenerating the root, as opposed to 
performing a specific update. This may be necessary, for example, when updating 
a kernel flag or fstab entry. To regenerate the root, developers should use the 
Containerfile and avoid taking data from the current root.

It is important to note that regenerating the root does not require pulling a 
new image; rather, the latest image in the storage can be used. In ABRoot, 
Prometheus (the container runtime used) makes this process easier, as 
generating an image from the Containerfile does not execute a pull if the image 
is already present in the store. During the update process, however, it is 
necessary to force the pull to take the updated image (this may not happen
during the development process).

In the case of a rollback, the AtomicSwap function should be used to swap the 
current grub configuration with the future one.

## Function Data Files

Some functions like kargs, fstab and pkg, need to store data to be used in the
future. Those should always be placed in the current `/etc/abroot/..` path.

If you are concerned that the current root should never be changed, you are 
correct but in this case the `/etc` path is a combined overlay of the root and 
respective `/etc` path files in the var partition:

```
/etc -> /var/lib/abroot/etc/a
```

Even if the root is read-only, the `/etc` path is not.
