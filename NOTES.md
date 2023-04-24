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
    "digest":"sha256:e9f2773ab60fabf4e456e4549d86e345b38754f1f8397183ce4dc28d52bab66e","
    timestamp":"2023-04-23T15:13:19.066903531+02:00","
    image":"registry.vanillaos.org/vanillaos/desktop:main"
}
```
