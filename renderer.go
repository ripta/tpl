package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
)

type Renderer struct {
	Inputs      []string
	StopOnError bool
}

func (r *Renderer) Execute(values map[string]interface{}, out string) error {
	for _, fn := range r.Inputs {
		f, err := os.Open(fn)
		if err != nil {
			return err
		}

		fi, err := f.Stat()
		if err != nil {
			return err
		}

		// Render files directly
		if !fi.IsDir() {
			err := r.render(values, fn, r.getOutputPath(out, path.Base(fn)))
			if err != nil {
				return err
			}
			continue
		}

		// Loop through each file in a directory and render it
		eis, err := f.Readdirnames(-1)
		if err != nil {
			return err
		}
		for _, ei := range eis {
			err := r.render(values, filepath.Join(fn, ei), r.getOutputPath(out, ei))
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *Renderer) getOutputPath(base, fn string) string {
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

func (r *Renderer) render(values map[string]interface{}, iname, oname string) error {
	if oname == "" {
		return errors.New("Output name cannot be blank")
	}

	var out *os.File
	var err error
	if oname == "-" {
		out = os.Stdout
		log.Printf("Rendering %s to STDOUT\n", iname)
	} else {
		if strings.Contains(oname, "/") {
			if err := os.MkdirAll(path.Dir(oname), 0755); err != nil {
				return fmt.Errorf("Error creating directory for %q: %v", oname, err)
			}
		}

		out, err = os.OpenFile(oname, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return fmt.Errorf("Cannot open output file %q: %v", oname, err)
		}

		log.Printf("Rendering %s into %s\n", iname, oname)
		defer func() { out.Sync(); out.Close() }()

	}

	tpl, err := template.New(filepath.Base(iname)).Funcs(sprig.TxtFuncMap()).ParseFiles(iname)
	if err != nil {
		return fmt.Errorf("Cannot parse template %s: %v", iname, err)
	}

	if r.StopOnError {
		tpl.Option("missingkey=error")
	} else {
		tpl.Option("missingkey=zero")
	}

	return tpl.Execute(out, values)
}
