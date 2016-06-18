#!/usr/bin/env bash

killall gur
sleep 3
cd ~/go-ur

if [ $1 = "staging" ]; then
  extra_params="--port 19596 --genesis build/testnet/testnet_genesis.json"
else
  extra_params=""
fi
echo "starting gur with extra params: $extra_params"
nohup build/bin/gur $extra_params </dev/null >>/tmp/gur.log 2>&1 &
sleep 3
