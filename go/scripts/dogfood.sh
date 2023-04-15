#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)

if [ -f "$ROOT/.env" ]; then
	source "$ROOT/.env"
fi

if [ -z "$OPENAI_API_KEY" ]; then
	echo "Missing OPENAI_API_KEY environment variable"
	exit 1
fi

LIMIT=0

if [ -n "$OPENDOCS_LIMIT" ]; then
	LIMIT=$OPENDOCS_LIMIT
fi

go run "$ROOT/cmd/opendocs/main.go" "$ROOT" --key $OPENAI_API_KEY --limit $OPENDOCS_LIMIT
