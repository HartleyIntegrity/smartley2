package blockchain

import (
	"bytes"
	"crypto/sha256"
	"math/big"
	"strconv"
	"time"
)

func (b *Blockchain) ProofOfWork() int {
	var proof int
	var target big.Int

	// Update the target value initialization
	target.Lsh(big.NewInt(1), uint(256-24)) // Change the 24 to adjust the difficulty

	for {
		var hash big.Int
		hash.SetBytes(b.HashWithProof(proof))

		if hash.Cmp(&target) == -1 {
			break
		}

		proof++
	}

	return proof
}

func (b *Blockchain) HashWithProof(proof int) []byte {
	hash := sha256.New()

	data := bytes.Join(
		[][]byte{
			[]byte(strconv.Itoa(proof)),
			[]byte(b.LastBlock().PreviousHash),
		},
		[]byte{},
	)

	hash.Write(data)

	return hash.Sum(nil)
}

func (b *Blockchain) LastBlock() *Block {
	if len(b.chain) == 0 {
		return &Block{
			Index:        0,
			Timestamp:    strconv.FormatInt(time.Now().Unix(), 10),
			Transactions: []*Transaction{},
			Proof:        0,
			PreviousHash: "",
		}
	}

	return b.chain[len(b.chain)-1]
}
