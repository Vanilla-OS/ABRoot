package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/vanilla-os/abroot/core"
)

func TestKargsWrite(t *testing.T) {
	core.KargsPath = fmt.Sprintf("%s/kargs-%s", os.TempDir(), uuid.New().String())

	err := core.KargsWrite("test")
	if err != nil {
		t.Error(err)
	}

	content, err := ioutil.ReadFile(core.KargsPath)
	if err != nil {
		t.Error(err)
	}

	t.Log(string(content))
}

func TestKargsRead(t *testing.T) {
	core.KargsPath = fmt.Sprintf("%s/kargs-%s", os.TempDir(), uuid.New().String())

	err := core.KargsWrite("test")
	if err != nil {
		t.Error(err)
	}

	content, err := core.KargsRead()
	if err != nil {
		t.Error(err)
	}

	t.Log(content)
}
