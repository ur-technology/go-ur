package core

import (
	"encoding/binary"
	"errors"
	"math/big"

	"github.com/ur-technology/go-ur/common"
	"github.com/ur-technology/go-ur/core/types"
)

type rewards struct {
	URFutureFundFee      *big.Int
	ManagementFee        *big.Int
	SignupReward         *big.Int
	MembersSignupRewards []*big.Int
	TotalSignupRewards   *big.Int
	MinerReward          *big.Int
	Total                *big.Int
}

// privileged addresses
var (
	FullRatio     = big.NewInt(100)
	ScaledRewards = map[*big.Int]rewards{
		FullRatio: rewards{
			URFutureFundFee: floatUrToWei("5000"),
			ManagementFee:   floatUrToWei("1000"),
			SignupReward:    floatUrToWei("2000"),
			MembersSignupRewards: []*big.Int{
				floatUrToWei("60.60"),
				floatUrToWei("60.60"),
				floatUrToWei("121.21"),
				floatUrToWei("181.81"),
				floatUrToWei("303.03"),
				floatUrToWei("484.84"),
				floatUrToWei("787.91"),
			},
			MinerReward:        BlockReward,
			TotalSignupRewards: floatUrToWei("2000"),
			Total:              new(big.Int).Add(floatUrToWei("9000"), BlockReward),
		},
	}

	privSendReceiveAddresses = map[string]receiverAddressPairString{
		"0x482cf297b08d4523c97ec3a54e80d2d07acd76fa": receiverAddressPairString{
			receiver: "0x59ab9bb134b529709333f7ae68f3f93c204d280b",
			urff:     "46c0b8e0e95a772ad8764d3190a34cd4a60c7a98",
		},
		"0xcc74e28cec33a784c5cd40e14836dd212a937045": receiverAddressPairString{
			receiver: "0x0ec37d90610b7665517a2d813dc85a7f83852aee",
			urff:     "ac5fbbd56b1d6a31ad722de419433eeb5b9a9fc4",
		},
		"0xc07a55758f896449805bae3851f57e25bb7ee7ef": receiverAddressPairString{
			receiver: "0x78021bd6fb0f0353bb49e2cc63a8aea051c902ca",
			urff:     "57b1f656e88fc66e8fe1cf0eb65ce045004777f4",
		},
		"0x48a24dd26a32564e2697f25fc8605700ec4c0337": receiverAddressPairString{
			receiver: "0xb8c4f8e04d3341690cfb9ebc11246bd8806884ce",
			urff:     "b0e314f5b39a1c71de5dbc86c3e9b22251a6d394",
		},
		"0x3cac5f7909f9cb666cc4d7ef32047b170e454b16": receiverAddressPairString{
			receiver: "0x85b44964bb0d83fa1329dc969d853d710fde339e",
			urff:     "e5780543d87f8b8921e65789ba3c7eb69aba21c7",
		},
		"0x0827d93936df936134dd7b7acaeaea04344b11f2": receiverAddressPairString{
			receiver: "0x5dc1a06fa3717b6084c4e19395ab1651185b6477",
			urff:     "7c4da38909148d56b8e6cc37922e992c2a0a1063",
		},
		"0xa63e936e0eb36c103f665d53bd7ca9c31ec7e1ad": receiverAddressPairString{
			receiver: "0x53372c0fce8ce636ac77cf502c51d5f15868dc64",
			urff:     "4e2c9b2b57fd17a45d28fb4a6d42e932468afaee",
		},
		// temp key
		"0x987181d1385ed1a6fa3a6f645f4f3c9d299a2a12": receiverAddressPairString{
			receiver: "0x0123456789abcdef0123456789abcdef01234567",
			urff:     "89abcdef0123456789abcdef0123456789abcdef",
		},
	}
	PrivilegedAddressesReceivers map[common.Address]ReceiverAddressPair

	ReductionFactors = []struct {
		NSignups *big.Int
		Factor   *big.Int
	}{
		{NSignups: big.NewInt(100000), Factor: big.NewInt(90)},
		{NSignups: big.NewInt(300000), Factor: big.NewInt(80)},
		{NSignups: big.NewInt(1000000), Factor: big.NewInt(70)},
		{NSignups: big.NewInt(3000000), Factor: big.NewInt(60)},
		{NSignups: big.NewInt(10000000), Factor: big.NewInt(50)},
		{NSignups: big.NewInt(30000000), Factor: big.NewInt(40)},
		{NSignups: big.NewInt(100000000), Factor: big.NewInt(30)},
		{NSignups: big.NewInt(300000000), Factor: big.NewInt(20)},
		{NSignups: big.NewInt(1000000000), Factor: big.NewInt(10)},
		{NSignups: big.NewInt(3000000000), Factor: big.NewInt(5)},
		{NSignups: big.NewInt(7500000000), Factor: big.NewInt(0)},
	}
)

