package core

import (
	"math/big"

	"github.com/urcapital/go-ur/common"
	"github.com/urcapital/go-ur/core/state"
	"github.com/urcapital/go-ur/core/types"
	"github.com/urcapital/go-ur/core/vm"
	"github.com/urcapital/go-ur/crypto"
	"github.com/urcapital/go-ur/logger"
	"github.com/urcapital/go-ur/logger/glog"
)

const (
	BonusMultiplier = 1e+15
	BonusCapUR      = 2000
)

var (
	big8                = big.NewInt(8)
	big32               = big.NewInt(32)
	PrivilegedAddresses = []common.Address{
		common.HexToAddress("0xf27fd61d8f2801a841cabd99e383b0408f28e90e"),
		common.HexToAddress("0x7ca2ce2adf9af2d89ab5450d69518026028cd13e"),
		common.HexToAddress("0xf464a5389d24700c90d6c5d9c7e4ec1b544e7f95"),
		common.HexToAddress("0xf07606fa5f94cc5ee353a554b3d86d77b4b2d949"),
		common.HexToAddress("0x243b92e57f103148c99fb2cea66549467433ed50"),
		common.HexToAddress("0x71dd1c9323d631b3527542225040fe492b23697d"),
		common.HexToAddress("0x71b2a451451f89f8afac236630d4f266f74f0788"),
		common.HexToAddress("0x7df116084ff8d815ded7059dcb777ca5bd23b54e"),
		common.HexToAddress("0xadc912e14aab228a9bd15000269dfd7b1b643a66"),
		common.HexToAddress("0xe05931172771bc2a680c7e14e233eeb2f13a435e"),
	}
)

type StateProcessor struct {
	bc *BlockChain
}

func NewStateProcessor(bc *BlockChain) *StateProcessor {
	return &StateProcessor{bc}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB) (types.Receipts, vm.Logs, *big.Int, error) {
	var (
		receipts     types.Receipts
		totalUsedGas = big.NewInt(0)
		err          error
		header       = block.Header()
		allLogs      vm.Logs
		gp           = new(GasPool).AddGas(block.GasLimit())
	)

	for i, tx := range block.Transactions() {
		statedb.StartRecord(tx.Hash(), block.Hash(), i)
		receipt, logs, _, err := ApplyTransaction(p.bc, gp, statedb, header, tx, totalUsedGas)
		if err != nil {
			return nil, nil, totalUsedGas, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, logs...)
	}
	AccumulateRewards(statedb, header, block.Uncles())
	AccumulateBonuses(statedb, block.Transactions())

	return receipts, allLogs, totalUsedGas, err
}

// ApplyTransaction attemps to apply a transaction to the given state database
// and uses the input parameters for its environment.
//
// ApplyTransactions returns the generated receipts and vm logs during the
// execution of the state transition phase.
func ApplyTransaction(bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *big.Int) (*types.Receipt, vm.Logs, *big.Int, error) {
	_, gas, err := ApplyMessage(NewEnv(statedb, bc, tx, header), tx, gp)
	if err != nil {
		return nil, nil, nil, err
	}

	// Update the state with pending changes
	usedGas.Add(usedGas, gas)
	receipt := types.NewReceipt(statedb.IntermediateRoot().Bytes(), usedGas)
	receipt.TxHash = tx.Hash()
	receipt.GasUsed = new(big.Int).Set(gas)
	if MessageCreatesContract(tx) {
		from, _ := tx.From()
		receipt.ContractAddress = crypto.CreateAddress(from, tx.Nonce())
	}

	logs := statedb.GetLogs(tx.Hash())
	receipt.Logs = logs
	receipt.Bloom = types.CreateBloom(types.Receipts{receipt})

	glog.V(logger.Debug).Infoln(receipt)

	return receipt, logs, gas, err
}

// AccumulateRewards credits the coinbase of the given block with the
// mining reward. The total reward consists of the static block reward
// and rewards for included uncles. The coinbase of each uncle block is
// also rewarded.
func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Header) {
	reward := new(big.Int).Set(BlockReward)
	r := new(big.Int)
	for _, uncle := range uncles {
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, BlockReward)
		r.Div(r, big8)
		statedb.AddBalance(uncle.Coinbase, r)

		r.Div(BlockReward, big32)
		reward.Add(reward, r)
	}
	statedb.AddBalance(header.Coinbase, reward)
}

// AccumulateBonuses credits the receiving address who has received funds
// from a priviliged address at the rate of 1 eth for every 10,000 wei
func AccumulateBonuses(statedb *state.StateDB, transactions types.Transactions) {
	for _, transaction := range transactions {
		from, _ := transaction.From()
		to := transaction.To()
		if isPrivilegedAddress(from) {
			statedb.AddBalance(*to, calculateBonusReward(transaction.Value()))
		}
	}
}

func isPrivilegedAddress(address common.Address) bool {
	for _, privilegedAddress := range PrivilegedAddresses {
		if address == privilegedAddress {
			return true
		}
	}
	return false
}

func calculateBonusReward(transactionValue *big.Int) *big.Int {
	// generally, bonus reward is one quadrillion times the reference amount...
	bonusRewardWei := new(big.Int).Mul(transactionValue, big.NewInt(BonusMultiplier))
	// but is capped at 2000 UR
	bonusRewardCapWei := new(big.Int).Mul(big.NewInt(BonusCapUR), common.Ether)

	return common.BigMin(bonusRewardWei, bonusRewardCapWei)
}
