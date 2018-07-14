package main

import (
	"fmt"
	"text/template"
)

func FuncMap() template.FuncMap {
	f := make(template.FuncMap)
	f["foo"] = foo
	return f
}

func foo() string {
	return fmt.Sprint("foo")
}
