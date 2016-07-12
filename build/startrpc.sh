#!/usr/bin/env bash

cd ~/go-ur
 ./build/bin/gur --exec "admin.startRPC('127.0.0.1',9595,'*')" attach
