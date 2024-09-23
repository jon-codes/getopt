package main

import (
	"fmt"
	"os"

	"github.com/jon-codes/getopt/testgen"
)

const (
	inpath  = "testdata/cases.csv"
	outpath = "testdata/fixtures.csv"
)

func main() {
	infile, err := os.Open(inpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening infile: %v\n", err)
		os.Exit(1)
	}
	defer infile.Close()

	outfile, err := os.Create(outpath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error opening outfile: %v\n", err)
		os.Exit(1)
	}
	defer outfile.Close()

	testgen.ProcessCases(infile, outfile)
}
