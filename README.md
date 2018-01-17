# tpl

A very simplistic CLI tool that allows rendering arbitrary text/template files,
pulling data in from any YAML file.

There's a docker image:

```
docker pull ripta/tpl
```

Or, get it from the source:

```
go get github.com/ripta/tpl
```

Run it:

```
tpl -values=data/a.yaml -out=rendered.txt data/template.txt
```

Multiple templates can be provided on the command line. Each template is
rendered individually using the same values file, and into the same output
file. The default `-` output file can be used to write to STDOUT.


