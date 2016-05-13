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
	"bytes"
	"encoding/json"
	"math/big"

	"fmt"

	"github.com/ur/go-ur/common"
	"github.com/ur/go-ur/common/natspec"
	"github.com/ur/go-ur/ur"
	"github.com/ur/go-ur/rlp"
	"github.com/ur/go-ur/rpc/codec"
	"github.com/ur/go-ur/rpc/shared"
	"github.com/ur/go-ur/xur"
	"gopkg.in/fatih/set.v0"
)

const (
	EthApiVersion = "1.0"
)

// ur api provider
// See https://github.com/ur/wiki/wiki/JSON-RPC
type urApi struct {
	xur     *xur.XEth
	ur *ur.UR
	methods  map[string]urhandler
	codec    codec.ApiCoder
}

// ur callback handler
type urhandler func(*urApi, *shared.Request) (interface{}, error)

var (
	urMapping = map[string]urhandler{
		"eth_accounts":                            (*urApi).Accounts,
		"eth_blockNumber":                         (*urApi).BlockNumber,
		"eth_getBalance":                          (*urApi).GetBalance,
		"eth_protocolVersion":                     (*urApi).ProtocolVersion,
		"eth_coinbase":                            (*urApi).Coinbase,
		"eth_mining":                              (*urApi).IsMining,
		"eth_syncing":                             (*urApi).IsSyncing,
		"eth_gasPrice":                            (*urApi).GasPrice,
		"eth_getStorage":                          (*urApi).GetStorage,
		"eth_storageAt":                           (*urApi).GetStorage,
		"eth_getStorageAt":                        (*urApi).GetStorageAt,
		"eth_getTransactionCount":                 (*urApi).GetTransactionCount,
		"eth_getBlockTransactionCountByHash":      (*urApi).GetBlockTransactionCountByHash,
		"eth_getBlockTransactionCountByNumber":    (*urApi).GetBlockTransactionCountByNumber,
		"eth_getUncleCountByBlockHash":            (*urApi).GetUncleCountByBlockHash,
		"eth_getUncleCountByBlockNumber":          (*urApi).GetUncleCountByBlockNumber,
		"eth_getData":                             (*urApi).GetData,
		"eth_getCode":                             (*urApi).GetData,
		"eth_getNatSpec":                          (*urApi).GetNatSpec,
		"eth_sign":                                (*urApi).Sign,
		"eth_sendRawTransaction":                  (*urApi).SubmitTransaction,
		"eth_submitTransaction":                   (*urApi).SubmitTransaction,
		"eth_sendTransaction":                     (*urApi).SendTransaction,
		"eth_signTransaction":                     (*urApi).SignTransaction,
		"eth_transact":                            (*urApi).SendTransaction,
		"eth_estimateGas":                         (*urApi).EstimateGas,
		"eth_call":                                (*urApi).Call,
		"eth_flush":                               (*urApi).Flush,
		"eth_getBlockByHash":                      (*urApi).GetBlockByHash,
		"eth_getBlockByNumber":                    (*urApi).GetBlockByNumber,
		"eth_getTransactionByHash":                (*urApi).GetTransactionByHash,
		"eth_getTransactionByBlockNumberAndIndex": (*urApi).GetTransactionByBlockNumberAndIndex,
		"eth_getTransactionByBlockHashAndIndex":   (*urApi).GetTransactionByBlockHashAndIndex,
		"eth_getUncleByBlockHashAndIndex":         (*urApi).GetUncleByBlockHashAndIndex,
		"eth_getUncleByBlockNumberAndIndex":       (*urApi).GetUncleByBlockNumberAndIndex,
		"eth_getCompilers":                        (*urApi).GetCompilers,
		"eth_compileSolidity":                     (*urApi).CompileSolidity,
		"eth_newFilter":                           (*urApi).NewFilter,
		"eth_newBlockFilter":                      (*urApi).NewBlockFilter,
		"eth_newPendingTransactionFilter":         (*urApi).NewPendingTransactionFilter,
		"eth_uninstallFilter":                     (*urApi).UninstallFilter,
		"eth_getFilterChanges":                    (*urApi).GetFilterChanges,
		"eth_getFilterLogs":                       (*urApi).GetFilterLogs,
		"eth_getLogs":                             (*urApi).GetLogs,
		"eth_hashrate":                            (*urApi).Hashrate,
		"eth_getWork":                             (*urApi).GetWork,
		"eth_submitWork":                          (*urApi).SubmitWork,
		"eth_submitHashrate":                      (*urApi).SubmitHashrate,
		"eth_resend":                              (*urApi).Resend,
		"eth_pendingTransactions":                 (*urApi).PendingTransactions,
		"eth_getTransactionReceipt":               (*urApi).GetTransactionReceipt,
	}
)

