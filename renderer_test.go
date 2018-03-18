package main_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/blang/vfs"
	"github.com/blang/vfs/memfs"
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
	name   string
	ins    []fileSpec
	render renderSpec
	outs   []fileSpec
}

var fileTests = []fileTest{
	//
	{
		name: "simple",
		ins: []fileSpec{
			{"/in/test.tpl", "{{.foo}}-baz"},
		},
		render: renderSpec{
			[]string{"/in/test.tpl"},
			"/out",
		},
		outs: []fileSpec{
			{"/out", "bar-baz"},
		},
	},
}

func TestRendering(t *testing.T) {
	for _, test := range fileTests {
		test := test // range capture
		t.Run(test.name, func(t *testing.T) {
			mem := memfs.Create()
			tpl.FS = mem
			for _, in := range test.ins {
				writeFile(t, mem, in.name, in.content)
			}

			v := make(map[string]interface{})
			v["foo"] = "bar"
			r := &tpl.Renderer{test.render.ins, true}
			err := r.Execute(test.render.out, v)
			if err != nil {
				t.Errorf("Execute: %v", err)
			}

			// dumpFS(t, mem, "/")
			for _, out := range test.outs {
				if actual := readFile(t, mem, out.name); actual != out.content {
					t.Errorf("Renderer output %q, expected %v, got %v", out.name, out.content, actual)
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
	dumpRec(dir, 3)
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
