package blockChain

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"log"
	"math"
	"math/big"
)

/*Steps*/
// Take the data from the block

// create a counter (nonce) which starts at 0

// create a hash of the data plus the counter

// check the hash to see if it meets a set of requirements

// Requirements:
// The First few bytes must contain 0s depending on the difficulty

const Difficulty = 12

type ProofOfWork struct {
	Block  *Block
	Target *big.Int // Requirement derived from difficulty
}

func NewProof(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-Difficulty)) //left shift 1 to the left

	pow := &ProofOfWork{b, target}

	return pow
}

func (pow *ProofOfWork) InitData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevHash,
			pow.Block.HashTransactions(),
			ToHex(int64(nonce)),
			ToHex(int64(Difficulty)),
		},
		[]byte{},
	)

	return data
}

/*Main funciton of pow*/
func (pow *ProofOfWork) Run() (int, []byte) {
	var intHash big.Int
	var hash [32]byte

	nonce := 0

	//Difficult computation
	for nonce < math.MaxInt64 {
		//Prepare data
		data := pow.InitData(nonce)
		hash = sha256.Sum256(data)

		//see the hash calculation
		fmt.Printf("\r%x", hash)
		intHash.SetBytes(hash[:])

		// negative means the result has more preceding 0s than target -> break
		if intHash.Cmp(pow.Target) == -1 {
			break
		} else {
			//Add count to change hash
			nonce++
		}

	}
	fmt.Println()

	return nonce, hash[:]
}

/* 	validate if created new hash is valid.
	this will run quick because we already know what count(nonce) to use.
 */
func (pow *ProofOfWork) Validate() bool {
	var intHash big.Int

	data := pow.InitData(pow.Block.Nonce)

	hash := sha256.Sum256(data)
	intHash.SetBytes(hash[:])

	return intHash.Cmp(pow.Target) == -1
}

/*Convert int to hex byte*/
func ToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)

	}

	return buff.Bytes()
}
