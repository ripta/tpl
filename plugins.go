package main

import (
	"errors"
	"fmt"
	"path"
	"plugin"
	"regexp"
	"text/template"
)

var re *regexp.Regexp

func init() {
	// match '_' | unicode_letter | unicode_digit
	// (ie characters valid in template identifiers)
	re = regexp.MustCompile(`^([_\pL\p{Nd}]+)\.so$`)
}

func loadPlugin(so *string, fm *template.FuncMap) error {
	matches := re.FindStringSubmatch(path.Base(*so))
	if matches == nil {
		return errors.New("invalid characters in filename - must use underscores and unicode letters/digits only")
	}
	ns := matches[1]

	plug, err := plugin.Open(*so)
	if err != nil {
		return err
	}
	sym, err := plug.Lookup("FuncMap")
	if err != nil {
		return err
	}
	f, ok := sym.(func() template.FuncMap)
	if !ok {
		return errors.New("could not assert symbol 'FuncMap' to be of type func() template.FuncMap")
	}

	for k, v := range f() {
		(*fm)[fmt.Sprintf("%s_%s", ns, k)] = v
	}
	return nil
}
