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

var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
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
	AccumulateRewards(statedb, header, block.Uncles())

	return receipts, allLogs, totalUsedGas, err
}

// ApplyTransaction attempts to apply a transaction to the given state database
// and uses the input parameters for its environment.
//
// ApplyTransactions returns the generated receipts and vm logs during the
// execution of the state transition phase.
func ApplyTransaction(config *ChainConfig, bc *BlockChain, gp *GasPool, statedb *state.StateDB, header *types.Header, tx *types.Transaction, usedGas *big.Int, cfg vm.Config) (*types.Receipt, vm.Logs, *big.Int, error) {
	// check for a signup transaction
	if isSignupTransaction(tx) {
		if signupChain, err := getSignupChain(bc, tx.Data()); err == nil {
			// pay the miner BlockReward for every signup
			statedb.AddBalance(header.Coinbase, BlockReward)
			// pay the member being signed up
			statedb.AddBalance(*tx.To(), SignupReward)
			// pay the referral members
			remRewards := TotalSingupRewards
			for i, m := range signupChain {
				statedb.AddBalance(m, MembersSingupRewards[i])
				remRewards = new(big.Int).Sub(remRewards, MembersSingupRewards[i])
			}
			txFrom, _ := tx.From()
			recvAddr := PrivilegedAddressesReceivers[txFrom]
			// pay 5000 UR to the UR Future Fund
			statedb.AddBalance(recvAddr.URFF, URFutureFundFee)
			// pay the receiver address any remaining fees from the members and the management fee
			pBlock := bc.GetBlock(header.ParentHash)
			mngFee := calculateTxManagementFee(pBlock.NSignups(), pBlock.TotalWei())
			statedb.AddBalance(PrivilegedAddressesReceivers[txFrom].Receiver, new(big.Int).Add(mngFee, remRewards))
		}
	}

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

func calculateAccumulatedRewards(header *types.Header, uncles []*types.Header) map[common.Address]*big.Int {
	rew := make(map[common.Address]*big.Int, len(uncles)+1)
	reward := new(big.Int).Set(BlockReward)
	r := new(big.Int)
	for _, uncle := range uncles {
		// the miner for the uncle block receives
		// ((uncleBlockNumber + 8 - currentBlockNumber) * BlockReward) / 8
		r.Add(uncle.Number, big8)
		r.Sub(r, header.Number)
		r.Mul(r, BlockReward)
		r.Div(r, big8)
		ub, ok := rew[uncle.Coinbase]
		if !ok {
			ub = big.NewInt(0)
		}
		rew[uncle.Coinbase] = ub.Add(ub, r)

		// the miner receives 1/32 * BlockReward for every uncle block
		r.Div(BlockReward, big32)
		reward.Add(reward, r)
	}
	ub, ok := rew[header.Coinbase]
	if !ok {
		ub = big.NewInt(0)
	}
	rew[header.Coinbase] = ub.Add(ub, reward)
	return rew
}

// AccumulateRewards credits the coinbase of the given block with the
// mining reward. The total reward consists of the static block reward
// and rewards for included uncles. The coinbase of each uncle block is
// also rewarded.
func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Header) {
	rewards := calculateAccumulatedRewards(header, uncles)
	for a, r := range rewards {
		statedb.AddBalance(a, r)
	}
}

// func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Header) {
// 	reward := new(big.Int).Set(BlockReward)
// 	r := new(big.Int)
// 	for _, uncle := range uncles {
// 		r.Add(uncle.Number, big8)
// 		r.Sub(r, header.Number)
// 		r.Mul(r, BlockReward)
// 		r.Div(r, big8)
// 		statedb.AddBalance(uncle.Coinbase, r)

// 		r.Div(BlockReward, big32)
// 		reward.Add(reward, r)
// 	}

// 	statedb.AddBalance(header.Coinbase, reward)
// }
