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
	// Successfully render one input file to one output file
	{
		name: "simple-1to1",
		ins: []fileSpec{
			{"in/test.tpl", "{{.foo}}-baz"},
		},
		render: renderSpec{
			[]string{"in/test.tpl"},
			"out",
		},
		outs: []fileSpec{
			{"out", "bar-baz"},
		},
	},
	// Fails to render one file, because key is missing
	{
		name: "simple-fail-1",
		ins: []fileSpec{
			{"in/test.tpl", "{{.foo}}-{{.hello}}-test"},
		},
		render: renderSpec{
			[]string{"in/test.tpl"},
			"out",
		},
		renderErr: `map has no entry for key "hello"`,
	},
	// Successfully handle multiple input file by writing to direcory
	// (a trailing slash in `render.out`)
	{
		name: "multifile-to-multifile",
		ins: []fileSpec{
			{"in/test1.tpl", "#1-{{.foo}}"},
			{"in/test2.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in/test1.tpl", "in/test2.tpl"},
			"out/",
		},
		outs: []fileSpec{
			{"out/test1", "#1-bar"},
			{"out/test2", "#2-ripta"},
		},
	},
	// Succeed to write multiple input file into one single output file
	{
		name: "multifile-to-singlefile",
		ins: []fileSpec{
			{"in/test1.tpl", "#1-{{.foo}}"},
			{"in/test2.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in/test1.tpl", "in/test2.tpl"},
			"out",
		},
		outs: []fileSpec{
			{"out", "#1-bar#2-ripta"},
		},
	},
	// Successfully handle multiple input file by writing to direcory
	// (a trailing slash in `render.out`)
	{
		name: "directory-to-multifile",
		ins: []fileSpec{
			{"in/test1.tpl", "#1-{{.foo}}"},
			{"in/test2.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in"},
			"out/",
		},
		outs: []fileSpec{
			{"out/in/test1", "#1-bar"},
			{"out/in/test2", "#2-ripta"},
		},
	},
	// Succeeds, but a bit weird: because input is a directory, output is forced to a directory
	{
		name: "directory-to-pretendsinglefile",
		ins: []fileSpec{
			{"in/test1.tpl", "#1-{{.foo}}"},
			{"in/test2.tpl", "#2-{{.user.name}}"},
		},
		render: renderSpec{
			[]string{"in"},
			"out",
		},
	},
	// Successfully handle nested input files, preserving output structure
	{
		name: "nested-ok",
		ins: []fileSpec{
			{"in/test1.tpl", "#1-{{.foo}}"},
			{"in/shallow/test2.tpl", "#2-{{.user.name}}"},
			{"in/rather/deep/nested/dirs/test3.tpl", "#3-{{.price}}"},
		},
		render: renderSpec{
			[]string{"in"},
			"out",
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
			tmpdir, err := ioutil.TempDir("", "tpl-test")
			if err != nil {
				t.Error(err)
			}
			t.Logf("Using %s as testing directory", tmpdir)
			if err := os.Chdir(tmpdir); err != nil {
				t.Error(err)
			}
			defer os.RemoveAll(tmpdir)

			// mem := memfs.Create()
			// tpl.FS = mem
			for _, in := range test.ins {
				writeFile(t, tpl.FS, in.name, in.content)
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

			// dumpFS(t, tpl.FS, tmpdir+"/")
			for _, out := range test.outs {
				if actual := readFile(t, tpl.FS, out.name); actual != out.content {
					t.Errorf("Renderer output %s, expected %q, got %q", out.name, out.content, actual)
				}
			}
		})
	}
}

func dumpFS(t *testing.T, fs vfs.Filesystem, dir string) {
	var dumpRec func(string, int)
	dumpRec = func(dir string, depthLeft int) {
		if depthLeft < 0 {
			t.Errorf("dumpFS: recursed too deep")
		}
		fis, err := fs.ReadDir(dir)
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

func writeFile(t *testing.T, fs vfs.Filesystem, fp, content string) {
	err := vfs.MkdirAll(fs, filepath.Dir(fp), 0755)
	if err != nil {
		t.Error(err)
	}

	f, err := fs.OpenFile(fp, os.O_WRONLY|os.O_CREATE, 0644)
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
