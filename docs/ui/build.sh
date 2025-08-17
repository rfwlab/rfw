#!/bin/bash
set -e
cd "$(dirname "$0")"
cp $(go env GOROOT)/misc/wasm/wasm_exec.js .
rm -rf docs
mkdir docs
cp -r ../api ../guide ../index.md ../getting-started.md ../legacy.md ../sidebar.json docs/
GOARCH=wasm GOOS=js go build -o main.wasm main.go
