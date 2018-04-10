package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"os/exec"

	yaml "gopkg.in/yaml.v2"
)

type execMap struct {
	Paths     []string       `yaml:"paths,omitempty"`
	Whitelist []*execSetting `yaml:"whitelist"`
}

type execSetting struct {
	Name   string `yaml:"name"`
	Path   string `yaml:"path,omitempty"`
	Stdin  bool   `yaml:"stdin"`
	Stdout bool   `yaml:"stdout"`
	Stderr bool   `yaml:"stderr"`
}

func loadExecMap(filename string) (*execMap, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("cannot read file %s: %v", filename, err)
	}

	var em execMap
	if err := yaml.Unmarshal(data, &em); err != nil {
		return nil, err
	}

	if len(em.Paths) > 0 {
		oldPaths := os.Getenv("PATH")
		defer os.Setenv("PATH", oldPaths)
	}
	for i, s := range em.Whitelist {
		if s.Name == "" {
			return nil, fmt.Errorf("whitelist item #%d in exec-map %q is missing a 'name' field", i, filename)
		}
		if !s.Stderr && !s.Stdout {
			return nil, fmt.Errorf("whitelist for %q in exec-map %q has neither 'stdout' nor 'stderr' enabled", s.Name, filename)
		}
		if s.Path == "" {
			p, err := exec.LookPath(s.Name)
			if err != nil {
				return nil, err
			}
			em.Whitelist[i].Path = p
		} else {
			e, err := os.Stat(s.Path)
			if err != nil {
				return nil, err
			}
			if m := e.Mode(); m.IsDir() || m&0111 == 0 {
				return nil, os.ErrPermission
			}
		}
	}

	return &em, nil
}

func (em *execMap) Get(name string) (*execSetting, error) {
	for _, s := range em.Whitelist {
		if s.Name == name {
			return s, nil
		}
	}
	return nil, fmt.Errorf("Executable %q is not in whitelist", name)
}

func (es *execSetting) Run(args []string, in io.Reader) (string, string, error) {
	msg := fmt.Sprintf("Executing %q with arguments %+v", es.Path, args)
	cmd := exec.Command(es.Path, args...)
	if in != nil {
		msg = msg + " and STDIN"
		cmd.Stdin = in
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	log.Println(msg)
	err := cmd.Run()
	if err != nil {
		log.Printf("Exited with error: %v", err)
	}
	return stdout.String(), stderr.String(), err
}
