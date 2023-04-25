# Setup Notes

## Root Structure

- /
- /abimage.abr
- /bin -> .system/usr/bin
- /boot
- /dev
- /etc -> .system/etc
- /home -> var/home
- /lib -> .system/usr/lib
- /lib32 -> .system/usr/lib32
- /lib64 -> .system/usr/lib64
- /libx32 -> .system/usr/libx32
- /media
- /mnt
- /opt
- /part-future
- /proc
- /root
- /run
- /sbin -> .system/usr/sbin
- /srv
- /sys
- /tmp
- /usr -> .system/usr
- /var -(mount)-> partData
- /tmp -> var/tmp

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
have 1 file with A (present) and B (future) and another file with B (present)
and A (future).

We can use `set default=0` too but this way the result should be more
understandable for the user.

After a successful update, `grub.cfg` and `grub.cfg.future` are swapped.

## Root Boot Structure

Each root has a `/.system/boot` folder with the following structure:

- /.system/boot
- /.system/boot/grub
- /.system/boot/grub/abroot.cfg

> Check the `/samples/grub` folder for examples.

The `abroot.cfg` file is the same as the `rootPart.abroot.cfg` file but
with the `root=` parameter set to the correct UUID.

Note that this file is being loaded as a configuration using the `configfile`
command. Do not include a `menuentry`, otherwise it will result in
a submenu. (Good for advanced cases?).
