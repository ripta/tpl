{{ $ident := exec "aws" "sts" "get-caller-identity" "--output=json" | fromJson -}}
{{ $ident.Arn }}
