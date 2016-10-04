package core_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"

	"github.com/urcapital/go-ur/common"
	"github.com/urcapital/go-ur/core"
	"github.com/urcapital/go-ur/core/types"
	"github.com/urcapital/go-ur/crypto"
	"github.com/urcapital/go-ur/ethdb"
	"github.com/urcapital/go-ur/event"
	"github.com/urcapital/go-ur/params"
)

var defaultChainConfig = core.MakeChainConfig()

type Simulator struct {
	account    core.GenesisAccount
	db         ethdb.Database
	pendingTxs []*TxData
	BlockChain *core.BlockChain
	Coinbase   common.Address
}

func newSimulator(account core.GenesisAccount) (*ethdb.MemDatabase, *core.BlockChain, error) {
	db, err := ethdb.NewMemDatabase()
	if err != nil {
		return nil, nil, err
	}
	core.WriteGenesisBlockForTesting(db, account)
	blockchain, err := core.NewBlockChain(db, defaultChainConfig, &core.FakePow{}, &event.TypeMux{})
	if err != nil {
		return nil, nil, err
	}
	return db, blockchain, nil
}

func NewSimulator(account core.GenesisAccount) (*Simulator, error) {
	db, bc, err := newSimulator(account)
	if err != nil {
		return nil, err
	}
	return &Simulator{
		db:         db,
		BlockChain: bc,
		account:    account,
	}, nil
}

// Commits pending transactions and return the a slice of []TxData with the Tx field set.
func (b *Simulator) Commit() (commitedTxs []*TxData, err error) {
	defer func() {
		p := recover()
		if p == nil {
			return
		}
		if e, ok := p.(error); ok && e != nil {
			err = e
		} else {
			panic(p)
		}
	}()
	blocks, _ := core.GenerateChain(defaultChainConfig, b.BlockChain, b.BlockChain.CurrentBlock(), b.db, 1, func(n int, block *core.BlockGen) {
		block.SetCoinbase(b.Coinbase)
		for _, stx := range b.pendingTxs {
			tx, err := sendTx(block, stx)
			if err != nil {
				panic(fmt.Errorf("failed at block %d: %s", b.BlockChain.CurrentBlock().Number(), err))
			}
			stx.Tx = tx
		}
	})
	if _, err = b.BlockChain.InsertChain(blocks); err != nil {
		return
	}
	commitedTxs = b.pendingTxs
	b.pendingTxs = []*TxData{}
	return
}

// Rollback all blocks and all pending transactions.
func (b *Simulator) RollbackBlockChain() error {
	db, bc, err := newSimulator(b.account)
	if err != nil {
		return err
	}
	b.db = db
	b.BlockChain = bc
	b.pendingTxs = nil
	b.Coinbase = common.Address{}
	return nil
}

// Rollback pending transactions.
func (b *Simulator) RollbackPendingTxs() { b.pendingTxs = nil }

// AddPendingTx adds a transaction to the pending list.
func (b *Simulator) AddPendingTx(tx *TxData) {
	b.pendingTxs = append(b.pendingTxs, tx)
}

// TxData holds transaction data.
type TxData struct {
	From  *ecdsa.PrivateKey
	To    common.Address
	Value *big.Int
	Data  []byte
	Tx    *types.Transaction
}

func (t *TxData) String() string {
	f := crypto.PubkeyToAddress(t.From.PublicKey)
	return fmt.Sprintf(
		"From: %s\tTo: %s\tValue: %s\tData: %s",
		hex.EncodeToString(f[:]),
		hex.EncodeToString(t.To[:]),
		t.Value.String(),
		hex.EncodeToString(t.Data),
	)
}

func sendTx(bg *core.BlockGen, simTx *TxData) (*types.Transaction, error) {
	nonce := bg.TxNonce(crypto.PubkeyToAddress(simTx.From.PublicKey))
	tx, err := types.NewTransaction(nonce, simTx.To, simTx.Value, new(big.Int).Mul(params.TxGas, big.NewInt(100)), nil, simTx.Data).SignECDSA(simTx.From)
	if err != nil {
		return nil, err
	}
	bg.AddTx(tx)
	return tx, nil
}
