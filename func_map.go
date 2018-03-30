package main

import (
	"hash/fnv"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig"
	"gopkg.in/yaml.v2"
)

func funcMap() template.FuncMap {
	f := sprig.TxtFuncMap()
	f["baseConvert"] = baseConvert
	f["fnv64sum"] = fnv64sum
	f["toYaml"] = toYaml
	f["trimLeft"] = trimLeft
	f["trimRight"] = trimRight
	return f
}

func baseConvert(from, to int, s string) string {
	i, err := strconv.ParseInt(s, from, 64)
	if err != nil {
		return ""
	}

	return strconv.FormatInt(i, to)
}

func fnv64sum(s string) string {
	return string(fnv.New64().Sum([]byte(s)))
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
