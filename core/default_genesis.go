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

//{"Nonce":"0x0000000000000032","Timestamp":"0x57366EE1","ParentHash":"0x0000000000000000000000000000000000000000000000000000000000000000","ExtraData":"0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa","GasLimit":"0x61A8","Difficulty":"0x10000000","Mixhash":"0x0000000000000000000000000000000000000000000000000000000000000000","Coinbase":"0x8805317929d0a8cd1e7a19a4a2523b821ed05e42","Alloc":{"0x8805317929d0a8cd1e7a19a4a2523b821ed05e42":{"Code":"","Storage":null,"Balance":"7000000000000000000"}}}
const defaultGenesisBlock = "H4sIAAAAAAAA/62RwWrDMAyG38XnHuwojp3eujZsh20MtheQbGU1OMlIXMgoefe56WEMeilMB4Es6ftl6Sxeh96x2Ao5yz8GhdiIj9DxlLD7Wgu0gapqGpUTbzhyn55wOt5ovd8ysZnTiAdMuAKVImLrqWQoTfbW1WCcchaM5JI0ewD0BJ6q2pEng2BZsZakyBYtZuAjTs+hC2nlVWpn89shtG1wp5i+ryq/8i9hPv7jb/ZD6Amn62KtlRqUqYvaS7TOKzaoaiyx0AXkcRV7qbm8LHwX4+DE9nxPV67eD/4ilQHvaRjxMwf9KcaNeMCI1/uaG3Muy/IDIA1NsgECAAA="
// const defaultGenesisBlock = "H4sIAAAJbogA/6yRzUrEMBSF3yXrWSRN09zMbuyILlQEfYF78+ME0lbaDFSGvruxXYigiwHPIpDknO/k58Keht56tmd85j8kK7Zjr7HzU8bufTUoLXl7VG3ZeMbR9/kep9Mv0etViLdzHvGIGVegEEQeHNVe1rqMYI3UVliQmvualHdSoiPpqDGWHGmU4IVXnARBFbAA73B6iF3MK68RByhrxxhCtOeUP7aW7/rHOJ/+8TbtEHvCaXtYEZQLAYTmNTUNNwAgQzA1OAStvTIamqoyruQOKQ2W7S/XpIq7HdxXVQG85GHEtzLpzynt2A0m3P5X/HXYZVk+AwAA//9qlMK7BgIAAA=="
