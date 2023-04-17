#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)/go

BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ]; then
	git checkout main
fi

if [ "$BRANCH" = "opendocs-patch" ]; then
	echo "Deleting opendocs-patch branch ..."
	git branch -D opendocs-patch
fi
