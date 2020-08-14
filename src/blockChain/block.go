package blockChain

import (
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

type Block struct {
	Timestamp 	 int64
	Hash         []byte
	Transactions []*Transaction
	PrevHash     []byte //Hash of previous block
	Nonce        int // counter
	Height 		 int // order of blocks
}

/*provide unique hash of transactions combined*/
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte //arr of txns

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.Serialize())
	}
	tree := NewMerkleTree(txHashes)
	return tree.RootNode.Data
}

func CreateBlock(txs []*Transaction, prevHash []byte, height int) *Block {
	block := &Block{time.Now().Unix(), []byte{}, txs, prevHash, 0, height}
	pow := NewProof(block)
	nonce, hash := pow.Run() //generate hash with pow

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func Genesis(moneybase *Transaction) *Block {
	return CreateBlock([]*Transaction{moneybase}, []byte{}, 0)
}

/*Encode block into bytes*/
func (b *Block) Serialize() []byte {
	var res bytes.Buffer
	encoder := gob.NewEncoder(&res)

	err := encoder.Encode(b)

	Handle(err)

	//return bytes representation of block
	return res.Bytes()
}

/*Decode bytes into a block*/
func Deserialize(data []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(data))

	err := decoder.Decode(&block)

	Handle(err)

	return &block
}

/*function to handle error*/
func Handle(err error) {
	if err != nil {
		log.Panic(err)
	}
}
