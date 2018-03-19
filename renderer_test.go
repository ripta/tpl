package main_test

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/blang/vfs"
	tpl "github.com/ripta/tpl"
)

type renderSpec struct {
	ins []string
	out string
}

type fileSpec struct {
	name    string
	content string
}

type fileTest struct {
	name      string
	ins       []fileSpec
	render    renderSpec
	renderErr string
	outs      []fileSpec
}

var fileTests = []fileTest{
	// Render one input file to one output file
	{
		name: "file-to-file",
		ins: []fileSpec{
			{"in/test.txt.tpl", "{{.foo}}-baz"},
		},
		render: renderSpec{
			[]string{"in/test.txt.tpl"},
			"out.txt",
		},
		outs: []fileSpec{
			{"out.txt", "bar-baz"},
		},
	},
	// Fails to render, because key is missing
	{
		name: "fail-missing-key",
		ins: []fileSpec{
			{"in/test.txt.tpl", "{{.foo}}-{{.hello}}-test"},
		},
		render: renderSpec{
			[]string{"in/test.txt.tpl"},
			"out.txt",
		},
		renderErr: `map has no entry for key "hello"`,
	},
	// Render an explicit list of input files to an output directory
	// (a trailing slash in `out/`)
	{
		name: "list-to-dir",
		ins: []fileSpec{
			{"in/test1.txt.tpl", "#1-{{.foo}}"},
			{"in/test2.txt.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in/test1.txt.tpl", "in/test2.txt.tpl"},
			"out/",
		},
		outs: []fileSpec{
			{"out/test1.txt", "#1-bar"},
			{"out/test2.txt", "#2-ripta"},
		},
	},
	// Succeed to write multiple input file into one single output file
	{
		name: "list-to-file",
		ins: []fileSpec{
			{"in/test1.txt.tpl", "#1-{{.foo}}"},
			{"in/test2.txt.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in/test1.txt.tpl", "in/test2.txt.tpl"},
			"out.txt",
		},
		outs: []fileSpec{
			{"out.txt", "#1-bar#2-ripta"},
		},
	},
	// Successfully handle multiple input file by writing to direcory
	// (a trailing slash in `render.out`)
	{
		name: "dir-to-dir",
		ins: []fileSpec{
			{"in/test1.txt.tpl", "#1-{{.foo}}"},
			{"in/test2.txt.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in"},
			"out/",
		},
		outs: []fileSpec{
			{"out/in/test1.txt", "#1-bar"},
			{"out/in/test2.txt", "#2-ripta"},
		},
	},
	// Succeeds, but a bit weird: because input is a directory, output is forced to a directory
	{
		name: "dir-to-dir2",
		ins: []fileSpec{
			{"in/test1.txt.tpl", "#1-{{.foo}}"},
			{"in/test2.txt.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in"},
			"out",
		},
		outs: []fileSpec{
			{"out/in/test1.txt", "#1-bar"},
			{"out/in/test2.txt", "#2-ripta"},
		},
	},
	// Successfully handle nested input files, preserving output structure
	{
		name: "nested-dir-to-dir",
		ins: []fileSpec{
			{"in/test1.txt.tpl", "#1-{{.foo}}"},
			{"in/shallow/test2.txt.tpl", "#2-{{.user.name}}"},
			{"in/rather/deep/nested/dirs/test3.txt.tpl", "#3-{{.price}}"},
		},
		render: renderSpec{
			[]string{"in"},
			"out/",
		},
		outs: []fileSpec{
			{"out/in/test1.txt", "#1-bar"},
			{"out/in/shallow/test2.txt", "#2-ripta"},
			{"out/in/rather/deep/nested/dirs/test3.txt", "#3-2.34"},
		},
	},
}

var staticValues = map[string]interface{}{
	"foo":   "bar",
	"price": 2.34,
	"user": map[string]string{
		"name": "ripta",
	},
}

func TestRendering(t *testing.T) {
	for _, test := range fileTests {
		test := test // range capture
		t.Run(test.name, func(t *testing.T) {
			if test.outs == nil && test.renderErr == "" {
				t.Fatal("Invalid test: either test.outs or test.rendererr must be non-empty")
			}

			tmpdir, err := ioutil.TempDir("", "tpl-test")
			if err != nil {
				t.Error(err)
			}
			t.Logf("Using %s as testing directory", tmpdir)
			if err := os.Chdir(tmpdir); err != nil {
				t.Error(err)
			}
			defer os.RemoveAll(tmpdir)

			for _, in := range test.ins {
				writeFile(t, in.name, in.content)
			}

			r := &tpl.Renderer{test.render.ins, true}
			err = r.Execute(test.render.out, staticValues)
			if err != nil {
				if test.renderErr == "" {
					t.Errorf("Unexpected error during execution: %v", err)
				} else if !strings.Contains(err.Error(), test.renderErr) {
					t.Errorf("Different error during execution; got %q, expected %q", err, test.renderErr)
				}
			} else if test.renderErr != "" {
				t.Errorf("Expected execution to fail with %v, but it succeeded", test.renderErr)
			}

			for _, out := range test.outs {
				content, err := ioutil.ReadFile(out.name)
				if err != nil {
					t.Error(err)
				}
				actual := string(content)
				if actual != out.content {
					dumpFS(t, tmpdir+"/")
					t.Errorf("Renderer output %s, expected %q, got %q", out.name, out.content, actual)
				}
			}
		})
	}
}

func dumpFS(t *testing.T, dir string) {
	var dumpRec func(string, int)
	dumpRec = func(dir string, depthLeft int) {
		if depthLeft < 0 {
			t.Errorf("dumpFS: recursed too deep")
		}
		d, err := os.Open(dir)
		fis, err := d.Readdir(-1)
		if err != nil {
			t.Errorf("Listing %s: %v", dir, err)
		}
		for _, fi := range fis {
			t.Logf("%s %7d %s", fi.Mode().String(), fi.Size(), dir+fi.Name())
			if fi.IsDir() {
				dumpRec(dir+fi.Name()+"/", depthLeft-1)
			}
		}
	}
	dumpRec(dir, 5)
}

func readFile(t *testing.T, fs vfs.Filesystem, fp string) string {
	content, err := vfs.ReadFile(fs, fp)
	if err != nil {
		t.Error(err)
	}
	return string(content)
}

func writeFile(t *testing.T, fp, content string) {
	err := os.MkdirAll(filepath.Dir(fp), 0755)
	if err != nil {
		t.Error(err)
	}

	f, err := os.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		t.Error(err)
	}

	n, err := f.Write([]byte(content))
	if err != nil {
		t.Error(err)
	}
	if n != len(content) {
		t.Errorf("Wrote %d bytes, expected %d bytes", n, len(content))
	}

	if err := f.Sync(); err != nil {
		t.Error(err)
	}

	if err = f.Close(); err != nil {
		t.Error(err)
	}
}
