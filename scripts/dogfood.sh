#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)
SCRIPTS="$ROOT/scripts"

if [ -f "$ROOT/.env" ]; then
	set -o allexport
	source "$ROOT/.env"
	set +o allexport
fi

if [ -z "$OPENAI_API_KEY" ]; then
	echo "Missing OPENAI_API_KEY environment variable"
	exit 1
fi

go run "$ROOT/cmd/jotbot/main.go" generate -v "$ROOT"
