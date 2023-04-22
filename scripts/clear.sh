#!/bin/sh

ROOT=$(git rev-parse --show-toplevel)

BRANCH=$(git branch --show-current)
if [ "$BRANCH" != "main" ]; then
	git checkout main
fi

echo "Deleting jotbot-patch branches ..."
BRANCHES=$(git branch | grep jotbot-patch)

if [ -z "$BRANCHES" ]; then
	echo "No branches to delete"
	exit 0
fi

echo "$BRANCHES" | xargs git branch -D
