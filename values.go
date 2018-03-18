package main

import (
	"fmt"
	"io/ioutil"
	"log"

	yaml "gopkg.in/yaml.v2"
)

// Values is the merged key-value pairs
type Values map[string]interface{}

// LoadFile will load the contents of fname, parse it for key-value pairs,
// and merge them into the current object.
func (v Values) LoadFile(fname string) error {
	if fname == "" {
		return fmt.Errorf("Filename must not be empty")
	}
	log.Printf("Loading values from %s\n", fname)
	// Slurp the data file as one byteslice
	data, err := ioutil.ReadFile(fname)
	if err != nil {
		return fmt.Errorf("Cannot read file %s: %v", fname, err)
	}
	err = v.Load(data)
	if err != nil {
		return fmt.Errorf("Cannot parse values from %s: %v", fname, err)
	}
	return nil
}

// Load will parse the data string for key-value pairs, and merge them into
// the current object.
func (v Values) Load(data []byte) error {
	// Parse the data file into values
	values := make(map[string]interface{})
	if err := yaml.Unmarshal(data, &values); err != nil {
		return err
	}
	// Merge top level only
	for km, vm := range values {
		v[km] = vm
	}
	return nil
}
