package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
)

var (
	dataFile = flag.String("values", "", "Comma-separated paths to YAML files containing values (only top-level keys are merged)")
	onError  = flag.String("on-error", "die", "What to do on render error: die, warn, quiet (stop processing without printing), ignore (continue rendering)")
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

	allValues := make(map[string]interface{})
	for _, fname := range dataFiles {
		if fname == "" {
			continue
		}
		log.Printf("Loading values from %s\n", fname)
		// Slurp the data file as one byteslice
		data, err := ioutil.ReadFile(fname)
		if err != nil {
			log.Fatalf("Cannot read file %s: %v", fname, err)
		}
		// Parse the data file into values
		values := make(map[string]interface{})
		if err := yaml.Unmarshal(data, &values); err != nil {
			log.Fatalf("Cannot parse values from %s: %v", fname, err)
		}
		// Merge top level only
		for km, vm := range values {
			allValues[km] = vm
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

	for _, fn := range flag.Args() {
		f, err := os.Open(fn)
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		fi, err := f.Stat()
		if err != nil {
			log.Fatalf("%v\n", err)
		}

		// Render files directly
		if !fi.IsDir() {
			render(allValues, fn, getOutputPath(*outFile, path.Base(fn)))
			continue
		}

		// Loop through each file in a directory and render it
		eis, err := f.Readdirnames(-1)
		if err != nil {
			log.Fatalf("%v", err)
		}
		for _, ei := range eis {
			render(allValues, filepath.Join(fn, ei), getOutputPath(*outFile, ei))
		}
	}
}

func getOutputPath(base, fn string) string {
	if base == "" || base == "-" {
		return "-"
	}
	if strings.HasSuffix(fn, ".tpl") {
		fn = strings.TrimSuffix(fn, ".tpl")
	} else if strings.HasSuffix(fn, ".tmpl") {
		fn = strings.TrimSuffix(fn, ".tmpl")
	}
	if strings.HasSuffix(base, "/") {
		return filepath.Join(base, fn)
	}
	if f, err := os.Open(base); err == nil {
		if fi, err := f.Stat(); err == nil {
			if fi.IsDir() {
				return filepath.Join(base, fn)
			}
		}
	}
	return base
}

func render(values map[string]interface{}, iname, oname string) {
	if oname == "" {
		log.Fatalf("Output name cannot be blank")
	}

	var out *os.File
	var err error
	if oname == "-" {
		out = os.Stdout
		log.Printf("Rendering %s\n", iname)
	} else {
		if strings.Contains(oname, "/") {
			if err := os.MkdirAll(path.Dir(oname), 0755); err != nil {
				log.Fatalf("Error creating directory for %q: %v", oname, err)
			}
		}

		out, err = os.OpenFile(oname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			log.Fatalf("Cannot open output file %q: %v", oname, err)
		}

		log.Printf("Rendering %s into %s\n", iname, oname)
		defer func() { out.Sync(); out.Close() }()
	}

	tpl, err := template.New(filepath.Base(iname)).Funcs(sprig.TxtFuncMap()).ParseFiles(iname)
	if err != nil {
		log.Fatalf("Cannot parse template %s: %v", iname, err)
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
			log.Printf("Cannot render template %s: %v", iname, err)
		default:
			log.Fatalf("Cannot render template %s: %v", iname, err)
		}
	}
}
