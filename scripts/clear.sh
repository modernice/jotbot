#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)

BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ]; then
	git checkout main
fi

echo "Deleting opendocs-patch branches ..."
git branch -D `git branch | grep opendocs-patch`
