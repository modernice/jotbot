#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)

BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ]; then
	git checkout main
fi

echo "Deleting opendocs-patch branches ..."
BRANCHES=$(git branch | grep opendocs-patch)

# check if there are any branches to delete
if [ -z "$BRANCHES" ]; then
	echo "No branches to delete"
	exit 0
fi

git branch -D "$BRANCHES"
