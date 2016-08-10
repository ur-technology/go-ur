#!/usr/bin/env bash

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DATA_DIR=${DIR}/ur-test-data
rm -rf ${TEST_DATA_DIR}
mkdir ${TEST_DATA_DIR}
cp -r ${DIR}/privileged_keystore ${TEST_DATA_DIR}/keystore

gur --exec "loadScript('${DIR}/mining_test.js')" --genesis ${DIR}/genesis.json --datadir ${TEST_DATA_DIR} --networkid 123 --nodiscover --maxpeers 0 console
