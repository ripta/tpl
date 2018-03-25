package main

import (
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
)

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	f["toYaml"] = toYaml
	return f
}

func toYaml(v interface{}) string {
	d, err := yaml.Marshal(v)
	if err != nil {
		return "" // swallow in-template errors for now
	}
	return string(d)
}
