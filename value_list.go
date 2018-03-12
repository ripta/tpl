package main

import (
	"fmt"
	"strings"
)

type valueMapFlag map[string]string

func (m *valueMapFlag) String() string {
	return fmt.Sprintf("%v", map[string]string(*m))
}

func (m *valueMapFlag) Set(value string) error {
	if value == "" {
		return fmt.Errorf("Value must not be blank")
	}
	if !strings.Contains(value, "=") {
		return fmt.Errorf("Value %q must be in the format 'key=value'", value)
	}
	c := strings.SplitN(value, "=", 2)
	(*m)[c[0]] = c[1]
	return nil
}
