package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	dataFile = flag.String("values", "", "Comma-separated paths to YAML files containing values (only top-level keys are merged)")
	onError  = flag.String("on-error", "die", "What to do on render error: die, ignore")
	outFile  = flag.String("out", "-", "Output file (or '-' for STDOUT)")
	valueMap = make(valueMapFlag)
)

var BuildDate string
var BuildVersion string

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

func init() {
	flag.Var(&valueMap, "value", "Additional values to inject in the form of key=value")
}

func main() {
	// Parse command line flags
	flag.Usage = usage
	flag.Parse()

	dataFiles := []string{}
	if *dataFile != "" {
		dataFiles = strings.Split(*dataFile, ",")
	}

	allValues := make(Values)
	for _, fname := range dataFiles {
		if err := allValues.Load(fname); err != nil {
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
		os.Exit(-1)
	}

	r := &Renderer{
		Inputs:      flag.Args(),
		StopOnError: (*onError != "ignore"),
	}
	if err := r.Execute(allValues, *outFile); err != nil {
		log.Fatal(err)
	}
}
