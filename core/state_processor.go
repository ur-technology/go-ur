// Copyright 2015 The go-ethereum Authors
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
	"math/big"

	"github.com/ur-technology/go-ur/common"
	"github.com/ur-technology/go-ur/core/state"
	"github.com/ur-technology/go-ur/core/types"
	"github.com/ur-technology/go-ur/core/vm"
	"github.com/ur-technology/go-ur/crypto"
	"github.com/ur-technology/go-ur/logger"
	"github.com/ur-technology/go-ur/logger/glog"
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

// StateProcessor is a basic Processor, which takes care of transitioning
// state from one point to another.
//
// StateProcessor implements Processor.
type StateProcessor struct {
	config *ChainConfig
	bc     *BlockChain
}

// NewStateProcessor initialises a new StateProcessor.
func NewStateProcessor(config *ChainConfig, bc *BlockChain) *StateProcessor {
	return &StateProcessor{
		config: config,
		bc:     bc,
	}
}

// Process processes the state changes according to the Ethereum rules by running
// the transaction messages using the statedb and applying any rewards to both
// the processor (coinbase) and any included uncles.
//
// Process returns the receipts and logs accumulated during the process and
// returns the amount of gas that was used in the process. If any of the
// transactions failed to execute due to insufficient gas it will return an error.
func (p *StateProcessor) Process(block *types.Block, statedb *state.StateDB, cfg vm.Config) (types.Receipts, vm.Logs, *big.Int, error) {
	var (
		receipts     types.Receipts
		totalUsedGas = big.NewInt(0)
		err          error
		header       = block.Header()
		allLogs      vm.Logs
		gp           = new(GasPool).AddGas(block.GasLimit())
	)
	// Mutate the the block and state according to any hard-fork specs
	if p.config.DAOForkSupport && p.config.DAOForkBlock != nil && p.config.DAOForkBlock.Cmp(block.Number()) == 0 {
		ApplyDAOHardFork(statedb)
	}
	// Iterate over and process the individual transactions
	for i, tx := range block.Transactions() {
		statedb.StartRecord(tx.Hash(), block.Hash(), i)
		receipt, logs, _, err := ApplyTransaction(p.config, p.bc, gp, statedb, header, tx, totalUsedGas, cfg)
		if err != nil {
			return nil, nil, totalUsedGas, err
		}
		receipts = append(receipts, receipt)
		allLogs = append(allLogs, logs...)
	}
	AccumulateRewards(statedb, header, block.Uncles(), block.Transactions())

	return receipts, allLogs, totalUsedGas, err
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment.
//
// ApplyTransactions returns the generated receipts and vm logs during the
// execution of the state transition phase.
func ApplyTransaction(config *ChainConfig, bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *big.Int, cfg vm.Config) (*types.Receipt, vm.Logs, *big.Int, error) {
	_, gas, err := ApplyMessage(NewEnv(statedb, config, bc, tx, header, cfg), tx, gp)
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
