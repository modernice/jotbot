#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)/go

if [ -f "$ROOT/.env" ]; then
	set -o allexport
	source "$ROOT/.env"
	set +o allexport
fi

if [ -z "$OPENAI_API_KEY" ]; then
	echo "Missing OPENAI_API_KEY environment variable"
	exit 1
fi

BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ]; then
	git checkout main
fi

if [ "$BRANCH" = "opendocs-patch" ]; then
	git branch -D opendocs-patch
fi

go run "$ROOT/cmd/opendocs/main.go" "$ROOT"
