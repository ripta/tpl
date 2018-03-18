package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

// Build information
var (
	BuildDate    string
	BuildVersion string
)

func usage() {
	fmt.Fprintf(os.Stderr, "%s v%s built %s\n\n", os.Args[0], BuildVersion, BuildDate)
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [options...] <templates...>\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "where <templates...> may be one or more template files or directories.")
	fmt.Fprintf(os.Stderr, "Directories are processed only single depth.")
	fmt.Fprintf(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\n")
}

func main() {
	dataFile := flag.String("values", "", "Comma-separated paths to YAML files containing values (only top-level keys are merged)")
	onError := flag.String("on-error", "die", "What to do on render error: die, ignore")
	outFile := flag.String("out", "-", "Output file (or '-' for STDOUT)")

	valueMap := make(valueMapFlag)
	flag.Var(&valueMap, "value", "Additional values to inject in the form of key=value")

	// Parse command line flags
	flag.Usage = usage
	flag.Parse()

	dataFiles := []string{}
	if *dataFile != "" {
		dataFiles = strings.Split(*dataFile, ",")
	}

	allValues := make(Values)
	for _, fname := range dataFiles {
		if err := allValues.LoadFile(fname); err != nil {
			log.Fatal(err)
		}
	}

	if len(valueMap) > 0 {
		log.Printf("Loading values from command line\n")
		for km, vm := range valueMap {
			allValues[km] = vm
		}
	}

	if flag.NArg() < 1 {
		usage()
		log.Fatalln("At least one <template> path is required.")
	}

	r := &Renderer{
		Inputs:      flag.Args(),
		StopOnError: (*onError != "ignore"),
	}
	if err := r.Execute(*outFile, allValues); err != nil {
		log.Fatal(err)
	}
}
