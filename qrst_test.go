package qrst

import (
	"fmt"
	"os"
	"testing"
)

var (
	testFile     = "DIAGS80B._01"
	testFileSize = 720 * 1024
)

func TestRead(t *testing.T) {
	fr, err := os.Open(testFile)
	if err != nil {
		t.Fatal(err)
	}

	file, err := LoadFile(fr)
	if err != nil {
		t.Fatal(err)
	}

	image, err := file.Assemble()
	if err != nil {
		t.Fatal(err)
	}

	if len(image) != testFileSize {
		t.Fatal("qrst: failed to produce a valid image")
	}

	fmt.Println("Checksum:", checksum(image))

	err = file.VerifyChecksum()
	if err != nil {
		t.Fatal(err)
	}
}
