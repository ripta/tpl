package main

import (
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

// Values is the merged key-value pairs
type Values map[string]interface{}

// Load will load key-values from fname, and merge them into the current object
func (v Values) Load(fname string) error {
	if fname == "" {
		return fmt.Errorf("Filename must not be empty")
	}
	log.Printf("Loading values from %s\n", fname)
	// Slurp the data file as one byteslice
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("Cannot read file %s: %v", fname, err)
	}
	// Parse the data file into values
	values := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &values); err != nil {
		return fmt.Errorf("Cannot parse values from %s: %v", fname, err)
	}
	// Merge top level only
	for km, vm := range values {
		v[km] = vm
	}
	return nil
}
