package core

import (
	"math/big"
	"fmt"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/state"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/core/vm"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/logger"
	"github.com/ethereum/go-ethereum/logger/glog"

)

var (
	big8  = big.NewInt(8)
	big32 = big.NewInt(32)
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
func AccumulateRewards(statedb *state.StateDB, header *types.Header, uncles []*types.Head{

	// create bonus rewards for any reference transactions that are part of this block
	privilegedAddresses := []common.Address{ common.HexToAddress("0x8805317929d0a8cd1e7a19a4a2523b821ed05e42") }
	for _, transaction := range transactions {
		from, _ := transaction.From()
		fmt.Printf("Transaction - from: #%s\n", from)

		for i := 0; i < len(privilegedAddresses); i++ {
			fmt.Printf("privilegedAddress - i: #%d\n", i)

			if from == privilegedAddresses[i] {

				fmt.Printf("a match!\n")

				// this is a 'bonus reference transaction', so let's use it to create a bonus reward
				bonusReward := new(big.Int).Mul(transaction.Value(), big.NewInt(1e+15)) // generally, bonus reward is one quadrillion times the reference amount...
				bonusRewardCap := new(big.Int).Mul(big.NewInt(2000), big.NewInt(1e+18)) // but is capped at 2000 UR
				bonusReward = common.BigMin(bonusReward, bonusRewardCap)
				statedb.AddBalance(*transaction.To(), bonusReward)
				fmt.Printf("did it!\n")
				break

			}
		}
	}

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
