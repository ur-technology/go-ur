package core

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/ur-technology/go-ur/common"
	"github.com/ur-technology/go-ur/core/state"
	"github.com/ur-technology/go-ur/core/types"
	"github.com/ur-technology/go-ur/crypto"
	"github.com/ur-technology/go-ur/ethdb"
	"github.com/ur-technology/go-ur/event"
	"github.com/ur-technology/go-ur/params"
)

func Test_ItDoesntApplyBonusesForNonQualifyingTransactions(t *testing.T) {
	transactionValue := big.NewInt(1000)
	randomSeed := time.Now().UnixNano()
	rand.Seed(randomSeed)

	for n, i := 100, 0; i <= n; i++ {
		var (
			gendb, _  = ethdb.NewMemDatabase()
			key, _    = crypto.HexToECDSA(RandHex(64))
			address   = crypto.PubkeyToAddress(key.PublicKey)
			funds     = big.NewInt(1000000000)
			toKey, _  = crypto.HexToECDSA(RandHex(64))
			toAddress = crypto.PubkeyToAddress(toKey.PublicKey)
			genesis   = GenesisBlockForTesting(gendb, address, funds)
		)

		var hasCollided bool
		for _, privilegedAddress := range PrivilegedAddresses {
			if privilegedAddress.Hex() == address.Hex() {
				hasCollided = true
			}
		}
		if hasCollided {
			continue
		}

		blocks, _ := GenerateChain(genesis, gendb, 1, func(i int, block *BlockGen) {
			block.SetCoinbase(common.Address{0x00})
			// If the block number is multiple of 3, send a few bonus transactions to the miner
			tx, err := types.NewTransaction(block.TxNonce(address), toAddress, transactionValue, params.TxGas, nil, nil).SignECDSA(key)
			if err != nil {
				panic(err)
			}
			block.AddTx(tx)
		})

		statedb := buildBlockChain(t, gendb, genesis, blocks)

		expectedBalance := transactionValue
		assert.False(
			t,
			statedb.GetBalance(toAddress).Cmp(expectedBalance) == 1,
			fmt.Sprintf(
				"Wallet balance larger than expected, wanted '%s' got '%s'. Random seed: %d\n",
				expectedBalance,
				statedb.GetBalance(toAddress),
				randomSeed,
			),
		)
	}
}

