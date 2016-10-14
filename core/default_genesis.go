// Copyright 2015 The go-ur Authors
// This file is part of the go-ur library.
//
// The go-ur library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ur library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ur library. If not, see <http://www.gnu.org/licenses/>.

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
// NOTE: Uncompressed Genesis Block Params:{"Nonce":"0x0000000000000032","Timestamp":"0x5800E836","ParentHash":"0x0000000000000000000000000000000000000000000000000000000000000000","ExtraData":"0x11bbe8db4e347b4e8c937c1c8370e4b5ed33adb3db69cbdb7a38e1e50b1b82fa","GasLimit":"0x61A8","Difficulty":"0x1000","Mixhash":"0x0000000000000000000000000000000000000000000000000000000000000000","Coinbase":"0x0000000000000000000000000000000000000000","Alloc":{"0x482cf297b08d4523c97ec3a54e80d2d07acd76fa":{"Balance":"1000000000000000000"},"0xcc74e28cec33a784c5cd40e14836dd212a937045":{"Balance":"1000000000000000000"},"0xc07a55758f896449805bae3851f57e25bb7ee7ef":{"Balance":"1000000000000000000"},"0x48a24dd26a32564e2697f25fc8605700ec4c0337":{"Balance":"1000000000000000000"},"0x3cac5f7909f9cb666cc4d7ef32047b170e454b16":{"Balance":"1000000000000000000"},"0x0827d93936df936134dd7b7acaeaea04344b11f2":{"Balance":"1000000000000000000"},"0xa63e936e0eb36c103f665d53bd7ca9c31ec7e1ad":{"Balance":"1000000000000000000"}}}
const defaultGenesisBlock = "H4sIAAAAAAAA/62Sy2pcMQyG3+Wss5Aty7KzS5uQLJqSRV9AlmVyYC4lcwoTwrx7lZlVYSinUBm0sKVPF/8f0/f9Tm26neAIfxjG6Wb6MW/tsMj25zmACsBDwewPL/Jmu+VJDq9XUv/dnPhwXN7kXhY5A0NozUpvyTCx+6IVWYMWZLDUyDqi9Ia95aqtNxYsFoyghVbiEAc+yuHbvJ2XMy+Hu+J39/MYs/7aLO+XKpfSz/Px9T9O8nU/75ocri31r3l3m81ep9sPz0ol6oiVG5SeKKJWNkUhXwT02IFFO2cf06O/yEYuXxiuYE83jlPlZLGoM1C4JCXtCSwk/8zeY4ji24VEa3Fen4ipjFJzSrUANTEsFAaxRWqNzdjGSlwqEpO3kQUjZW80Vx6RhpYMxACmSQGRV+JQRWlwhTpcGjln1dS9G4zgUgqf+qHUQl6JgxK5V6y+qeEuoLfKrjcV8wMJk8PCiCtxktEcY2ANswbAkTN1wtZZpSoGU7YgfQXudPoNVQHfIL4DAAA="
