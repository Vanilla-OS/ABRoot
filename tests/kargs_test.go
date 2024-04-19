package tests

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/vanilla-os/abroot/core"
)

// TestKargsWrite tests the KargsWrite function by writing a string to a file,
// mimicking the kernel command line arguments.

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
	t.Log("TestKargsWrite: done")
}

// TestKargsRead tests the KargsRead function by reading the content of the file
// that was written by the TestKargsWrite function.
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
	t.Log("TestKargsRead: done")
}
