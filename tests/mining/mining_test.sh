#!/usr/bin/env bash

SOURCE_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
TEST_DATA_DIR=${SOURCE_DIR}/ur-test-data
rm -rf ${TEST_DATA_DIR}
mkdir ${TEST_DATA_DIR}
cp -r ${SOURCE_DIR}/privileged_keystore ${TEST_DATA_DIR}/keystore

gur --datadir ${TEST_DATA_DIR} --networkid 123 --nodiscover --maxpeers 0 init ${SOURCE_DIR}/genesis.json

gur --exec "loadScript('${SOURCE_DIR}/mining_test.js')" --datadir ${TEST_DATA_DIR} --networkid 123 --nodiscover --maxpeers 0 console
