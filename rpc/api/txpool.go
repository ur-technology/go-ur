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

package api

import (
	"github.com/ur/go-ur/ur"
	"github.com/ur/go-ur/rpc/codec"
	"github.com/ur/go-ur/rpc/shared"
	"github.com/ur/go-ur/xur"
)

const (
	TxPoolApiVersion = "1.0"
)

var (
	// mapping between methods and handlers
	txpoolMapping = map[string]txpoolhandler{
		"txpool_status": (*txPoolApi).Status,
	}
)

// net callback handler
type txpoolhandler func(*txPoolApi, *shared.Request) (interface{}, error)

// txpool api provider
type txPoolApi struct {
	xur     *xur.XEth
	ur *ur.UR
	methods  map[string]txpoolhandler
	codec    codec.ApiCoder
}

// create a new txpool api instance
func NewTxPoolApi(xur *xur.XEth, ur *ur.UR, coder codec.Codec) *txPoolApi {
	return &txPoolApi{
		xur:     xur,
		ur: ur,
		methods:  txpoolMapping,
		codec:    coder.New(nil),
	}
}

// collection with supported methods
func (self *txPoolApi) Methods() []string {
	methods := make([]string, len(self.methods))
	i := 0
	for k := range self.methods {
		methods[i] = k
		i++
	}
	return methods
}

// Execute given request
func (self *txPoolApi) Execute(req *shared.Request) (interface{}, error) {
	if callback, ok := self.methods[req.Method]; ok {
		return callback(self, req)
	}

	return nil, shared.NewNotImplementedError(req.Method)
}

func (self *txPoolApi) Name() string {
	return shared.TxPoolApiName
}

func (self *txPoolApi) ApiVersion() string {
	return TxPoolApiVersion
}

func (self *txPoolApi) Status(req *shared.Request) (interface{}, error) {
	pending, queue := self.ur.TxPool().Stats()
	return map[string]int{
		"pending": pending,
		"queued":  queue,
	}, nil
}