// create new urApi instance
func NewEthApi(xur *xur.XEth, ur *ur.UR, codec codec.Codec) *urApi {
	return &urApi{xur, ur, urMapping, codec.New(nil)}
}

// collection with supported methods
func (self *urApi) Methods() []string {
	methods := make([]string, len(self.methods))
	i := 0
	for k := range self.methods {
		methods[i] = k
		i++
	}
	return methods
}

// Execute given request
func (self *urApi) Execute(req *shared.Request) (interface{}, error) {
	if callback, ok := self.methods[req.Method]; ok {
		return callback(self, req)
	}

	return nil, shared.NewNotImplementedError(req.Method)
}

func (self *urApi) Name() string {
	return shared.EthApiName
}

func (self *urApi) ApiVersion() string {
	return EthApiVersion
}

func (self *urApi) Accounts(req *shared.Request) (interface{}, error) {
	return self.xur.Accounts(), nil
}

func (self *urApi) Hashrate(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xur.HashRate()), nil
}

func (self *urApi) BlockNumber(req *shared.Request) (interface{}, error) {
	num := self.xur.CurrentBlock().Number()
	return newHexNum(num.Bytes()), nil
}

func (self *urApi) GetBalance(req *shared.Request) (interface{}, error) {
	args := new(GetBalanceArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return self.xur.AtStateNum(args.BlockNumber).BalanceAt(args.Address), nil
}

func (self *urApi) ProtocolVersion(req *shared.Request) (interface{}, error) {
	return self.xur.EthVersion(), nil
}

func (self *urApi) Coinbase(req *shared.Request) (interface{}, error) {
	return newHexData(self.xur.Coinbase()), nil
}

func (self *urApi) IsMining(req *shared.Request) (interface{}, error) {
	return self.xur.IsMining(), nil
}

func (self *urApi) IsSyncing(req *shared.Request) (interface{}, error) {
	origin, current, height := self.ur.Downloader().Progress()
	if current < height {
		return map[string]interface{}{
			"startingBlock": newHexNum(big.NewInt(int64(origin)).Bytes()),
			"currentBlock":  newHexNum(big.NewInt(int64(current)).Bytes()),
			"highestBlock":  newHexNum(big.NewInt(int64(height)).Bytes()),
		}, nil
	}
	return false, nil
}

func (self *urApi) GasPrice(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xur.DefaultGasPrice().Bytes()), nil
}

func (self *urApi) GetStorage(req *shared.Request) (interface{}, error) {
	args := new(GetStorageArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return self.xur.AtStateNum(args.BlockNumber).State().SafeGet(args.Address).Storage(), nil
}

func (self *urApi) GetStorageAt(req *shared.Request) (interface{}, error) {
	args := new(GetStorageAtArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return self.xur.AtStateNum(args.BlockNumber).StorageAt(args.Address, args.Key), nil
}

func (self *urApi) GetTransactionCount(req *shared.Request) (interface{}, error) {
	args := new(GetTxCountArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	count := self.xur.AtStateNum(args.BlockNumber).TxCountAt(args.Address)
	return fmt.Sprintf("%#x", count), nil
}

func (self *urApi) GetBlockTransactionCountByHash(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	block := self.xur.EthBlockByHash(args.Hash)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Transactions())), nil
}

func (self *urApi) GetBlockTransactionCountByNumber(req *shared.Request) (interface{}, error) {
	args := new(BlockNumArg)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xur.EthBlockByNumber(args.BlockNumber)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Transactions())), nil
}

func (self *urApi) GetUncleCountByBlockHash(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xur.EthBlockByHash(args.Hash)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Uncles())), nil
}

