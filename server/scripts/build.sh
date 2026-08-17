#!/usr/bin/env bash

swag init \
  --v3.1 \
  --parseDependency \
  --parseInternal \
  --md ./additional-docs \
  -g ./main.go \
  -o ./docs \
  --ot json,yaml

./fixswagger.sh

go build .
