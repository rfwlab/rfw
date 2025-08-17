#!/bin/bash
set -e
cd "$(dirname "$0")"
bash build.sh
go run server/main.go "$@"
