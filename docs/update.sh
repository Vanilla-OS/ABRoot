#!/bib/sh

if [ $(basename $(pwd)) == "docs" ]; then
    cd ..
fi

godoc2md github.com/vanilla-os/abroot/core > docs/core.md
godoc2md github.com/vanilla-os/abroot/cmd > docs/cmd.md
godoc2md github.com/vanilla-os/abroot/extras > docs/extras.md
godoc2md github.com/vanilla-os/abroot/settings > docs/settings.md
godoc2md github.com/vanilla-os/abroot/tests > docs/tests.md