type receiverAddressPairString struct{ receiver, urff string }
type ReceiverAddressPair struct{ Receiver, URFF common.Address }

func init() {
	PrivilegedAddressesReceivers = make(map[common.Address]ReceiverAddressPair, len(privSendReceiveAddresses))
	for s, r := range privSendReceiveAddresses {
		PrivilegedAddressesReceivers[common.HexToAddress(s)] = ReceiverAddressPair{
			Receiver: common.HexToAddress(r.receiver),
			URFF:     common.HexToAddress(r.urff),
		}
	}
	fr := ScaledRewards[FullRatio]
	for i := range ReductionFactors {
		fact := ReductionFactors[i].Factor
		sr := rewards{
			URFutureFundFee: scaleReward(fr.URFutureFundFee, fact),
			ManagementFee:   scaleReward(fr.ManagementFee, fact),
			SignupReward:    scaleReward(fr.SignupReward, fact),
			MembersSignupRewards: []*big.Int{
				scaleReward(fr.MembersSignupRewards[0], fact),
				scaleReward(fr.MembersSignupRewards[1], fact),
				scaleReward(fr.MembersSignupRewards[2], fact),
				scaleReward(fr.MembersSignupRewards[3], fact),
				scaleReward(fr.MembersSignupRewards[4], fact),
				scaleReward(fr.MembersSignupRewards[5], fact),
				scaleReward(fr.MembersSignupRewards[6], fact),
			},
			MinerReward: scaleReward(fr.MinerReward, fact),
		}
		sr.TotalSignupRewards = big.NewInt(0)
		for _, i := range sr.MembersSignupRewards {
			sr.TotalSignupRewards.Add(sr.TotalSignupRewards, i)
		}
		sr.Total = new(big.Int).Set(sr.TotalSignupRewards)
		for _, i := range []*big.Int{
			sr.URFutureFundFee,
			sr.SignupReward,
			sr.MinerReward,
		} {
			sr.Total.Add(sr.Total, i)
		}
		ScaledRewards[fact] = sr
	}
}

func scaleReward(v, f *big.Int) *big.Int {
	return new(big.Int).Div(new(big.Int).Mul(v, f), FullRatio)
}

func GetFactor(nSignups *big.Int) *big.Int {
	r := FullRatio
	for _, f := range ReductionFactors {
		if nSignups.Cmp(f.NSignups) > 0 {
			r = f.Factor
		} else {
			break
		}
	}
	return r
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

func isSignupTransaction(msg types.Message) bool {
	return isSignupTx(msg.From(), msg.Value(), msg.Data())
}

func IsPrivilegedAddress(address common.Address) bool {
	_, ok := PrivilegedAddressesReceivers[address]
	return ok
}

var Big10k = new(big.Int).Mul(common.Ether, big.NewInt(10000))

func CalculateTxManagementFee(nSignups, totaWei *big.Int) *big.Int {
	mngFee := ScaledRewards[GetFactor(nSignups)].ManagementFee
	if nSignups.Cmp(common.Big0) == 0 {
		return mngFee
	}
	avg := new(big.Int).Div(totaWei, nSignups)
	if avg.Cmp(Big10k) <= 0 {
		return mngFee
	}
	return common.Big0
}

func calculateBlockTotals(cNSignups, cTotalWei *big.Int, header *types.Header, uncles []*types.Header, msgs []types.Message) (*big.Int, *big.Int) {
	newNSignups := new(big.Int).Set(cNSignups)
	newTotalWei := new(big.Int).Set(cTotalWei)
	blockMngFee := CalculateTxManagementFee(cNSignups, cTotalWei)
	fact := GetFactor(cNSignups)
	rewards := ScaledRewards[fact]
	for _, r := range calculateAccumulatedRewards(header, uncles) {
		newTotalWei.Add(newTotalWei, r)
	}
	for _, m := range msgs {
		if isSignupTransaction(m) {
			newNSignups.Add(newNSignups, common.Big1)
			newTotalWei.Add(newTotalWei, new(big.Int).Add(rewards.Total, blockMngFee))
		}
	}
	return newNSignups, newTotalWei
}

// returns number of sign
func UpdateBlockTotals(parent, header *types.Header, uncles []*types.Header, msgs []types.Message) {
	header.NSignups, header.TotalWei = calculateBlockTotals(parent.NSignups, parent.TotalWei, header, uncles, msgs)
}

func TransactionsToMessages(txs types.Transactions, signer types.Signer) ([]types.Message, error) {
	msgs := make([]types.Message, 0, len(txs))
	for _, t := range txs {
		msg, err := t.AsMessage(signer)
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}
