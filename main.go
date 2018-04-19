package main

import (
	"flag"
	"fmt"
	"io"
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
	execMapFile := flag.String("exec-map-file", "", "File from which exec rules can be read")
	onError := flag.String("on-error", "die", "What to do on render error: die, ignore")
	outFile := flag.String("out", "-", "Output file (or '-' for STDOUT)")

	preloadFiles := make(stringSliceFlag, 0)
	flag.Var(&preloadFiles, "preload", "Additional files to preload")

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

	fm := funcMap()
	fm["exec"] = func(name string, args ...string) string {
		log.Fatalf("the 'exec' template function is disabled; you must specify -exec-map-file=FILE to enable it")
		return ""
	}
	if *execMapFile != "" {
		exmap, err := loadExecMap(*execMapFile)
		if err != nil {
			log.Fatal(err)
		}
		// log.Printf("%+v\n", exmap)
		fm["exec"] = func(name string, args ...string) string {
			exset, err := exmap.Get(name)
			if err != nil {
				log.Printf("could not exec %q %v: %v", name, args, err)
				return ""
			}
			var stdin io.Reader
			if exset.Stdin {
				stdin = strings.NewReader(args[len(args)-1])
				args = args[:len(args)-1]
			}
			stdout, stderr, err := exset.Run(args, stdin)
			if stderr != "" {
				log.Printf("exec %q %v, STDERR output was: %s", name, args, stderr)
			}
			if err != nil {
				log.Printf("exec %q %v failed with error: %v", name, args, err)
				return ""
			}
			if exset.Stdout {
				return stdout
			}
			if exset.Stderr {
				return stderr
			}
			return ""
		}
	}
	r := &Renderer{
		FuncMap:      fm,
		Inputs:       flag.Args(),
		PreloadFiles: preloadFiles,
		StopOnError:  (*onError != "ignore"),
	}
	if err := r.Execute(*outFile, allValues); err != nil {
		log.Fatal(err)
	}
}
