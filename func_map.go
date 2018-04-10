package main

import (
	"encoding/json"
	"hash/fnv"
	"log"
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
	f["fromJson"] = fromJson
	f["fromYaml"] = fromYaml
	f["toYaml"] = toYaml
	f["trimLeft"] = trimLeft
	f["trimRight"] = trimRight
	return f
}

func baseConvert(from, to int, s string) string {
	i, err := strconv.ParseInt(s, from, 64)
	if err != nil {
		log.Printf("warning: cannot parse integer base %d from string %q: %v", from, s, err)
		return "0"
	}

	return strconv.FormatInt(i, to)
}

func fnv64sum(s string) string {
	return string(fnv.New64().Sum([]byte(s)))
}

func fromJson(p string) interface{} {
	var v interface{}
	err := json.Unmarshal([]byte(p), &v)
	if err != nil {
		log.Printf("cannot unmarshal JSON: %v (in %q)", err, p)
	}
	return v
}

func fromYaml(p string) interface{} {
	var v interface{}
	err := yaml.Unmarshal([]byte(p), &v)
	if err != nil {
		log.Printf("cannot unmarshal YAML: %v (in %q)", err, p)
	}
	return v
}

func toYaml(v interface{}) string {
	d, err := yaml.Marshal(v)
	if err != nil {
		log.Printf("cannot marshal YAML: %v", err)
	}
	return string(d)
}

func trimLeft(cut, s string) string {
	return strings.TrimLeft(s, cut)
}

func trimRight(cut, s string) string {
	return strings.TrimRight(s, cut)
}
