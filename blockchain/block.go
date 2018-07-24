package blockchain

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"
)

type Block struct {
	Index        int           `json:"index"`
	Timestamp    time.Time     `json:"timestamp"`
	Transactions []Transaction `json:"transactions"`
	Proof        int           `json:"proof"`
	PrevHash     string        `json:"prev_hash"`
}

// Hash provides a Sha256 hash of the block
func (b *Block) Hash() (string, error) {
	m, err := json.Marshal(b)
	if err != nil {
		return "", err
	}
	h := sha256.New()
	h.Write(m)
	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
