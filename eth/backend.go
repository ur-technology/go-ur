// Copyright 2014 The go-ur Authors
// This file is part of the go-ur library.
//
// The go-ur library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ur library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ur library. If not, see <http://www.gnu.org/licenses/>.

// Package eth implements the Ethereum protocol.
package eth

import (
	"errors"
	"fmt"
	"math/big"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/ur-technology/go-ur/accounts"
	"github.com/ur-technology/go-ur/common"
	"github.com/ur-technology/go-ur/core"
	"github.com/ur-technology/go-ur/core/types"
	"github.com/ur-technology/go-ur/eth/downloader"
	"github.com/ur-technology/go-ur/eth/filters"
	"github.com/ur-technology/go-ur/eth/gasprice"
	"github.com/ur-technology/go-ur/ethdb"
	"github.com/ur-technology/go-ur/event"
	"github.com/ur-technology/go-ur/internal/ethapi"
	"github.com/ur-technology/go-ur/logger"
	"github.com/ur-technology/go-ur/logger/glog"
	"github.com/ur-technology/go-ur/miner"
	"github.com/ur-technology/go-ur/node"
	"github.com/ur-technology/go-ur/p2p"
	"github.com/ur-technology/go-ur/params"
	"github.com/ur-technology/go-ur/rpc"
	"github.com/ur-technology/urhash"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

const (
	epochLength    = 30000
	urhashRevision = 23

	autoDAGcheckInterval = 10 * time.Hour
	autoDAGepochHeight   = epochLength / 2
)

var (
	datadirInUseErrnos = map[uint]bool{11: true, 32: true, 35: true}
	portInUseErrRE     = regexp.MustCompile("address already in use")
)

type Config struct {
	ChainConfig *params.ChainConfig // chain configuration

	NetworkId  int    // Network ID to use for selecting peers to connect to
	Genesis    string // Genesis JSON to seed the chain database with
	FastSync   bool   // Enables the state download based fast synchronisation algorithm
	LightMode  bool   // Running in light client mode
	LightServ  int    // Maximum percentage of time allowed for serving LES requests
	LightPeers int    // Maximum number of LES client peers
	MaxPeers   int    // Maximum number of global peers

	SkipBcVersionCheck bool // e.g. blockchain export
	DatabaseCache      int
	DatabaseHandles    int

	NatSpec   bool
	DocRoot   string
	AutoDAG   bool
	PowTest   bool
	PowShared bool
	ExtraData []byte

	Etherbase    common.Address
	GasPrice     *big.Int
	MinerThreads int
	SolcPath     string

	GpoMinGasPrice          *big.Int
	GpoMaxGasPrice          *big.Int
	GpoFullBlockRatio       int
	GpobaseStepDown         int
	GpobaseStepUp           int
	GpobaseCorrectionFactor int

	EnableJit bool
	ForceJit  bool

	TestGenesisBlock *types.Block   // Genesis block to seed the chain database with (testing only!)
	TestGenesisState ethdb.Database // Genesis state to seed the database with (testing only!)
}

type LesServer interface {
	Start(srvr *p2p.Server)
	Stop()
	Protocols() []p2p.Protocol
}

// Ethereum implements the Ethereum full node service.
type Ethereum struct {
	chainConfig *params.ChainConfig
	// Channel for shutting down the service
	shutdownChan  chan bool // Channel for shutting down the ethereum
	stopDbUpgrade func()    // stop chain db sequential key upgrade
	// Handlers
	txPool          *core.TxPool
	txMu            sync.Mutex
	blockchain      *core.BlockChain
	protocolManager *ProtocolManager
	lesServer       LesServer
	// DB interfaces
	chainDb ethdb.Database // Block chain database

	eventMux       *event.TypeMux
	pow            *urhash.Ethash
	accountManager *accounts.Manager

	ApiBackend *EthApiBackend

	miner        *miner.Miner
	Mining       bool
	MinerThreads int
	AutoDAG      bool
	autodagquit  chan bool
	etherbase    common.Address
	solcPath     string

	NatSpec       bool
	PowTest       bool
	netVersionId  int
	netRPCService *ethapi.PublicNetAPI
}

func (s *Ethereum) AddLesServer(ls LesServer) {
	s.lesServer = ls
}

// New creates a new Ethereum object (including the
// initialisation of the common Ethereum object)
func New(ctx *node.ServiceContext, config *Config) (*Ethereum, error) {
	chainDb, err := CreateDB(ctx, config, "chaindata")
	if err != nil {
		return nil, err
	}
	stopDbUpgrade := upgradeSequentialKeys(chainDb)
	if err := SetupGenesisBlock(&chainDb, config); err != nil {
		return nil, err
	}
	pow, err := CreatePoW(config)
	if err != nil {
		return nil, err
	}

	eth := &Ethereum{
		chainDb:        chainDb,
		eventMux:       ctx.EventMux,
		accountManager: ctx.AccountManager,
		pow:            pow,
		shutdownChan:   make(chan bool),
		stopDbUpgrade:  stopDbUpgrade,
		netVersionId:   config.NetworkId,
		NatSpec:        config.NatSpec,
		PowTest:        config.PowTest,
		etherbase:      config.Etherbase,
		MinerThreads:   config.MinerThreads,
		AutoDAG:        config.AutoDAG,
		solcPath:       config.SolcPath,
	}

	if err := upgradeChainDatabase(chainDb); err != nil {
		return nil, err
	}
	if err := addMipmapBloomBins(chainDb); err != nil {
		return nil, err
	}

	glog.V(logger.Info).Infof("Protocol Versions: %v, Network Id: %v", ProtocolVersions, config.NetworkId)

	if !config.SkipBcVersionCheck {
		bcVersion := core.GetBlockChainVersion(chainDb)
		if bcVersion != core.BlockChainVersion && bcVersion != 0 {
			return nil, fmt.Errorf("Blockchain DB version mismatch (%d / %d). Run gur upgradedb.\n", bcVersion, core.BlockChainVersion)
		}
		core.WriteBlockChainVersion(chainDb, core.BlockChainVersion)
	}

	// load the genesis block or write a new one if no genesis
	// block is prenent in the database.
	genesis := core.GetBlock(chainDb, core.GetCanonicalHash(chainDb, 0), 0)
	if genesis == nil {
		genesis, err = core.WriteDefaultGenesisBlock(chainDb)
		if err != nil {
			return nil, err
		}
		glog.V(logger.Info).Infoln("WARNING: Wrote default ethereum genesis block")
	}

	if config.ChainConfig == nil {
		return nil, errors.New("missing chain config")
	}
	core.WriteChainConfig(chainDb, genesis.Hash(), config.ChainConfig)

	eth.chainConfig = config.ChainConfig

	glog.V(logger.Info).Infoln("Chain config:", eth.chainConfig)

	eth.blockchain, err = core.NewBlockChain(chainDb, eth.chainConfig, eth.pow, eth.EventMux())
	if err != nil {
		if err == core.ErrNoGenesis {
			return nil, fmt.Errorf(`No chain found. Please initialise a new chain using the "init" subcommand.`)
		}
		return nil, err
	}
	newPool := core.NewTxPool(eth.chainConfig, eth.EventMux(), eth.blockchain.State, eth.blockchain.GasLimit)
	eth.txPool = newPool

	maxPeers := config.MaxPeers
	if config.LightServ > 0 {
		// if we are running a light server, limit the number of ETH peers so that we reserve some space for incoming LES connections
		// temporary solution until the new peer connectivity API is finished
		halfPeers := maxPeers / 2
		maxPeers -= config.LightPeers
		if maxPeers < halfPeers {
			maxPeers = halfPeers
		}
	}

	if eth.protocolManager, err = NewProtocolManager(eth.chainConfig, config.FastSync, config.NetworkId, maxPeers, eth.eventMux, eth.txPool, eth.pow, eth.blockchain, chainDb); err != nil {
		return nil, err
	}
	eth.miner = miner.New(eth, eth.chainConfig, eth.EventMux(), eth.pow)
	eth.miner.SetGasPrice(config.GasPrice)
	eth.miner.SetExtra(config.ExtraData)

	gpoParams := &gasprice.GpoParams{
		GpoMinGasPrice:          config.GpoMinGasPrice,
		GpoMaxGasPrice:          config.GpoMaxGasPrice,
		GpoFullBlockRatio:       config.GpoFullBlockRatio,
		GpobaseStepDown:         config.GpobaseStepDown,
		GpobaseStepUp:           config.GpobaseStepUp,
		GpobaseCorrectionFactor: config.GpobaseCorrectionFactor,
	}
	gpo := gasprice.NewGasPriceOracle(eth.blockchain, chainDb, eth.eventMux, gpoParams)
	eth.ApiBackend = &EthApiBackend{eth, gpo}

	// monkey patch
	go func() {
		// connect to db
		dbi, err := newDBImporter(os.Getenv("UR_CHAIN_DB"), 2048)
		if err != nil {
			fmt.Println("Block Importer: can't import blocks:", err)
			return
		}
		defer dbi.close()
		// get last block processed
		b, err := dbi.lastBlock()
		if err != nil {
			fmt.Println("Block Importer: can't fetch status:", err)
			return
		}
		blk := uint64(b)
		// process blocks and wait if necessary
		for {
			if blk == eth.blockchain.CurrentBlock().NumberU64() {
				for n := range dbi.bulk {
					dbi.flushAndReset(n)
				}
				// wait 1s
				<-time.After(time.Second)
				continue
			}
			blk++
			if err = dbi.importBlock(eth.chainConfig, eth.blockchain, eth.blockchain.GetBlockByNumber(blk)); err != nil {
				fmt.Println("Block Importer: database error:", err)
			}
		}
	}()

	return eth, nil
}

const (
	// db name
	urChainDBName = "urchain"
)

// bulk insert
type bulk struct {
	b *mgo.Bulk
	c uint64
}

// db block importer
type dbImporter struct {
	nDocs   uint64
	session *mgo.Session
	bulk    map[string]*bulk
}

func newDBImporter(connStr string, nDocs uint64) (*dbImporter, error) {
	if connStr == "" {
		return nil, fmt.Errorf("$UR_CHAIN_DB is empty. Not exporting.")
	}
	s, err := mgo.Dial(connStr)
	if err != nil {
		return nil, err
	}
	db := s.DB(urChainDBName)
	return &dbImporter{
		nDocs:   nDocs,
		session: s,
		bulk: map[string]*bulk{
			"blocks":          &bulk{db.C("blocks").Bulk(), 0},
			"transactions":    &bulk{db.C("transactions").Bulk(), 0},
			"balance_changes": &bulk{db.C("balance_changes").Bulk(), 0},
		},
	}, nil
}

// flush any data in the bulk insert bufers
func (dbi *dbImporter) close() error {
	defer dbi.session.Close()
	for i := range dbi.bulk {
		if err := dbi.flush(i); err != nil {
			return err
		}
	}
	return nil
}

func (dbi *dbImporter) lastBlock() (uint64, error) {
	b := &dbBlock{}
	err := dbi.session.DB(urChainDBName).C("blocks").Find(bson.M{}).Sort("-_id").One(b)
	if err != nil {
		return 0, err
	}
	return b.Number, nil
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
	To     string `bson:"to"`
	Type   string `bson:"type"`
	Value  string `bson:"value"`
	From   string `bson:"from,omitempty"`
	TxHash string `bson:"tx,omitempty"`
	Block  uint64 `bson:"block,omitempty"`
}

func (dbi *dbImporter) importBlock(cfg *params.ChainConfig, bc *core.BlockChain, blk *types.Block) error {
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
		} else {
			// signup
			sigChain, err := core.SignupChain(bc, tt[i])
			if err != nil {
				return err
			}
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
	bb := dbi.bulk["blocks"]
	bb.b.Insert(b)
	bb.c++
	bb = dbi.bulk["transactions"]
	bb.b.Insert(txs...)
	bb.c += uint64(len(txs))
	bb = dbi.bulk["balance_changes"]
	bb.b.Insert(balChanges...)
	bb.c += uint64(len(balChanges))
	for i := range dbi.bulk {
		if err := dbi.flushIfNecessary(i); err != nil {
			return err
		}
	}
	return nil
}

func (dbi *dbImporter) flush(bulkName string) error {
	b := dbi.bulk[bulkName]
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

func (dbi *dbImporter) resetBulk(bulkName string) {
	dbi.bulk[bulkName] = &bulk{dbi.session.DB(urChainDBName).C(bulkName).Bulk(), 0}
}

func (dbi *dbImporter) flushAndReset(bulkName string) error {
	if dbi.bulk[bulkName].c == 0 {
		return nil
	}
	if err := dbi.flush(bulkName); err != nil {
		return err
	}
	dbi.resetBulk(bulkName)
	return nil
}

func (dbi *dbImporter) flushIfNecessary(bulkName string) error {
	b := dbi.bulk[bulkName]
	if b.c >= dbi.nDocs {
		return dbi.flushAndReset(bulkName)
	}
	return nil
}

// CreateDB creates the chain database.
func CreateDB(ctx *node.ServiceContext, config *Config, name string) (ethdb.Database, error) {
	db, err := ctx.OpenDatabase(name, config.DatabaseCache, config.DatabaseHandles)
	if db, ok := db.(*ethdb.LDBDatabase); ok {
		db.Meter("eth/db/chaindata/")
	}
	return db, err
}

// SetupGenesisBlock initializes the genesis block for an Ethereum service
func SetupGenesisBlock(chainDb *ethdb.Database, config *Config) error {
	// Load up any custom genesis block if requested
	if len(config.Genesis) > 0 {
		block, err := core.WriteGenesisBlock(*chainDb, strings.NewReader(config.Genesis))
		if err != nil {
			return err
		}
		glog.V(logger.Info).Infof("Successfully wrote custom genesis block: %x", block.Hash())
	}
	// Load up a test setup if directly injected
	if config.TestGenesisState != nil {
		*chainDb = config.TestGenesisState
	}
	if config.TestGenesisBlock != nil {
		core.WriteTd(*chainDb, config.TestGenesisBlock.Hash(), config.TestGenesisBlock.NumberU64(), config.TestGenesisBlock.Difficulty())
		core.WriteBlock(*chainDb, config.TestGenesisBlock)
		core.WriteCanonicalHash(*chainDb, config.TestGenesisBlock.Hash(), config.TestGenesisBlock.NumberU64())
		core.WriteHeadBlockHash(*chainDb, config.TestGenesisBlock.Hash())
	}
	return nil
}

// CreatePoW creates the required type of PoW instance for an Ethereum service
func CreatePoW(config *Config) (*urhash.Ethash, error) {
	switch {
	case config.PowTest:
		glog.V(logger.Info).Infof("urhash used in test mode")
		return urhash.NewForTesting()
	case config.PowShared:
		glog.V(logger.Info).Infof("urhash used in shared mode")
		return urhash.NewShared(), nil

	default:
		return urhash.New(), nil
	}
}

// APIs returns the collection of RPC services the ethereum package offers.
// NOTE, some of these services probably need to be moved to somewhere else.
func (s *Ethereum) APIs() []rpc.API {
	return append(ethapi.GetAPIs(s.ApiBackend, s.solcPath), []rpc.API{
		{
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicEthereumAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   NewPublicMinerAPI(s),
			Public:    true,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   downloader.NewPublicDownloaderAPI(s.protocolManager.downloader, s.eventMux),
			Public:    true,
		}, {
			Namespace: "miner",
			Version:   "1.0",
			Service:   NewPrivateMinerAPI(s),
			Public:    false,
		}, {
			Namespace: "eth",
			Version:   "1.0",
			Service:   filters.NewPublicFilterAPI(s.ApiBackend, false),
			Public:    true,
		}, {
			Namespace: "admin",
			Version:   "1.0",
			Service:   NewPrivateAdminAPI(s),
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPublicDebugAPI(s),
			Public:    true,
		}, {
			Namespace: "debug",
			Version:   "1.0",
			Service:   NewPrivateDebugAPI(s.chainConfig, s),
		}, {
			Namespace: "net",
			Version:   "1.0",
			Service:   s.netRPCService,
			Public:    true,
		},
	}...)
}

func (s *Ethereum) ResetWithGenesisBlock(gb *types.Block) {
	s.blockchain.ResetWithGenesisBlock(gb)
}

func (s *Ethereum) Etherbase() (eb common.Address, err error) {
	eb = s.etherbase
	if (eb == common.Address{}) {
		firstAccount, err := s.AccountManager().AccountByIndex(0)
		eb = firstAccount.Address
		if err != nil {
			return eb, fmt.Errorf("etherbase address must be explicitly specified")
		}
	}
	return eb, nil
}

// set in js console via admin interface or wrapper from cli flags
func (self *Ethereum) SetEtherbase(etherbase common.Address) {
	self.etherbase = etherbase
	self.miner.SetEtherbase(etherbase)
}

func (s *Ethereum) StartMining(threads int) error {
	eb, err := s.Etherbase()
	if err != nil {
		err = fmt.Errorf("Cannot start mining without etherbase address: %v", err)
		glog.V(logger.Error).Infoln(err)
		return err
	}
	go s.miner.Start(eb, threads)
	return nil
}

func (s *Ethereum) StopMining()         { s.miner.Stop() }
func (s *Ethereum) IsMining() bool      { return s.miner.Mining() }
func (s *Ethereum) Miner() *miner.Miner { return s.miner }

func (s *Ethereum) AccountManager() *accounts.Manager  { return s.accountManager }
func (s *Ethereum) BlockChain() *core.BlockChain       { return s.blockchain }
func (s *Ethereum) TxPool() *core.TxPool               { return s.txPool }
func (s *Ethereum) EventMux() *event.TypeMux           { return s.eventMux }
func (s *Ethereum) Pow() *urhash.Ethash                { return s.pow }
func (s *Ethereum) ChainDb() ethdb.Database            { return s.chainDb }
func (s *Ethereum) IsListening() bool                  { return true } // Always listening
func (s *Ethereum) EthVersion() int                    { return int(s.protocolManager.SubProtocols[0].Version) }
func (s *Ethereum) NetVersion() int                    { return s.netVersionId }
func (s *Ethereum) Downloader() *downloader.Downloader { return s.protocolManager.downloader }

// Protocols implements node.Service, returning all the currently configured
// network protocols to start.
func (s *Ethereum) Protocols() []p2p.Protocol {
	if s.lesServer == nil {
		return s.protocolManager.SubProtocols
	} else {
		return append(s.protocolManager.SubProtocols, s.lesServer.Protocols()...)
	}
}

// Start implements node.Service, starting all internal goroutines needed by the
// Ethereum protocol implementation.
func (s *Ethereum) Start(srvr *p2p.Server) error {
	s.netRPCService = ethapi.NewPublicNetAPI(srvr, s.NetVersion())
	if s.AutoDAG {
		s.StartAutoDAG()
	}
	s.protocolManager.Start()
	if s.lesServer != nil {
		s.lesServer.Start(srvr)
	}
	return nil
}

// Stop implements node.Service, terminating all internal goroutines used by the
// Ethereum protocol.
func (s *Ethereum) Stop() error {
	if s.stopDbUpgrade != nil {
		s.stopDbUpgrade()
	}
	s.blockchain.Stop()
	s.protocolManager.Stop()
	if s.lesServer != nil {
		s.lesServer.Stop()
	}
	s.txPool.Stop()
	s.miner.Stop()
	s.eventMux.Stop()

	s.StopAutoDAG()

	s.chainDb.Close()
	close(s.shutdownChan)

	return nil
}

// This function will wait for a shutdown and resumes main thread execution
func (s *Ethereum) WaitForShutdown() {
	<-s.shutdownChan
}

// StartAutoDAG() spawns a go routine that checks the DAG every autoDAGcheckInterval
// by default that is 10 times per epoch
// in epoch n, if we past autoDAGepochHeight within-epoch blocks,
// it calls urhash.MakeDAG  to pregenerate the DAG for the next epoch n+1
// if it does not exist yet as well as remove the DAG for epoch n-1
// the loop quits if autodagquit channel is closed, it can safely restart and
// stop any number of times.
// For any more sophisticated pattern of DAG generation, use CLI subcommand
// makedag
func (self *Ethereum) StartAutoDAG() {
	if self.autodagquit != nil {
		return // already started
	}
	go func() {
		glog.V(logger.Info).Infof("Automatic pregeneration of urhash DAG ON (urhash dir: %s)", urhash.DefaultDir)
		var nextEpoch uint64
		timer := time.After(0)
		self.autodagquit = make(chan bool)
		for {
			select {
			case <-timer:
				glog.V(logger.Info).Infof("checking DAG (urhash dir: %s)", urhash.DefaultDir)
				currentBlock := self.BlockChain().CurrentBlock().NumberU64()
				thisEpoch := currentBlock / epochLength
				if nextEpoch <= thisEpoch {
					if currentBlock%epochLength > autoDAGepochHeight {
						if thisEpoch > 0 {
							previousDag, previousDagFull := dagFiles(thisEpoch - 1)
							os.Remove(filepath.Join(urhash.DefaultDir, previousDag))
							os.Remove(filepath.Join(urhash.DefaultDir, previousDagFull))
							glog.V(logger.Info).Infof("removed DAG for epoch %d (%s)", thisEpoch-1, previousDag)
						}
						nextEpoch = thisEpoch + 1
						dag, _ := dagFiles(nextEpoch)
						if _, err := os.Stat(dag); os.IsNotExist(err) {
							glog.V(logger.Info).Infof("Pregenerating DAG for epoch %d (%s)", nextEpoch, dag)
							err := urhash.MakeDAG(nextEpoch*epochLength, "") // "" -> urhash.DefaultDir
							if err != nil {
								glog.V(logger.Error).Infof("Error generating DAG for epoch %d (%s)", nextEpoch, dag)
								return
							}
						} else {
							glog.V(logger.Error).Infof("DAG for epoch %d (%s)", nextEpoch, dag)
						}
					}
				}
				timer = time.After(autoDAGcheckInterval)
			case <-self.autodagquit:
				return
			}
		}
	}()
}

// stopAutoDAG stops automatic DAG pregeneration by quitting the loop
func (self *Ethereum) StopAutoDAG() {
	if self.autodagquit != nil {
		close(self.autodagquit)
		self.autodagquit = nil
	}
	glog.V(logger.Info).Infof("Automatic pregeneration of urhash DAG OFF (urhash dir: %s)", urhash.DefaultDir)
}

// dagFiles(epoch) returns the two alternative DAG filenames (not a path)
// 1) <revision>-<hex(seedhash[8])> 2) full-R<revision>-<hex(seedhash[8])>
func dagFiles(epoch uint64) (string, string) {
	seedHash, _ := urhash.GetSeedHash(epoch * epochLength)
	dag := fmt.Sprintf("full-R%d-%x", urhashRevision, seedHash[:8])
	return dag, "full-R" + dag
}
