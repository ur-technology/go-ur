package core_test

import (
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"math/big"
	"math/rand"
	"testing"

	"encoding/binary"

	"github.com/ur-technology/go-ur/accounts"
	"github.com/ur-technology/go-ur/common"
	"github.com/ur-technology/go-ur/core"
	"github.com/ur-technology/go-ur/crypto"
)

var (
	privKey     *ecdsa.PrivateKey
	privKeyAddr common.Address
	privKeyJson = []byte(`{"address":"5d32e21bf3594aa66c205fde8dbee3dc726bd61d","Crypto":{"cipher":"aes-128-ctr","ciphertext":"bd9b82bdeecdf80c22747c2c18c389f2ce8a653c16dfbe830b66843f25c96543","cipherparams":{"iv":"7506def4dfb65d150541d45322feefbe"},"kdf":"scrypt","kdfparams":{"dklen":32,"n":262144,"p":1,"r":8,"salt":"459c5c5cb4bcd402fbee2fa47b7c495d8b73e18fca476a191327cf970550ec4a"},"mac":"4cf2812e2e8bb628480ad16732dc51a82602bae192b4c2f09ce607485d5bde3a"},"id":"aa8ff3a6-826c-4ae8-967b-be398508baed","version":3}`)
)

// convert privileged key from JSON to *accounts.Key
func init() {
	k, err := accounts.DecryptKey(privKeyJson, "password")
	if err != nil {
		panic(err)
	}
	privKey = k.PrivateKey
	privKeyAddr = crypto.PubkeyToAddress(privKey.PublicKey)
}

// test the miners reward. the block miner should
// receive core.BlockRewards for mining the block
// and core.BonusRewads for every signup transaction
func TestMinersReward(t *testing.T) {
	// simulated backend
	sim, err := NewSimulator(core.GenesisAccount{Address: privKeyAddr, Balance: common.Ether})
	if err != nil {
		t.Error(err)
		return
	}
	// setup the miner account
	_, minerAddr, err := newKeyAddr()
	if err != nil {
		t.Error(err)
		return
	}
	// set coinbase
	sim.Coinbase = minerAddr
	// setup user account
	_, userAddr, err := newKeyAddr()
	if err != nil {
		t.Error(err)
		return
	}
	// mine for 100 blocks without any transaction
	minerBal := big.NewInt(0)
	for i := int64(0); i < 100; i++ {
		minerBal = new(big.Int).Add(minerBal, core.BlockReward)
		_, err := sim.Commit()
		if err != nil {
			t.Error(err)
			return
		}
		if err := addressHasBalance(sim.BlockChain, minerAddr, minerBal); err != nil {
			t.Error("block:", sim.BlockChain.CurrentBlock().Number(), err)
			return
		}
	}
	// mine another 100 blocks, with 1 signup transaction
	for i := int64(0); i < 100; i++ {
		addedBal := new(big.Int).Mul(big.NewInt(2), core.BlockReward)
		minerBal = new(big.Int).Add(minerBal, addedBal)
		sim.AddPendingTx(&TxData{From: privKey, To: userAddr, Value: big.NewInt(1), Data: []byte{1}})
		if _, err := sim.Commit(); err != nil {
			t.Error("block:", sim.BlockChain.CurrentBlock().Number(), err)
		}
		if err := addressHasBalance(sim.BlockChain, minerAddr, minerBal); err != nil {
			t.Error("block:", sim.BlockChain.CurrentBlock().Number(), err)
			return
		}
	}
	// mine another 100 blocks, with 2 signup transaction
	for i := int64(0); i < 100; i++ {
		addedBal := new(big.Int).Mul(big.NewInt(3), core.BlockReward)
		minerBal = new(big.Int).Add(minerBal, addedBal)
		for i := 0; i < 2; i++ {
			sim.AddPendingTx(&TxData{From: privKey, To: userAddr, Value: big.NewInt(1), Data: []byte{1}})
		}
		if _, err := sim.Commit(); err != nil {
			t.Error("block:", sim.BlockChain.CurrentBlock().Number(), err)
		}
		if err := addressHasBalance(sim.BlockChain, minerAddr, minerBal); err != nil {
			t.Error("block:", sim.BlockChain.CurrentBlock().Number(), err)
			return
		}
	}
}

