// Copyright 2015 The go-ethereum Authors
// This file is part of go-ethereum.
//
// go-ethereum is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// go-ethereum is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with go-ethereum. If not, see <http://www.gnu.org/licenses/>.

package utils

import "github.com/ur-technology/go-ur/p2p/discover"

// FrontierBootNodes are the enode URLs of the P2P bootstrap nodes running on
// the Frontier network.
var FrontierBootNodes = []*discover.Node{
	// ETH/DEV Go Bootnodes
	discover.MustParseNode("enode://fcf730cf678d6296ffa75a1b2c06aa07d9558788762d0bbefbc209ccbfb4e840f7dcfc2f7a188eb2e65056d989de3722df3fc4df286eb3690d4586992c1c6d82@45.55.7.66:19595"), // bootnode1
	discover.MustParseNode("enode://d846b3c0445b7a91cfeb56fbeaece55ca9e559a6e5810cc41c54e2b88790fa7a24444508f16eb983630da1367ab73a6db1b705cc36134d9e61a2df070284d3f4@45.55.7.75:19595"), // bootnode2
}

// TestNetBootNodes are the enode URLs of the P2P bootstrap nodes running on the
// Morden test network.
var TestNetBootNodes = []*discover.Node{
	// ETH/DEV Go Bootnodes
	discover.MustParseNode("enode://d1778b1d1e3c2e7053310b9749ceffc28ea8d5fd0f066f1ffd8b227e9be4f9a0dcb5340ca16003c3c54b664bd12f146a510b3fc58c5140303e333e836f0c4bb6@138.68.63.204:19595"), // testbootnode1
	discover.MustParseNode("enode://ee5e03c62179f6b3269ace0b9cee23aa29724251027de4b507080f2f793e0119e5e5838ddce47e3c5b5733fb08a65cb60bf43aeb4a6493c8a74eb20bf207bc38@138.68.56.175:19595"), // testbootnode2
}