func Test_ItAppliesBonusesForQualifyingTransactions(t *testing.T) {
	tests := []struct {
		TransactionValue        *big.Int
		ExpectedReceiverBalance *big.Int
	}{
		{
			TransactionValue:        big.NewInt(1),
			ExpectedReceiverBalance: big.NewInt(1000000000000000),
		},
		{
			TransactionValue:        big.NewInt(20),
			ExpectedReceiverBalance: big.NewInt(20000000000000000),
		},
		{
			TransactionValue:        big.NewInt(300),
			ExpectedReceiverBalance: big.NewInt(300000000000000000),
		},
		{
			TransactionValue:        big.NewInt(4000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(4), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(50000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(50), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(600000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(600), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(7000000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(80000000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(900000000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(1000000000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(20000000000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(300000000000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        big.NewInt(4000000000000),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        new(big.Int).Mul(big.NewInt(1), common.Ether),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        new(big.Int).Mul(big.NewInt(20), common.Ether),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        new(big.Int).Mul(big.NewInt(300), common.Ether),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
		{
			TransactionValue:        new(big.Int).Mul(big.NewInt(4000), common.Ether),
			ExpectedReceiverBalance: new(big.Int).Mul(big.NewInt(2000), common.Ether),
		},
	}

	for _, test := range tests {

		var (
			funds           = new(big.Int).Mul(big.NewInt(10000), common.Ether)
			key, _          = crypto.HexToECDSA(RandHex(64))
			transactionAddr = crypto.PubkeyToAddress(key.PublicKey)
		)
		privKey, privAddress := setupPrivilegedAddress(t)
		genesis, gendb := setupGenesis(t, privAddress, funds)

		blocks, _ := GenerateChain(genesis, gendb, 1, func(i int, block *BlockGen) {
			block.SetCoinbase(common.Address{0x00})

			tx, err := types.NewTransaction(block.TxNonce(privAddress), transactionAddr, test.TransactionValue, params.TxGas, nil, nil).SignECDSA(privKey)
			if err != nil {
				panic(err)
			}
			block.AddTx(tx)
		})

		statedb := buildBlockChain(t, gendb, genesis, blocks)

		assert.Equal(t, test.ExpectedReceiverBalance, statedb.GetBalance(transactionAddr))

		expectedPrivAddressBalance := new(big.Int).Sub(funds, test.TransactionValue)
		assert.Equal(t, expectedPrivAddressBalance, statedb.GetBalance(privAddress))
	}
}

func Test_ItAppliesMinerRewardBonusForNewSignupsInBlock(t *testing.T) {
	privKey, privAddress := setupPrivilegedAddress(t)

	tests := []struct {
		Description              string
		NumberOfSignups          int
		NumberOfBlocks           int
		AdditionalTransactionsFn func(int, *BlockGen, []common.Address)
	}{
		{
			Description:     "No signups",
			NumberOfSignups: 0,
			NumberOfBlocks:  1,
		},
		{
			Description:     "1 Signup",
			NumberOfSignups: 1,
			NumberOfBlocks:  1,
		},
		{
			Description:     "2 signups over 2 blocks",
			NumberOfSignups: 2,
			NumberOfBlocks:  2,
		},
		{
			Description:     "5 signups over 2 blocks",
			NumberOfSignups: 5,
			NumberOfBlocks:  2,
		},
		{
			Description:     "700 signups over 200 blocks",
			NumberOfSignups: 700,
			NumberOfBlocks:  200,
		},
		{
			Description:     "30 signups over 5 blocks, with non qualifying signup transactions",
			NumberOfSignups: 30,
			NumberOfBlocks:  5,
			AdditionalTransactionsFn: func(i int, block *BlockGen, nonSignupAddresses []common.Address) {
				tx, err := types.NewTransaction(block.TxNonce(privAddress), nonSignupAddresses[i], big.NewInt(int64(rand.Intn(1999999))), params.TxGas, nil, nil).SignECDSA(privKey)
				if err != nil {
					panic(err)
				}
				block.AddTx(tx)
			},
		},
	}

	transactionValue := big.NewInt(2000000)
	funds := new(big.Int).Mul(big.NewInt(1), common.Ether)

	for _, test := range tests {
		expectedBonusReward := new(big.Int).Mul(BlockReward, big.NewInt(int64(test.NumberOfSignups)))
		expectedBlockReward := new(big.Int).Mul(BlockReward, big.NewInt(int64(test.NumberOfBlocks)))
		genesis, gendb := setupGenesis(t, privAddress, funds)

		newAddresses := generateNewAddresses(t, test.NumberOfSignups)
		nonSignupAddresses := generateNewAddresses(t, test.NumberOfSignups)

		blocks, _ := GenerateChain(genesis, gendb, test.NumberOfBlocks, func(i int, block *BlockGen) {
			block.SetCoinbase(common.Address{0x00})

			if test.AdditionalTransactionsFn != nil {
				test.AdditionalTransactionsFn(i, block, nonSignupAddresses)
			}

			if test.NumberOfBlocks < test.NumberOfSignups {
				// Distribute new signup transactions across blocks
				for j := (i * int(test.NumberOfSignups/test.NumberOfBlocks)); j < ((i + 1) * int(test.NumberOfSignups/test.NumberOfBlocks)); j++ {
					tx, err := types.NewTransaction(block.TxNonce(privAddress), newAddresses[j], transactionValue, params.TxGas, nil, nil).SignECDSA(privKey)
					if err != nil {
						panic(err)
					}
					block.AddTx(tx)
				}
				// On last block
				if test.NumberOfBlocks-1 == i {
					// Do any remaining transactions for new signups
					for j := (test.NumberOfSignups - (test.NumberOfSignups % test.NumberOfBlocks)); j < test.NumberOfSignups; j++ {
						tx, err := types.NewTransaction(block.TxNonce(privAddress), newAddresses[j], transactionValue, params.TxGas, nil, nil).SignECDSA(privKey)
						if err != nil {
							panic(err)
						}
						block.AddTx(tx)
					}
				}
			} else if i < test.NumberOfSignups {
				// 1 transaction per block
				tx, err := types.NewTransaction(block.TxNonce(privAddress), newAddresses[i], transactionValue, params.TxGas, nil, nil).SignECDSA(privKey)
				if err != nil {
					panic(err)
				}
				block.AddTx(tx)
			}
		})

		statedb := buildBlockChain(t, gendb, genesis, blocks)

		expectedMinerBalance := new(big.Int).Add(expectedBlockReward, expectedBonusReward)
		assert.Equal(t, expectedMinerBalance, statedb.GetBalance(common.Address{0x00}), test.Description)
	}
}

// http://stackoverflow.com/a/31832326
func RandHex(n int) string {
	const letterBytes = "0123456789abcdef"

	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func setupGenesis(t *testing.T, privAddress common.Address, funds *big.Int) (*types.Block, *ethdb.MemDatabase) {
	gendb, err := ethdb.NewMemDatabase()
	assert.NoError(t, err)

	return GenesisBlockForTesting(gendb, privAddress, funds), gendb
}
func setupPrivilegedAddress(t *testing.T) (*ecdsa.PrivateKey, common.Address) {
	privKey, err := crypto.HexToECDSA(RandHex(64))
	assert.NoError(t, err)

	privAddress := crypto.PubkeyToAddress(privKey.PublicKey)
	PrivilegedAddresses = append(PrivilegedAddresses, privAddress)

	return privKey, privAddress
}
func generateNewAddresses(t *testing.T, n int) []common.Address {
	newAddresses := make([]common.Address, n)
	for i := 0; i < n; i++ {
		key, err := crypto.HexToECDSA(RandHex(64))
		assert.NoError(t, err)

		newAddresses[i] = crypto.PubkeyToAddress(key.PublicKey)
	}

	return newAddresses
}
func buildBlockChain(t *testing.T, gendb *ethdb.MemDatabase, genesis *types.Block, blocks types.Blocks) *state.StateDB {
	bchain, err := NewBlockChain(gendb, FakePow{}, &event.TypeMux{})
	assert.NoError(t, err)
	bchain.ResetWithGenesisBlock(genesis)

	_, err = bchain.InsertChain(types.Blocks(blocks))
	assert.NoError(t, err)

	statedb, err := bchain.State()
	assert.NoError(t, err)

	return statedb
}
