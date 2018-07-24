package hashcash

import (
	"crypto/sha256"
	"fmt"
	"strings"
)

// ProofOfWork works the node to produce a new block based on a simple
// Hashcash-like algorithm
func ProofOfWork(lastProof int) int {
	proof := 0
	for validProof(lastProof, proof) == false {
		proof++
	}
	return proof
}

func validProof(lastProof, proof int) bool {
	h := sha256.New()
	i := fmt.Sprintf("%d%d", lastProof, proof)
	h.Write([]byte(i))
	s := fmt.Sprintf("%x", h.Sum(nil))
	return strings.HasSuffix(s, "0000")
}
