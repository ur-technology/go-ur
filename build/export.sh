#!/usr/bin/env bash

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

git archive --prefix go-ur/ --format zip --output /tmp/go-ur.zip master
