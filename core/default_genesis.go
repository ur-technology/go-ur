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
// NOTE: Uncompressed Genesis Block Params: {"Nonce":"0x0000000000000032","Timestamp":"0x57843B6A","ParentHash":"0x0000000000000000000000000000000000000000000000000000000000000000","ExtraData":"0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa","GasLimit":"0x61A8","Difficulty":"0x1000000","Mixhash":"0x0000000000000000000000000000000000000000000000000000000000000000","Coinbase":"0x99cdff4e7d73962bf4aeca40d2265ab530ec792b","Alloc":{"0x99cdff4e7d73962bf4aeca40d2265ab530ec792b":{"Code":"","Storage":null,"Balance":"7000000000000000000"}}}
const defaultGenesisBlock = "H4sIAAAAAAAA/62RTWvDMAyG/4vPPcQfie3e+jG2wzYG2x+QbLk1OMlIXMgo+e/zEnYY9FKYDgLLep9Xlq/ste8csS2rpupPSME27CO2NGZoP5eGWhsl982uXLzBQF1+gvF8Q3p/FOLDlAc4QoYFyDkiGY+KpNIlG2eldtwZqStSWJOXEjxKj4116FGDNMSprpCjEQEK8BHG59jGvPAavjOldowhRHdJ+Wt1+TV/idP5H99y6GOHMK5rtdb5EBRpr6VtBAYF5EBVXoimBqxlRU5bgUW3S6l3bHu9R1W6D73/sSqA99wPcCqH7pLShu0hwfq7+sac8zx/A4oSBLL/AQAA"
