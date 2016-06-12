#!/usr/bin/env bash

cd ~/go-ur
 ./build/bin/gur --exec "admin.startRPC('0.0.0.0',9595,'*')" attach
