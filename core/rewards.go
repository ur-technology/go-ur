package core

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ur-technology/go-ur/common"
	"github.com/ur-technology/go-ur/core/types"
)

// privileged addresses
var (
	PrivilegedAddressesReward = floatUrToWei("6000")
	SignupReward              = floatUrToWei("2000")
	MembersSingupRewards      = []*big.Int{
		floatUrToWei("60.60"),
		floatUrToWei("60.60"),
		floatUrToWei("121.21"),
		floatUrToWei("181.81"),
		floatUrToWei("303.03"),
		floatUrToWei("484.84"),
		floatUrToWei("787.91"),
	}

	TotalSingupRewards       = floatUrToWei("2000")
	privSendReceiveAddresses = map[string]string{
		"0x5d32e21bf3594aa66c205fde8dbee3dc726bd61d": "0xb76a8e9d411b75a3c0fd6fafaf9b4af9da1ec6d7",
		"0x9194d1fa799d9feb9755aadc2aa28ba7904b0efd": "0xc4411a0446b503db00d597a6c6a0643cec1e9eda",
		"0xab4b7eeb95b56bae3b2630525b4d9165f0cab172": "0x5a0df72c29822efc70ff1cc41a4da6d951330b71",
		"0xea82e994a02fb137ffaca8051b24f8629b478423": "0xe74be964a4d69951e3a74d997553c2063cf9ac04",
		"0xb1626c3fc1662410d85d83553d395cabba148be1": "0x29199be72e1f692ad64ff25ea55b3b44c6ee3bb6",
		"0x65afd2c418a1005f678f9681f50595071e936d7c": "0x08b57c894cd81e98638553a6cdfbede0cdcaa3ad",
		"0x49158a28df943acd20be7c8e758d8f4a9dc07d05": "0xe2c362411658e9f1eb76e5f40a2d717109a8b33c",
	}
	PrivilegedAddressesReceivers map[common.Address]common.Address
)

func init() {
	PrivilegedAddressesReceivers = make(map[common.Address]common.Address, len(privSendReceiveAddresses))
	for s, r := range privSendReceiveAddresses {
		PrivilegedAddressesReceivers[common.HexToAddress(s)] = common.HexToAddress(r)
	}
}

func floatUrToWei(ur string) *big.Int {
	u, _ := new(big.Float).SetString(ur)
	urFloat, _ := new(big.Float).SetString(common.Ether.String())
	r, _ := new(big.Float).Mul(u, urFloat).Int(nil)
	return r
}

// a signup transaction is signaled by the value 1 and the data in the following format:
//     when a privileged address signs a member
//         "01" - the current version of the message
//     when a member signs a member:
//         "01" - the current version of the message
//         8 bytes in big endian for the block number of signup transaction of the referring member
//         32 bytes for the hash of the signup transaction of the referring member
func refTxFromData(bc *BlockChain, d []byte) (*types.Transaction, error) {
	if len(d) < 1 {
		return nil, errInvalidChain
	}
	if d[0] != currentSignupMessageVersion {
		return nil, errInvalidChain
	}
	if len(d) == 1 {
		return nil, errNoMoreMembers
	}
	if len(d) == 41 {
		bn := binary.BigEndian.Uint64(d[1:])
		var txh common.Hash
		copy(txh[:], d[9:])
		return bc.GetBlockByNumber(bn).Transaction(txh), nil
	}
	return nil, errInvalidChain
}

func getSignupChain(bc *BlockChain, data []byte) ([]common.Address, error) {
	r := make([]common.Address, 0, 7)
	txdata := data
	for len(r) < 7 {
		tx, err := refTxFromData(bc, txdata)
		if err == errInvalidChain {
			return nil, err
		}
		if err == errNoMoreMembers {
			return r, nil
		}
		if tx.Value().Cmp(big.NewInt(1)) != 0 {
			return nil, errInvalidChain
		}
		to := tx.To()
		r = append(r, *to)
		txdata = tx.Data()
	}
	return r, nil
}

// SignupChain returns the signup chain up to 7 levels
func SignupChain(bc *BlockChain, tx *types.Transaction) ([]common.Address, error) {
	return getSignupChain(bc, tx.Data())
}

var (
	errNoMoreMembers               = errors.New("no more members in the chain")
	errInvalidChain                = errors.New("detected an invalid signup chain")
	errInvalidSignupMessageVersion = errors.New("invalid signup message version")
)

const currentSignupMessageVersion byte = 1

func isSignupTx(from common.Address, value *big.Int, data []byte) bool {
	return IsPrivilegedAddress(from) && value.Cmp(big.NewInt(1)) == 0 && len(data) > 0 && data[0] == currentSignupMessageVersion
}

func isSignupTransaction(tx *types.Transaction) bool {
	addr, _ := tx.From()
	data := tx.Data()
	return isSignupTx(addr, tx.Value(), data)
}

func IsPrivilegedAddress(address common.Address) bool {
	_, ok := PrivilegedAddressesReceivers[address]
	return ok
}

var (
	big9007 = new(big.Int).Mul(common.Ether, big.NewInt(9007))
	big10k  = new(big.Int).Mul(common.Ether, big.NewInt(10000))
	big1k   = new(big.Int).Mul(common.Ether, big.NewInt(1000))
)

func calculateTxManagementFee(nSignups, totaWei *big.Int) *big.Int {
	if nSignups.Cmp(common.Big0) == 0 {
		return big1k
	}
	avg := new(big.Int).Div(totaWei, nSignups)
	if avg.Cmp(big10k) <= 0 {
		return big1k
	}
	return common.Big0
}

func calculateBlockTotals(cNSignups, cTotalWei *big.Int, header *types.Header, uncles []*types.Header, txs []*types.Transaction) (*big.Int, *big.Int) {
	newNSignups := new(big.Int).Set(cNSignups)
	newTotalWei := new(big.Int).Set(cTotalWei)
	for _, r := range calculateAccumulatedRewards(header, uncles) {
		newTotalWei.Add(newTotalWei, r)
	}
	for _, t := range txs {
		if isSignupTransaction(t) {
			mngFee := calculateTxManagementFee(newNSignups, newTotalWei)
			newNSignups.Add(newNSignups, common.Big1)
			newTotalWei.Add(newTotalWei, new(big.Int).Add(big9007, mngFee))
		}
	}
	return newNSignups, newTotalWei
}

// returns number of sign
func UpdateBlockTotals(header *types.Header, uncles []*types.Header, txs []*types.Transaction) {
	header.NSignups, header.TotalWei = calculateBlockTotals(header.NSignups, header.TotalWei, header, uncles, txs)
}
