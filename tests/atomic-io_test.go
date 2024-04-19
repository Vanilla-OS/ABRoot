package tests

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/google/uuid"
	"github.com/vanilla-os/abroot/core"
)

// TestAtomicSwap tests the AtomicSwap function by creating 2 files and swapping
// them. As a result, the 2 files should change their locations.
func TestAtomicSwap(t *testing.T) {
	tmpfile, err := ioutil.TempFile("", uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := tmpfile.Write([]byte("ABRoot")); err != nil {
		t.Fatal(err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatal(err)
	}

	file, err := os.Open(tmpfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	newfile, err := ioutil.TempFile("", uuid.New().String())
	if err != nil {
		t.Fatal(err)
	}

	if _, err := newfile.Write([]byte("ABRoot")); err != nil {
		t.Fatal(err)
	}

	if err := newfile.Close(); err != nil {
		t.Fatal(err)
	}

	newfile, err = os.Open(newfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	err = core.AtomicSwap(file.Name(), newfile.Name())
	if err != nil {
		t.Fatal(err)
	}

	_, err = ioutil.ReadFile(file.Name())
	if err != nil {
		t.Fatal(err)
	}

	t.Log("TestAtomicSwap: done")
}