// TestMembersRewardsTree creates a tree of members signups,
// signs the members and checks the balances
func TestMembersRewardsTree(t *testing.T) {
	// simulated backend
	sim, err := NewSimulator(core.GenesisAccount{Address: privKeyAddr, Balance: common.Ether})
	if err != nil {
		t.Error(err)
		return
	}
	// setup the miner account
	_, minerAddr, err := newKeyAddr()
	if err != nil {
		t.Error(err)
		return
	}
	// set coinbase
	sim.Coinbase = minerAddr
	// setup root node
	rootNode := &memberNode{key: privKey}
	rootNode.addr = crypto.PubkeyToAddress(rootNode.key.PublicKey)
	// create random member tree
	newRandomMemberTree(10, 3, rootNode)
	// save privileged address initial balance
	privInitialBal, err := addressBalance(sim.BlockChain, privKeyAddr)
	if err != nil {
		t.Error(err)
		return
	}
	// signup members and calculate balances
	balances := make(map[common.Address]*big.Int)
	signupMembers(sim, rootNode, minerAddr, []common.Address{}, balances)
	// add the privileged address initial balance
	addToBalance(balances, privKeyAddr, privInitialBal)
	// check address
	if err := checkBalances(sim.BlockChain, balances, minerAddr); err != nil {
		t.Error(err)
		return
	}
}

// TestMembersRewardChain creates a "chain" of referrals. privileged key signs member1,
// member1 signs member2 and so on until memberx-1 signs memberx
func TestMembersRewardChain(t *testing.T) {
	// simulated blockchain
	sim, err := NewSimulator(core.GenesisAccount{Address: privKeyAddr, Balance: common.Ether})
	if err != nil {
		t.Error(err)
		return
	}
	// setup the miner account
	_, minerAddr, err := newKeyAddr()
	if err != nil {
		t.Error(err)
		return
	}
	// set coinbase
	sim.Coinbase = minerAddr
	// setup root node
	rootNode := &memberNode{key: privKey}
	// create node chain
	rootNode.addr = crypto.PubkeyToAddress(rootNode.key.PublicKey)
	curNode := rootNode
	for i := 0; i < 20; i++ {
		n := newMember()
		curNode.signups = []*memberNode{n}
		curNode = n
	}
	// save privileged address initial balance
	privInitialBal, err := addressBalance(sim.BlockChain, privKeyAddr)
	if err != nil {
		t.Error(err)
		return
	}
	// signup members and calculate balances
	balances := make(map[common.Address]*big.Int)
	signupMembers(sim, rootNode, minerAddr, []common.Address{}, balances)
	// add the privileged address initial balance
	addToBalance(balances, privKeyAddr, privInitialBal)
	// check address
	if err := checkBalances(sim.BlockChain, balances, minerAddr); err != nil {
		t.Error(err)
		return
	}
}

func signupMembers(sim *Simulator, node *memberNode, minerAddr common.Address, chain []common.Address, balances map[common.Address]*big.Int) {
	var err error
	for _, m := range node.signups {
		m.signBlock, m.signTx, err = signMember(sim, m.addr, node.signBlock, node.signTx, node.addr == privKeyAddr)
		if err != nil {
			fmt.Println("oops:", err)
		}
		// the privileged address receives 1000 UR
		addToBalance(balances, privKeyAddr, core.PrivilegedAddressesReward)
		// the miner receives 7 UR for the block, 7 UR for the signup
		for i := 0; i < 2; i++ {
			addToBalance(balances, minerAddr, core.BlockReward)
		}
		// the member being signed up receives 2000 UR
		addToBalance(balances, m.addr, core.SignupReward)
		// build new reward chain
		newChain := make([]common.Address, 1, len(chain)+1)
		newChain[0] = m.addr
		newChain = append(newChain, chain...)
		if len(newChain) > 8 {
			newChain = newChain[:8]
		}
		// the remaining members receive depending on the level
		rem := core.TotalSingupRewards
		for i, a := range newChain[1:] {
			addToBalance(balances, a, core.MembersSingupRewards[i])
			rem = new(big.Int).Sub(rem, core.MembersSingupRewards[i])
		}
		// the privileged address receives the remaining rewards if any
		addToBalance(balances, privKeyAddr, rem)
		// continue down the tree
		signupMembers(sim, m, minerAddr, newChain, balances)
	}
}