func (self *urApi) GetUncleCountByBlockNumber(req *shared.Request) (interface{}, error) {
	args := new(BlockNumArg)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xur.EthBlockByNumber(args.BlockNumber)
	if block == nil {
		return nil, nil
	}
	return fmt.Sprintf("%#x", len(block.Uncles())), nil
}

func (self *urApi) GetData(req *shared.Request) (interface{}, error) {
	args := new(GetDataArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	v := self.xur.AtStateNum(args.BlockNumber).CodeAtBytes(args.Address)
	return newHexData(v), nil
}

func (self *urApi) Sign(req *shared.Request) (interface{}, error) {
	args := new(NewSigArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	v, err := self.xur.Sign(args.From, args.Data, false)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (self *urApi) SubmitTransaction(req *shared.Request) (interface{}, error) {
	args := new(NewDataArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	v, err := self.xur.PushTx(args.Data)
	if err != nil {
		return nil, err
	}
	return v, nil
}

// JsonTransaction is returned as response by the JSON RPC. It contains the
// signed RLP encoded transaction as Raw and the signed transaction object as Tx.
type JsonTransaction struct {
	Raw string `json:"raw"`
	Tx  *tx    `json:"tx"`
}

func (self *urApi) SignTransaction(req *shared.Request) (interface{}, error) {
	args := new(NewTxArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	// nonce may be nil ("guess" mode)
	var nonce string
	if args.Nonce != nil {
		nonce = args.Nonce.String()
	}

	var gas, price string
	if args.Gas != nil {
		gas = args.Gas.String()
	}
	if args.GasPrice != nil {
		price = args.GasPrice.String()
	}
	tx, err := self.xur.SignTransaction(args.From, args.To, nonce, args.Value.String(), gas, price, args.Data)
	if err != nil {
		return nil, err
	}

	data, err := rlp.EncodeToBytes(tx)
	if err != nil {
		return nil, err
	}

	return JsonTransaction{"0x" + common.Bytes2Hex(data), newTx(tx)}, nil
}

func (self *urApi) SendTransaction(req *shared.Request) (interface{}, error) {
	args := new(NewTxArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	// nonce may be nil ("guess" mode)
	var nonce string
	if args.Nonce != nil {
		nonce = args.Nonce.String()
	}

	var gas, price string
	if args.Gas != nil {
		gas = args.Gas.String()
	}
	if args.GasPrice != nil {
		price = args.GasPrice.String()
	}
	v, err := self.xur.Transact(args.From, args.To, nonce, args.Value.String(), gas, price, args.Data)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func (self *urApi) GetNatSpec(req *shared.Request) (interface{}, error) {
	args := new(NewTxArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	var jsontx = fmt.Sprintf(`{"params":[{"to":"%s","data": "%s"}]}`, args.To, args.Data)
	notice := natspec.GetNotice(self.xur, jsontx, self.ur.HTTPClient())

	return notice, nil
}

func (self *urApi) EstimateGas(req *shared.Request) (interface{}, error) {
	_, gas, err := self.doCall(req.Params)
	if err != nil {
		return nil, err
	}

	// TODO unwrap the parent method's ToHex call
	if len(gas) == 0 {
		return newHexNum(0), nil
	} else {
		return newHexNum(common.String2Big(gas)), err
	}
}

func (self *urApi) Call(req *shared.Request) (interface{}, error) {
	v, _, err := self.doCall(req.Params)
	if err != nil {
		return nil, err
	}

	// TODO unwrap the parent method's ToHex call
	if v == "0x0" {
		return newHexData([]byte{}), nil
	} else {
		return newHexData(common.FromHex(v)), nil
	}
}

func (self *urApi) Flush(req *shared.Request) (interface{}, error) {
	return nil, shared.NewNotImplementedError(req.Method)
}

func (self *urApi) doCall(params json.RawMessage) (string, string, error) {
	args := new(CallArgs)
	if err := self.codec.Decode(params, &args); err != nil {
		return "", "", err
	}

	return self.xur.AtStateNum(args.BlockNumber).Call(args.From, args.To, args.Value.String(), args.Gas.String(), args.GasPrice.String(), args.Data)
}

func (self *urApi) GetBlockByHash(req *shared.Request) (interface{}, error) {
	args := new(GetBlockByHashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	block := self.xur.EthBlockByHash(args.BlockHash)
	if block == nil {
		return nil, nil
	}
	return NewBlockRes(block, self.xur.Td(block.Hash()), args.IncludeTxs), nil
}

func (self *urApi) GetBlockByNumber(req *shared.Request) (interface{}, error) {
	args := new(GetBlockByNumberArgs)
	if err := json.Unmarshal(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	block := self.xur.EthBlockByNumber(args.BlockNumber)
	if block == nil {
		return nil, nil
	}
	return NewBlockRes(block, self.xur.Td(block.Hash()), args.IncludeTxs), nil
}

func (self *urApi) GetTransactionByHash(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	tx, bhash, bnum, txi := self.xur.EthTransactionByHash(args.Hash)
	if tx != nil {
		v := NewTransactionRes(tx)
		// if the blockhash is 0, assume this is a pending transaction
		if bytes.Compare(bhash.Bytes(), bytes.Repeat([]byte{0}, 32)) != 0 {
			v.BlockHash = newHexData(bhash)
			v.BlockNumber = newHexNum(bnum)
			v.TxIndex = newHexNum(txi)
		}
		return v, nil
	}
	return nil, nil
}

func (self *urApi) GetTransactionByBlockHashAndIndex(req *shared.Request) (interface{}, error) {
	args := new(HashIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xur.EthBlockByHash(args.Hash)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xur.Td(raw.Hash()), true)
	if args.Index >= int64(len(block.Transactions)) || args.Index < 0 {
		return nil, nil
	} else {
		return block.Transactions[args.Index], nil
	}
}

func (self *urApi) GetTransactionByBlockNumberAndIndex(req *shared.Request) (interface{}, error) {
	args := new(BlockNumIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xur.EthBlockByNumber(args.BlockNumber)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xur.Td(raw.Hash()), true)
	if args.Index >= int64(len(block.Transactions)) || args.Index < 0 {
		// return NewValidationError("Index", "does not exist")
		return nil, nil
	}
	return block.Transactions[args.Index], nil
}

func (self *urApi) GetUncleByBlockHashAndIndex(req *shared.Request) (interface{}, error) {
	args := new(HashIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xur.EthBlockByHash(args.Hash)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xur.Td(raw.Hash()), false)
	if args.Index >= int64(len(block.Uncles)) || args.Index < 0 {
		// return NewValidationError("Index", "does not exist")
		return nil, nil
	}
	return block.Uncles[args.Index], nil
}

func (self *urApi) GetUncleByBlockNumberAndIndex(req *shared.Request) (interface{}, error) {
	args := new(BlockNumIndexArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	raw := self.xur.EthBlockByNumber(args.BlockNumber)
	if raw == nil {
		return nil, nil
	}
	block := NewBlockRes(raw, self.xur.Td(raw.Hash()), true)
	if args.Index >= int64(len(block.Uncles)) || args.Index < 0 {
		return nil, nil
	} else {
		return block.Uncles[args.Index], nil
	}
}

func (self *urApi) GetCompilers(req *shared.Request) (interface{}, error) {
	var lang string
	if solc, _ := self.xur.Solc(); solc != nil {
		lang = "Solidity"
	}
	c := []string{lang}
	return c, nil
}

func (self *urApi) CompileSolidity(req *shared.Request) (interface{}, error) {
	solc, _ := self.xur.Solc()
	if solc == nil {
		return nil, shared.NewNotAvailableError(req.Method, "solc (solidity compiler) not found")
	}

	args := new(SourceArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	contracts, err := solc.Compile(args.Source)
	if err != nil {
		return nil, err
	}
	return contracts, nil
}

func (self *urApi) NewFilter(req *shared.Request) (interface{}, error) {
	args := new(BlockFilterArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	id := self.xur.NewLogFilter(args.Earliest, args.Latest, args.Skip, args.Max, args.Address, args.Topics)
	return newHexNum(big.NewInt(int64(id)).Bytes()), nil
}

func (self *urApi) NewBlockFilter(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xur.NewBlockFilter()), nil
}

func (self *urApi) NewPendingTransactionFilter(req *shared.Request) (interface{}, error) {
	return newHexNum(self.xur.NewTransactionFilter()), nil
}

func (self *urApi) UninstallFilter(req *shared.Request) (interface{}, error) {
	args := new(FilterIdArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	return self.xur.UninstallFilter(args.Id), nil
}

func (self *urApi) GetFilterChanges(req *shared.Request) (interface{}, error) {
	args := new(FilterIdArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	switch self.xur.GetFilterType(args.Id) {
	case xur.BlockFilterTy:
		return NewHashesRes(self.xur.BlockFilterChanged(args.Id)), nil
	case xur.TransactionFilterTy:
		return NewHashesRes(self.xur.TransactionFilterChanged(args.Id)), nil
	case xur.LogFilterTy:
		return NewLogsRes(self.xur.LogFilterChanged(args.Id)), nil
	default:
		return []string{}, nil // reply empty string slice
	}
}

func (self *urApi) GetFilterLogs(req *shared.Request) (interface{}, error) {
	args := new(FilterIdArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	return NewLogsRes(self.xur.Logs(args.Id)), nil
}

func (self *urApi) GetLogs(req *shared.Request) (interface{}, error) {
	args := new(BlockFilterArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	return NewLogsRes(self.xur.AllLogs(args.Earliest, args.Latest, args.Skip, args.Max, args.Address, args.Topics)), nil
}

func (self *urApi) GetWork(req *shared.Request) (interface{}, error) {
	self.xur.SetMining(true, 0)
	ret, err := self.xur.RemoteMining().GetWork()
	if err != nil {
		return nil, shared.NewNotReadyError("mining work")
	} else {
		return ret, nil
	}
}

func (self *urApi) SubmitWork(req *shared.Request) (interface{}, error) {
	args := new(SubmitWorkArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}
	return self.xur.RemoteMining().SubmitWork(args.Nonce, common.HexToHash(args.Digest), common.HexToHash(args.Header)), nil
}

func (self *urApi) SubmitHashrate(req *shared.Request) (interface{}, error) {
	args := new(SubmitHashRateArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return false, shared.NewDecodeParamError(err.Error())
	}
	self.xur.RemoteMining().SubmitHashrate(common.HexToHash(args.Id), args.Rate)
	return true, nil
}

func (self *urApi) Resend(req *shared.Request) (interface{}, error) {
	args := new(ResendArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	from := common.HexToAddress(args.Tx.From)

	pending := self.ur.TxPool().GetTransactions()
	for _, p := range pending {
		if pFrom, err := p.FromFrontier(); err == nil && pFrom == from && p.SigHash() == args.Tx.tx.SigHash() {
			self.ur.TxPool().RemoveTx(common.HexToHash(args.Tx.Hash))
			return self.xur.Transact(args.Tx.From, args.Tx.To, args.Tx.Nonce, args.Tx.Value, args.GasLimit, args.GasPrice, args.Tx.Data)
		}
	}

	return nil, fmt.Errorf("Transaction %s not found", args.Tx.Hash)
}

func (self *urApi) PendingTransactions(req *shared.Request) (interface{}, error) {
	txs := self.ur.TxPool().GetTransactions()

	// grab the accounts from the account manager. This will help with determining which
	// transactions should be returned.
	accounts, err := self.ur.AccountManager().Accounts()
	if err != nil {
		return nil, err
	}

	// Add the accouns to a new set
	accountSet := set.New()
	for _, account := range accounts {
		accountSet.Add(account.Address)
	}

	var ltxs []*tx
	for _, tx := range txs {
		if from, _ := tx.FromFrontier(); accountSet.Has(from) {
			ltxs = append(ltxs, newTx(tx))
		}
	}

	return ltxs, nil
}

func (self *urApi) GetTransactionReceipt(req *shared.Request) (interface{}, error) {
	args := new(HashArgs)
	if err := self.codec.Decode(req.Params, &args); err != nil {
		return nil, shared.NewDecodeParamError(err.Error())
	}

	txhash := common.BytesToHash(common.FromHex(args.Hash))
	tx, bhash, bnum, txi := self.xur.EthTransactionByHash(args.Hash)
	rec := self.xur.GetTxReceipt(txhash)
	// We could have an error of "not found". Should disambiguate
	// if err != nil {
	// 	return err, nil
	// }
	if rec != nil && tx != nil {
		v := NewReceiptRes(rec)
		v.BlockHash = newHexData(bhash)
		v.BlockNumber = newHexNum(bnum)
		v.TransactionIndex = newHexNum(txi)
		return v, nil
	}

	return nil, nil
}
