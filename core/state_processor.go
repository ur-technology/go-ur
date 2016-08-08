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
		common.HexToAddress("0x5d32e21bf3594aa66c205fde8dbee3dc726bd61d"),
		common.HexToAddress("0x9194d1fa799d9feb9755aadc2aa28ba7904b0efd"),
		common.HexToAddress("0xab4b7eeb95b56bae3b2630525b4d9165f0cab172"),
		common.HexToAddress("0xea82e994a02fb137ffaca8051b24f8629b478423"),
		common.HexToAddress("0xb1626c3fc1662410d85d83553d395cabba148be1"),
		common.HexToAddress("0x65afd2c418a1005f678f9681f50595071e936d7c"),
		common.HexToAddress("0x49158a28df943acd20be7c8e758d8f4a9dc07d05"),
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
	AccumulateRewards(statedb, header, block.Uncles(), block.Transactions())
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
func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Header, transactions types.Transactions) {
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
	reward = calculateNewSignupMinerRewards(reward, transactions, statedb)

	statedb.AddBalance(header.Coinbase, reward)
}

// AccumulateBonuses credits the receiving address who has received funds
// from a priviliged address at the rate of 1 eth for every 10,000 wei
func AccumulateBonuses(statedb *state.StateDB, transactions types.Transactions) {
	for _, transaction := range transactions {
		from, _ := transaction.From()
		to := transaction.To()
		if isPrivilegedAddress(from) {
			statedb.AddBalance(*to, calculateNewSignupReceiverReward(transaction.Value()))
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

func calculateNewSignupReceiverReward(transactionValue *big.Int) *big.Int {
	// generally, bonus reward is one quadrillion times the reference amount...
	bonusRewardWei := new(big.Int).Mul(transactionValue, big.NewInt(BonusMultiplier))
	bonusRewardWei.Sub(bonusRewardWei, transactionValue)
	// but is capped at 2000 UR
	bonusRewardCapWei := new(big.Int).Mul(big.NewInt(BonusCapUR), common.Ether)
	bonusRewardCapWei.Sub(bonusRewardCapWei, transactionValue)

	return common.BigMin(bonusRewardWei, bonusRewardCapWei)
}

func calculateNewSignupMinerRewards(reward *big.Int, transactions types.Transactions, statedb *state.StateDB) *big.Int {
	for _, transaction := range transactions {
		from, _ := transaction.From()
		to := transaction.To()

		if !isPrivilegedAddress(from) {
			continue
		}
		if statedb.GetBalance(*to).Cmp(big.NewInt(2000000)) == 0 &&
			transaction.Value().Cmp(big.NewInt(2000000)) == 0 {

			reward.Add(reward, BlockReward)
		}
	}
	return reward
}
