package eth

import (
	"fmt"
	"math/big"
	"time"

	"github.com/ur-technology/go-ur/common"
	"github.com/ur-technology/go-ur/core"
	"github.com/ur-technology/go-ur/core/types"
	"github.com/ur-technology/go-ur/params"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var colNames = []string{"blocks", "transactions", "balance_changes", "links"}

// bulk insert
type bulk struct {
	b *mgo.Bulk
	c uint64
}

// db block importer
type dbDumper struct {
	session *mgo.Session
	dbName  string
	nDocs   uint64
	bulk    map[string]*bulk
}

type dbBlock struct {
	Number       uint64   `bson:"_id"`
	Hash         string   `bson:"hash"`
	Timestamp    uint64   `bson:"timestamp"`
	Miner        string   `bson:"miner"`
	Transactions []string `bson:"transactions"`
	Difficulty   string   `bson:"difficulty"`
	GasLimit     uint64   `bson:"gasLimit"`
	GasUsed      uint64   `bson:"gasUsed"`
	ExtraData    []byte   `bson:"extraData"`
	NSignups     uint64   `bson:"nSignups"`
	TotalWei     string   `bson:"totalWei"`
}

type dbTransaction struct {
	Hash        string `bson:"_id"`
	BlockNumber uint64 `bson:"blockNumber"`
	From        string `bson:"from"`
	To          string `bson:"to"`
	Gas         uint64 `bson:"gas"`
	GasPrice    uint64 `bson:"gasPrice"`
	Input       []byte `bson:"input"`
	Value       string `bson:"value"`
	IsSignup    bool   `bson:"isSignup,omitempty"`
}

type balChange struct {
	To        string `bson:"to"`
	Type      string `bson:"type"`
	Value     string `bson:"value"`
	From      string `bson:"from,omitempty"`
	TxHash    string `bson:"tx,omitempty"`
	Block     uint64 `bson:"block,omitempty"`
	Processed bool   `bson:"processed"`
}

type addrLink struct {
	Type   string `bson:"type"`
	To     string `bson:"to"`
	From   string `bson:"from"`
	Value  string `bson:"value,omitempty"`
	TxHash string `bson:"tx"`
	Block  uint64 `bson:"block"`
}

func newDBDumper(connStr string, nDocs uint64) (*dbDumper, error) {
	if connStr == "" {
		return nil, fmt.Errorf("$UR_CHAIN_DB is empty. Not exporting.")
	}
	di, err := mgo.ParseURL(connStr)
	if err != nil {
		return nil, err
	}
	s, err := mgo.DialWithInfo(di)
	if err != nil {
		return nil, err
	}
	db := s.DB(di.Database)
	return &dbDumper{
		nDocs:   nDocs,
		session: s,
		dbName:  di.Database,
		bulk: map[string]*bulk{
			"blocks":          &bulk{db.C("blocks").Bulk(), 0},
			"transactions":    &bulk{db.C("transactions").Bulk(), 0},
			"balance_changes": &bulk{db.C("balance_changes").Bulk(), 0},
			"links":           &bulk{db.C("links").Bulk(), 0},
		},
	}, nil
}

func (dbd *dbDumper) dumpBlocks(eth *Ethereum) {
	// get last block processed
	b, err := dbd.lastBlock()
	if err != nil {
		fmt.Println("Block Importer: can't fetch status:", err)
		return
	}
	blk := uint64(b)
	// process blocks and wait if necessary
	for {
		if blk == eth.blockchain.CurrentBlock().NumberU64() {
			for n := range dbd.bulk {
				dbd.flushAndReset(n)
			}
			// wait 1s
			<-time.After(time.Second)
			continue
		}
		blk++
		if err = dbd.importBlock(eth.chainConfig, eth.blockchain, blk); err != nil {
			fmt.Println("Block Importer: database error:", err)
		}
	}
}

// flush any data in the bulk insert bufers
func (dbd *dbDumper) close() error {
	defer dbd.session.Close()
	for i := range dbd.bulk {
		if err := dbd.flush(i); err != nil {
			return err
		}
	}
	return nil
}

func (dbd *dbDumper) lastBlock() (uint64, error) {
	b := &dbBlock{}
	err := dbd.session.DB(dbd.dbName).C("blocks").Find(bson.M{}).Sort("-_id").One(b)
	if err != nil {
		return 0, err
	}
	return b.Number, nil
}

func (dbd *dbDumper) importBlock(cfg *params.ChainConfig, bc *core.BlockChain, blkNr uint64) error {
	blk := bc.GetBlockByNumber(blkNr)
	// block
	b := &dbBlock{
		Number:     blk.Number().Uint64(),
		Hash:       blk.Hash().Hex(),
		Timestamp:  blk.Time().Uint64(),
		Miner:      blk.Coinbase().Hex(),
		Difficulty: blk.Difficulty().String(),
		GasLimit:   blk.GasLimit().Uint64(),
		GasUsed:    blk.GasUsed().Uint64(),
		ExtraData:  blk.Extra(),
		NSignups:   blk.NSignups().Uint64(),
		TotalWei:   blk.TotalWei().String(),
	}
	// transactions
	tt := blk.Transactions()
	// account for balance changes
	balChanges := make([]interface{}, 0, len(tt)+10)
	// mining fees for coinbase and uncles coinbase
	miningFees := core.CalculateAccumulatedRewards(blk.Header(), blk.Uncles())
	// convert txs to msgs
	msgs, err := core.TransactionsToMessages(tt, types.MakeSigner(cfg, blk.Number()))
	if err != nil {
		return err
	}
	// transactions gas
	cb := blk.Coinbase()
	// add miner fees to balance changes
	cbFee := new(big.Int).Add(miningFees[cb], blk.GasUsed())
	delete(miningFees, cb)
	balChanges = append(balChanges, balChange{
		To:    cb.Hex(),
		Type:  "coinbase",
		Value: cbFee.String(),
		Block: blk.NumberU64(),
	})
	// add uncle blocks fees
	for a, v := range miningFees {
		balChanges = append(balChanges, balChange{
			To:    a.Hex(),
			Type:  "uncle_coinbase",
			Value: v.String(),
			Block: blk.NumberU64(),
		})
	}
	// links
	links := make([]interface{}, 0, len(msgs))
	// transactions
	txs := make([]interface{}, 0, len(msgs))
	for i, m := range msgs {
		// save hash in the block
		b.Transactions = append(b.Transactions, tt[i].Hash().Hex())
		// is this a signup tx
		inp := m.Data()
		isSignup := core.IsPrivilegedAddress(m.From()) &&
			m.Value().Cmp(common.Big1) == 0 &&
			(len(inp) == 1 || len(inp) == 41) &&
			inp[0] == 1
		// add tx to txs
		txs = append(txs, dbTransaction{
			Hash:        tt[i].Hash().Hex(),
			BlockNumber: blk.NumberU64(),
			From:        m.From().Hex(),
			To:          m.To().Hex(),
			Gas:         m.Gas().Uint64(),
			GasPrice:    m.GasPrice().Uint64(),
			Input:       inp,
			Value:       m.Value().String(),
			IsSignup:    isSignup,
		})
		if !isSignup {
			// not a signup
			balChanges = append(balChanges, balChange{
				To:     m.To().Hex(),
				Type:   "tx",
				Value:  m.Value().String(),
				From:   m.From().Hex(),
				TxHash: tt[i].Hash().Hex(),
				Block:  blk.NumberU64(),
			})
			links = append(links, addrLink{
				Type:   "tx",
				From:   m.From().Hex(),
				To:     m.To().Hex(),
				TxHash: tt[i].Hash().Hex(),
				Block:  blk.NumberU64(),
				Value:  m.Value().String(),
			})
		} else {
			// signup
			sigChain, err := core.SignupChain(bc, tt[i])
			if err != nil {
				return err
			}
			var ref common.Address
			if len(sigChain) != 0 {
				ref = sigChain[0]
			}
			links = append(links, addrLink{
				Type:   "signup",
				From:   ref.Hex(),
				To:     m.To().Hex(),
				TxHash: tt[i].Hash().Hex(),
				Block:  blk.NumberU64(),
			})
			// the miner receives core.BlockReward
			balChanges = append(balChanges, balChange{
				To:    cb.Hex(),
				Value: core.BlockReward.String(),
				Type:  "signup_coinbase",
				Block: blk.NumberU64(),
			})
			// the signup address receives core.SignupReward
			balChanges = append(balChanges, balChange{
				To:     cb.Hex(),
				Value:  core.SignupReward.String(),
				Type:   "signup_1",
				Block:  blk.NumberU64(),
				TxHash: tt[i].Hash().Hex(),
			})
			// signup chain
			remRewards := core.TotalSingupRewards
			for ii, m := range sigChain {
				balChanges = append(balChanges, balChange{
					To:     m.Hex(),
					Value:  core.MembersSingupRewards[ii].String(),
					Type:   fmt.Sprintf("signup_%d", ii+2),
					Block:  blk.NumberU64(),
					TxHash: tt[i].Hash().Hex(),
				})
				remRewards = new(big.Int).Sub(remRewards, core.MembersSingupRewards[ii])
			}
			// UR Future Fund
			balChanges = append(balChanges, balChange{
				To:     core.PrivilegedAddressesReceivers[m.From()].URFF.Hex(),
				Value:  core.URFutureFundFee.String(),
				Type:   "signup_urff",
				Block:  blk.NumberU64(),
				TxHash: tt[i].Hash().Hex(),
			})
			// calculate management fee
			if blk.NSignups().Cmp(common.Big0) == 0 {
				remRewards.Add(remRewards, core.ManagementFee)
			} else if avg := new(big.Int).Div(blk.TotalWei(), blk.NSignups()); avg.Cmp(core.Big10k) <= 0 {
				remRewards.Add(remRewards, core.ManagementFee)
			}
			// the receiver address gets the remaining rewards and (if appliable) core.ManagementFee
			balChanges = append(balChanges, balChange{
				To:     core.PrivilegedAddressesReceivers[m.From()].Receiver.Hex(),
				Value:  remRewards.String(),
				Type:   "signup_management_fee",
				Block:  blk.NumberU64(),
				TxHash: tt[i].Hash().Hex(),
			})
		}
	}
	bb := dbd.bulk["blocks"]
	bb.b.Insert(b)
	bb.c++
	bb = dbd.bulk["transactions"]
	bb.b.Insert(txs...)
	bb.c += uint64(len(txs))
	bb = dbd.bulk["balance_changes"]
	bb.b.Insert(balChanges...)
	bb.c += uint64(len(balChanges))
	bb = dbd.bulk["links"]
	bb.b.Insert(links...)
	bb.c += uint64(len(links))
	for i := range dbd.bulk {
		if err := dbd.flushIfNecessary(i); err != nil {
			return err
		}
	}
	return nil
}

func (dbd *dbDumper) flush(bulkName string) error {
	b := dbd.bulk[bulkName]
	if b.c == 0 {
		return nil
	}
	_, err := b.b.Run()
	if err != nil {
		return err
	}
	fmt.Printf("Block Importer: imported %d records to \"%s\"\n", b.c, bulkName)
	return nil
}

func (dbd *dbDumper) resetBulk(bulkName string) {
	dbd.bulk[bulkName] = &bulk{dbd.session.DB(dbd.dbName).C(bulkName).Bulk(), 0}
}

func (dbd *dbDumper) flushAndReset(bulkName string) error {
	if dbd.bulk[bulkName].c == 0 {
		return nil
	}
	if err := dbd.flush(bulkName); err != nil {
		return err
	}
	dbd.resetBulk(bulkName)
	return nil
}

func (dbd *dbDumper) flushIfNecessary(bulkName string) error {
	b := dbd.bulk[bulkName]
	if b.c >= dbd.nDocs {
		return dbd.flushAndReset(bulkName)
	}
	return nil
}
