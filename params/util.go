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

package params

import "math/big"

var (
	TestNetHomesteadBlock = big.NewInt(1)    // testnet homestead block
	MainNetHomesteadBlock = big.NewInt(1)   // mainnet homestead block
	HomesteadBlock        = MainNetHomesteadBlock // homestead block used to check against
)

func IsHomestead(blockNumber *big.Int) bool {
	// for unit tests TODO: flip to true after homestead is live
	if blockNumber == nil {
		return false
	}
	return blockNumber.Cmp(HomesteadBlock) >= 0
}
