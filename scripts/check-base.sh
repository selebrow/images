#!/bin/bash

git show origin/main:meta.json > meta-old.json
oldTag=$(jq -r '.build.base.images."ubuntu-noble".tags.notag' meta-old.json)
newTag=$(jq -r '.build.base.images."ubuntu-noble".tags.notag' meta.json)

if [ "$oldTag" == "$newTag" ]; then
    exit 1
else
    exit 0
fi
