package main

import (
	"fmt"
	"os"

	"github.com/jon-codes/getopt/internal/testgen"
)

const (
	inpath  = "testdata/cases.json"
	outpath = "testdata/fixtures.json"
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

	err = testgen.ProcessCases(infile, outfile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error processing test cases: %v\n", err)
		os.Exit(1)
	}
}
