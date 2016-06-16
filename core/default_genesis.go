// Copyright 2014 The go-ethereum Authors
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

// NOTE: coinbase will receive 7 UR
// NOTE: Uncompressed Genesis Block Params:{"Nonce":"0x0000000000000032","Timestamp":"0x576328E1","ParentHash":"0x0000000000000000000000000000000000000000000000000000000000000000","ExtraData":"0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa","GasLimit":"0x61A8","Difficulty":"0x10000000","Mixhash":"0x0000000000000000000000000000000000000000000000000000000000000000","Coinbase":"0x78c7a643923363023a192afc49473eed637b02cf","Alloc":{"0x8805317929d0a8cd1e7a19a4a2523b821ed05e42":{"Code":"","Storage":null,"Balance":"7000000000000000000"}}}
const defaultGenesisBlock = "H4sIAAAAAAAA/62Rz27DIAzG34VzD4AhQG9dW22HbZq0vYABp0XKnymhUqYq7z6aHKZJvUyaD5Yw/n4fNlf22neB2Jbxif8KkGzDPlJLY8b2c2nQpgJpj6JcvOFAXX7C8XxH+vcoxOOUBzxgxgUohPdko1cEypRsgwMTRLBgOCmvKQJg9BB95YKP3iBYEqS5F97KGgvwEcfn1Ka88Cqxs6V2SHWdwqXJX6vLj/1Lms7/OM2+T53HcV2sscFgpcBJgAq4BBROYh2UUwaIYgXGcxnqots1TR/Y9lpU1nINwjjpIkcboiBTdKhQagllSEGRa1Ly1r3v482qAN5zP+CpHLpL02zYAza4/q+58855nr8B2ikFkQECAAA="
