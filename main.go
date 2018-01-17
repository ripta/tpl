package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"text/template"

	"gopkg.in/yaml.v2"
)

var (
	dataFile = flag.String("values", "", "YAML file containing values")
	onError  = flag.String("on-error", "die", "What to do on render error: die, warn, quiet (stop processing without printing), ignore (continue rendering)")
	outFile  = flag.String("out", "-", "Output file (or '-' for STDOUT)")
)

func usage() {
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  %s [options...] <templates...>\n\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
}

func main() {
	// Parse command line flags
	flag.Usage = usage
	flag.Parse()

	// Flag validation
	if *dataFile == "" {
		log.Fatalf("-values is required")
	}

	// Slurp the data file as one byteslice
	data, err := ioutil.ReadFile(*dataFile)
	if err != nil {
		log.Fatalf("Cannot read file %s: %v", dataFile, err)
	}

	// Parse the data file into values
	var values map[string]interface{}
	if err := yaml.Unmarshal(data, &values); err != nil {
		log.Fatalf("Cannot parse values from %s: %v", dataFile, err)
	}

	// Render either to STDOUT or to a file
	var out *os.File
	if *outFile == "" || *outFile == "-" {
		out = os.Stdout
	} else {
		out, err = os.OpenFile(*outFile, os.O_RDWR|os.O_CREATE, 0644)
		if err != nil {
			log.Fatalf("Cannot open output file %s: %v", *outFile, err)
		}
	}
	defer func() { out.Sync(); out.Close() }()

	if flag.NArg() < 1 {
		usage()
		os.Exit(-1)
	}

	for _, f := range flag.Args() {
		tpl, err := template.ParseFiles(f)
		if err != nil {
			log.Fatalf("Cannot parse template %s: %v", f, err)
		}

		if *onError == "ignore" {
			tpl.Option("missingkey=zero")
		} else {
			tpl.Option("missingkey=error")
		}

		err = tpl.Execute(out, values)
		if err != nil {
			switch *onError {
			case "ignore", "quiet":
				// print nothing, but still fail
			case "warn":
				log.Printf("Cannot render template %s: %v", f, err)
			default:
				log.Fatalf("Cannot render template %s: %v", f, err)
			}
		}
	}
}
