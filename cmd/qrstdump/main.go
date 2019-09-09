package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/kisom/goutils/die"
	"github.com/kisom/qrst"
)

func dumpImage(path string) {
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

	err = ioutil.WriteFile(path+".img", image, 0644)
	die.If(err)
}

func main() {
	flag.Parse()

	for _, path := range flag.Args() {
		dumpImage(path)
	}
}
