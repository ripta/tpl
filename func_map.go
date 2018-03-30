package main

import (
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
)

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	f["toYaml"] = toYaml
	f["trimLeft"] = trimLeft
	f["trimRight"] = trimRight
	return f
}

func toYaml(v interface{}) string {
	d, err := yaml.Marshal(v)
	if err != nil {
		return "" // swallow in-template errors for now
	}
	return string(d)
}

func trimLeft(cut, s string) string {
	return strings.TrimLeft(s, cut)
}

func trimRight(cut, s string) string {
	return strings.TrimRight(s, cut)
}
