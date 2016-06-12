#!/usr/bin/env bash

killall gur
rm -f /tmp/gur.log
sleep 3
mkdir -p ~/.ur
cd ~/.ur
rm -rf chaindata dapp gur.ipc history nodes # remove everythiing except nodekey and keystore
cd ~
rm -rf ~/go-ur
unzip -q go-ur.zip
cd ~/go-ur
make >/tmp/makegur.log
