{{/* This file requires preload-funcs.tpl to be preloaded. */}}
apiVersion: v1
kind: OAuth2Client
metadata:
    name: {{ template "encode-dex-id" "this-is-a-test" }}
    namespace: default
