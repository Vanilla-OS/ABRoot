#!/bib/sh

if [ $(basename $(pwd)) == "docs" ]; then
    cd ..
fi

godoc2md github.com/vanilla-os/abroot/core > docs/abroot.md
godoc2md github.com/vanilla-os/abroot/core > docs/core.md
godoc2md github.com/vanilla-os/abroot/core > docs/cmd.md
godoc2md github.com/vanilla-os/abroot/core > docs/extras.md
godoc2md github.com/vanilla-os/abroot/core > docs/settings.md
godoc2md github.com/vanilla-os/abroot/core > docs/tests.md
