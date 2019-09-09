package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kisom/goutils/die"
	"github.com/kisom/qrst"
)

func scanFile(path string) {
	file, err := os.Open(path)
	die.If(err)
	defer file.Close()

	qrstFile, err := qrst.LoadFile(file)
	die.If(err)
	fmt.Printf("%s QRST file loaded.\n", qrst.CapacityToString(qrstFile.Header.DiskCapacity))

	image, err := qrstFile.Assemble()
	if err != nil {
		die.If(err)
	}

	fmt.Printf("Image is %d bytes\n", len(image))
}

func main() {
	flag.Parse()

	for _, path := range flag.Args() {
		scanFile(path)
	}
}