func signMember(sim *Simulator, addr common.Address, block uint64, txHash common.Hash, fromPrivileged bool) (uint64, common.Hash, error) {
	var d []byte
	if fromPrivileged {
		d = make([]byte, 1)
	} else {
		d = make([]byte, 41)
		binary.BigEndian.PutUint64(d[1:], block)
		copy(d[9:], txHash[:])
	}
	d[0] = 1
	sim.AddPendingTx(&TxData{
		From:  privKey,
		To:    addr,
		Value: big.NewInt(1),
		Data:  d,
	})
	comm, err := sim.Commit()
	if err != nil {
		return 0, common.Hash{}, err
	}
	return sim.BlockChain.CurrentBlock().NumberU64(), comm[0].Tx.Hash(), nil
}

func addToBalance(bal map[common.Address]*big.Int, addr common.Address, value *big.Int) {
	var b *big.Int
	if v, ok := bal[addr]; ok {
		b = v
	} else {
		b = big.NewInt(0)
	}
	bal[addr] = new(big.Int).Add(b, value)
}

func checkBalances(bc *core.BlockChain, balances map[common.Address]*big.Int, minerAddr common.Address) error {
	expBal, ok := balances[privKeyAddr]
	if !ok {
		return fmt.Errorf("no address for the privileged address")
	}
	bal, err := addressBalance(bc, privKeyAddr)
	if err != nil {
		return err
	}
	if expBal.Cmp(bal) != 0 {
		return fmt.Errorf("got a different balance for the privileged address than expected (%s): %s\n", expBal, bal)
	}
	delete(balances, privKeyAddr)
	if expBal, ok = balances[minerAddr]; !ok {
		return fmt.Errorf("no address for the miner")
	}
	if bal, err = addressBalance(bc, minerAddr); err != nil {
		return err
	}
	if expBal.Cmp(bal) != 0 {
		return fmt.Errorf("got a different balance for the miner address than expected (%s): %s\n", expBal, bal)
	}
	delete(balances, minerAddr)
	for a, expBal := range balances {
		if bal, err = addressBalance(bc, a); err != nil {
			return err
		}
		if expBal.Cmp(bal) != 0 {
			return fmt.Errorf("got a different balance for the member %s than expected (%s): %s", hex.EncodeToString(a[:]), expBal, bal)
		}
	}
	return nil
}

type memberNode struct {
	addr      common.Address
	key       *ecdsa.PrivateKey
	signups   []*memberNode
	signTx    common.Hash
	signBlock uint64
}

func newMember() *memberNode {
	k, err := crypto.GenerateKey()
	if err != nil {
		panic(err)
	}
	return &memberNode{addr: crypto.PubkeyToAddress(k.PublicKey), key: k}
}

func newRandomMemberTree(depth, maxNodes int, rootNode *memberNode) {
	if depth == 0 {
		return
	}
	nNodes := rand.Intn(maxNodes) + 1
	rootNode.signups = make([]*memberNode, 0, nNodes)
	for i := 0; i < nNodes; i++ {
		n := newMember()
		rootNode.signups = append(rootNode.signups, n)
		newRandomMemberTree(depth-1, maxNodes, n)
	}
}

func newKeyAddr() (*ecdsa.PrivateKey, common.Address, error) {
	minerk, err := crypto.GenerateKey()
	if err != nil {
		return nil, common.Address{}, err
	}
	return minerk, crypto.PubkeyToAddress(minerk.PublicKey), nil
}

func addressBalance(bchain *core.BlockChain, addr common.Address) (*big.Int, error) {
	state, err := bchain.State()
	if err != nil {
		return nil, err
	}
	return state.GetBalance(addr), nil
}

func addressHasBalance(bchain *core.BlockChain, addr common.Address, exp *big.Int) error {
	bal, err := addressBalance(bchain, addr)
	if err != nil {
		return nil
	}
	if bal.Cmp(exp) == 0 {
		return nil
	}
	return fmt.Errorf("got a different balance than expected at address %s: %s (expected %s)", addr.Hex(), bal.String(), exp.String())
}
