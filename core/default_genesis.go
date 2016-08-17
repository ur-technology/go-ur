// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package core

import (
	"compress/gzip"
	"encoding/base64"
	"io"
	"strings"
)

func NewDefaultGenesisReader() (io.Reader, error) {
	return gzip.NewReader(base64.NewDecoder(base64.StdEncoding, strings.NewReader(defaultGenesisBlock)))
}

// NOTE: 7 privileged address will be allocated 1 UR each
// NOTE: Uncompressed Genesis Block Params:{"Nonce":"0x0000000000000032","Timestamp":"0x57A963C4","ParentHash":"0x0000000000000000000000000000000000000000000000000000000000000000","ExtraData":"0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa","GasLimit":"0x61A8","Difficulty":"0x1000","Mixhash":"0x0000000000000000000000000000000000000000000000000000000000000000","Coinbase":"0x0000000000000000000000000000000000000000","Alloc":{"0x5d32e21bf3594aa66c205fde8dbee3dc726bd61d":{"Balance":"1000000000000000000"},"0x9194d1fa799d9feb9755aadc2aa28ba7904b0efd":{"Balance":"1000000000000000000"},"0xab4b7eeb95b56bae3b2630525b4d9165f0cab172":{"Balance":"1000000000000000000"},"0xea82e994a02fb137ffaca8051b24f8629b478423":{"Balance":"1000000000000000000"},"0xb1626c3fc1662410d85d83553d395cabba148be1":{"Balance":"1000000000000000000"},"0x65afd2c418a1005f678f9681f50595071e936d7c":{"Balance":"1000000000000000000"},"0x49158a28df943acd20be7c8e758d8f4a9dc07d05":{"Balance":"1000000000000000000"}}}
const defaultGenesisBlock = "H4sIAAAAAAAA/62RvW5bMQyF3+XOGURRlKhsbhK0Q1t06AuQIoVcwD9FfAu4CPzule2pgFHcAuXAgTr8KB6+T18P++bT4xRO4Y/AOD1M3+edHxfZ/bgKqGxqxqc0Hr7Jm++XT3J8vdP67zGIL6flTZ5lkSsQQNXZNDmmMjK3iqVBYyzBk5Ibopiiaa5NTYsgOzgFBeXYZQA/yvHzvJuXKy/Dhkftee59bj+3y6/blNvoL/Pp9T9u8nSY9yrHe6b+tW+z3R7a9Ph+cdowegTtSDWJ5NxioG4XR9zRWolZLYNd1B9kK7cTwh3s+WHgKtRk0KXUarW71kIkYi2KRNZRDkmD97U40aTFB4aUsoqjxoyBImmyCpl6aKJQ4kqcC0evY88QuwKW3qUJBwKNqXOOVVPhFHElTiHH3LA3yDkmCMZkjERoWGl8TAUSq8NKXCbpFlsClqGhngv3mhk6BaoUCnjFbKWtxKUKxMN16zWhNItBvTT2Qmzck1RroVigFbjz+TcQEoH0vgMAAA=="
