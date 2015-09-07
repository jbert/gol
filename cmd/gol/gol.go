package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/jbert/gol"
)

type options struct {
	displayResult bool
	fileName      string
}

func (o options) validate() error {
	if o.fileName == "" {
		return fmt.Errorf("Must specify filename")
	}
	return nil
}

func main() {
	o := options{}

	flag.BoolVar(&o.displayResult, "display-result", false, "Show result evaluation")
	flag.StringVar(&o.fileName, "filename", "", "Name of file to evaluate")
	flag.Parse()

	err := o.validate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bad options: %s\n\n", err)
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\n")
		os.Exit(-1)
	}

	g := gol.New()
	n, err := g.EvalFile(o.fileName)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to eval: %s\n", err)
		os.Exit(-1)
	}
	if o.displayResult {
		fmt.Fprintf(os.Stdout, "%s\n", n)
	}
}
