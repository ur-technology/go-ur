package core

import (
	"fmt"
	"math/big"
	"math/rand"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/urcapital/go-ur/common"
	"github.com/urcapital/go-ur/core/types"
	"github.com/urcapital/go-ur/crypto"
	"github.com/urcapital/go-ur/ethdb"
	"github.com/urcapital/go-ur/event"
	"github.com/urcapital/go-ur/params"
)

func Test_ItDoesntApplyBonusesForNonQualifyingTransactions(t *testing.T) {
	transactionValue := big.NewInt(1000)
	randomSeed := time.Now().UnixNano()
	rand.Seed(randomSeed)

	for n, i := 1000, 0; i <= n; i++ {
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

		bchain, err := NewBlockChain(gendb, FakePow{}, &event.TypeMux{})
		assert.NoError(t, err)
		bchain.ResetWithGenesisBlock(genesis)

		_, err = bchain.InsertChain(types.Blocks(blocks))
		assert.NoError(t, err)

		statedb, err := bchain.State()
		assert.NoError(t, err)

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
	smallTransactionValue := big.NewInt(1000)
	largeTransactionValue := new(big.Int).Mul(big.NewInt(2500), common.Ether)

	var (
		gendb, _             = ethdb.NewMemDatabase()
		privKey, _           = crypto.HexToECDSA(RandHex(64))
		privAddress          = crypto.PubkeyToAddress(privKey.PublicKey)
		funds                = new(big.Int).Mul(big.NewInt(2501), common.Ether)
		largeKey, _          = crypto.HexToECDSA(RandHex(64))
		largeTransactionAddr = crypto.PubkeyToAddress(largeKey.PublicKey)
		smallKey, _          = crypto.HexToECDSA(RandHex(64))
		smallTransactionAddr = crypto.PubkeyToAddress(smallKey.PublicKey)
		genesis              = GenesisBlockForTesting(gendb, privAddress, funds)
	)

	PrivilegedAddresses = append(PrivilegedAddresses, privAddress)

	blocks, _ := GenerateChain(genesis, gendb, 1, func(i int, block *BlockGen) {
		block.SetCoinbase(common.Address{0x00})
		tx, err := types.NewTransaction(block.TxNonce(privAddress), smallTransactionAddr, smallTransactionValue, params.TxGas, nil, nil).SignECDSA(privKey)
		if err != nil {
			panic(err)
		}
		block.AddTx(tx)

		tx, err = types.NewTransaction(block.TxNonce(privAddress), largeTransactionAddr, largeTransactionValue, params.TxGas, nil, nil).SignECDSA(privKey)
		if err != nil {
			panic(err)
		}
		block.AddTx(tx)
	})

	bchain, err := NewBlockChain(gendb, FakePow{}, &event.TypeMux{})
	assert.NoError(t, err)
	bchain.ResetWithGenesisBlock(genesis)

	_, err = bchain.InsertChain(types.Blocks(blocks))
	assert.NoError(t, err)

	statedb, err := bchain.State()
	assert.NoError(t, err)

	// transaction amount 1000 + bonus 1e+19
	assert.Equal(t, big.NewInt(1000000000000001000), statedb.GetBalance(smallTransactionAddr))
	// transaction amount 2.5e+22 + bonus 2e+22 = 4.5e+22
	assert.Equal(t, new(big.Int).Mul(big.NewInt(4500), common.Ether), statedb.GetBalance(largeTransactionAddr))

	// small transaction 1000 - large transaction 2.5e+22
	assert.Equal(t, big.NewInt(999999999999999000), statedb.GetBalance(privAddress))
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
