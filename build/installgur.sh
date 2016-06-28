#!/usr/bin/env bash

killall gur
rm -f /tmp/gur.log
sleep 3
mkdir -p ~/.ur
cd ~/.ur
rm -rf `ls | grep -v 'nodekey\|keystore'`
rm -rf ~/go-ur
cd ~
unzip -q go-ur.zip
cd ~/go-ur
mkdir -p tmp
make gur > tmp/makegur.log
