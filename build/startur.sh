#!/usr/bin/env bash

if [ ! -f "build/env.sh" ]; then
    echo "$0 must be run from the root of the repository."
    exit 2
fi

killall gur
sleep 5
make gur
nohup build/bin/gur </dev/null >>/tmp/gur.log 2>&1 &
