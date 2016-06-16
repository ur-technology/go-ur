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
// NOTE: Uncompressed Genesis Block Params:{"Nonce":"0x0000000000000032","Timestamp":"0x57366EE1","ParentHash":"0x0000000000000000000000000000000000000000000000000000000000000000","ExtraData":"0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa","GasLimit":"0x61A8","Difficulty":"0x10000000","Mixhash":"0x0000000000000000000000000000000000000000000000000000000000000000","Coinbase":"0x78c7a643923363023a192afc49473eed637b02cf","Alloc":{"0x8805317929d0a8cd1e7a19a4a2523b821ed05e42":{"Code":"","Storage":null,"Balance":"7000000000000000000"}}}
const defaultGenesisBlock = "H4sIAAAAAAAA/62RTWvDMAyG/4vPPdiW44/eujZsh20Mtj8g20pryMdIXMgo+e9zm8MY9DKYDgLLep/Xki/sdegDsS3jM/8VINmGfaSOpozd562hMqB1XYty8YYj9fkJp9Md6d+jEOs5j3jAjDegEN6TjV4RKFOyDQ5MEMGC4aR8RREAo4fotQs+eoNgSVDFvfBWNliAjzg9py7lG0+LnS21Q2qaFM5t/lpdfuxf0nz6x2n2Q+o9TutijQ0GtQInATRwCSicxCYopwwQRQ3Gcxmaotu17RDY9lJU1vIKhHHSRY42REGm6FChrCSUIQVFXpGS1+79EK9WBfCehxGP5dCf23bDHrDF9X/NnXcuy/INmzAjlQECAAA="
