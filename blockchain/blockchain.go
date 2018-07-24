package blockchain

import (
	"sync"
	"time"
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
