{{/* This file acts should only contain definitions. For example, an encoding function: */}}
{{ define "encode-dex-id" }}{{ fnv64sum . | b32enc | trimRight "=" | lower }}{{ end }}
