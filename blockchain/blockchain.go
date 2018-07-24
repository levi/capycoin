package blockchain

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/levi/capycoin/hashcash"
	"github.com/levi/capycoin/nodes"
)

// Blockchain represents an in-memory blockchain with yet-to-be recorded
// transactions under PendingTransactions
type Blockchain struct {
	Chain               []Block       `json:"chain"`
	PendingTransactions []Transaction `json:"pending_transactions"`
	mutex               *sync.Mutex
}

// New returns a new blockchain.Blockchain with a genesis block
func New() Blockchain {
	b := Blockchain{make([]Block, 0), make([]Transaction, 0), &sync.Mutex{}}
	b.NewBlock(100, "1")
	return b
}

// NewBlock creates a new block with pending transactions and appends it to the blockchain
func (b *Blockchain) NewBlock(proof int, prevHash string) *Block {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	block := Block{
		len(b.Chain),
		time.Now(),
		b.PendingTransactions,
		proof,
		prevHash,
	}
	b.Chain = append(b.Chain, block)
	b.PendingTransactions = make([]Transaction, 0)
	return &block
}

// NewTransaction records a new pending transaction to be recorded in the blockchain
func (b *Blockchain) NewTransaction(sender, recipient string, amount int) int {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.PendingTransactions = append(b.PendingTransactions, Transaction{
		sender,
		recipient,
		amount,
	})
	return b.unsafeLastBlock().Index + 1
}

// LastBlock returns the last block in the blockchain
func (b *Blockchain) LastBlock() Block {
	b.mutex.Lock()
	defer b.mutex.Unlock()
	return b.unsafeLastBlock()
}

func (b *Blockchain) unsafeLastBlock() Block {
	return b.Chain[len(b.Chain)-1]
}

// ResolveConflicts derives consensus by identifying the longest available chain
func (b *Blockchain) ResolveConflicts(nodes *nodes.Nodes) (bool, error) {
	maxLength := len(b.Chain)
	var newChain []Block

	for _, host := range nodes.Addresses {
		resp, err := http.Get(fmt.Sprintf("http://%s/chain", host))
		if err != nil {
			return false, err
		}
		if resp.StatusCode == 200 {
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				return false, err
			}

			var c ChainResponse
			err = json.Unmarshal(body, &c)
			if err != nil {
				return false, err
			}

			length := c.Length
			chain := c.Chain

			if length > maxLength {
				isValid, _ := b.ValidChain(chain)
				if isValid {
					maxLength = length
					newChain = chain
				}
			}
		}
	}

	if newChain != nil {
		b.Chain = newChain
		return true, nil
	}

	return false, nil
}

func (b *Blockchain) ValidChain(chain []Block) (bool, error) {
	lastBlock := chain[0]
	i := 1
	for i < len(chain) {
		block := chain[i]
		lastHash, err := lastBlock.Hash()
		if err != nil {
			return false, err
		}
		if block.PrevHash != lastHash {
			return false, nil
		}
		if hashcash.ValidProof(lastBlock.Proof, block.Proof) == false {
			return false, nil
		}
		lastBlock = block
		i++
	}

	return true, nil
}

type ChainResponse struct {
	Chain  []Block `json:"chain"`
	Length int     `json:"length"`
}
