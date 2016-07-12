#!/usr/bin/env bash

killall gur
sleep 3
cd ~/go-ur
nohup build/bin/gur </dev/null >>/tmp/gur.log 2>&1 &
sleep 3
