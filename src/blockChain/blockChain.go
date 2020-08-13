package blockChain

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"fmt"
	"github.com/dgraph-io/badger"
	"github.com/pkg/errors"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

const (
	dbPath      = "./tmp/blocks_%s"
	genesisData = "First Transaction from Genesis"
)

//Metadata of all block, Chain state
//Each block has one file
type BlockChain struct {
	LastHash []byte //Last hash of the last block
	Database *badger.DB
}

//iterator to go through chain
type BlockChainIterator struct {
	CurrentHash []byte
	Database    *badger.DB
}

func DBexists(path string) bool {
	if _, err := os.Stat(path + "/MANIFEST"); os.IsNotExist(err) {
		return false
	}

	return true
}

/**/
func ContinueBlockChain(nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	if DBexists(path) == false {
		fmt.Println("No existing blockchain found, create one!")
		runtime.Goexit()
	}

	var lastHash []byte

	//set db
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := openDB(path, opts)
	Handle(err)

	//get last hash to continue from existing block chain
	err = db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(lastHash)

		return err
	})
	Handle(err)

	chain := BlockChain{lastHash, db}

	return &chain
}

func InitBlockChain(address, nodeId string) *BlockChain {
	path := fmt.Sprintf(dbPath, nodeId)
	var lastHash []byte

	//check if db exists
	if DBexists(path) {
		fmt.Println("Blockchain already exists")
		runtime.Goexit()
	}

	//Set up path for the database
	opts := badger.DefaultOptions(dbPath)
	opts.Dir = dbPath
	opts.ValueDir = dbPath

	db, err := openDB(path, opts)
	Handle(err)

	err = db.Update(func(txn *badger.Txn) error {
		//make genesis block with first money transaction
		mtx := MoneybaseTx(address, genesisData)
		genesis := Genesis(mtx)
		fmt.Println("Genesis created")
		err = txn.Set(genesis.Hash, genesis.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), genesis.Hash)

		lastHash = genesis.Hash

		return err

	})

	Handle(err)

	blockchain := BlockChain{lastHash, db}
	return &blockchain
}

/**/
func (chain *BlockChain) MineBlock(transactions []*Transaction) *Block {
	var lastHash []byte
	var lastBlockData []byte
	var lastHeight int

	for _, tx := range transactions{
		if chain.VerifyTransaction(tx) != true{
			log.Panic("Invalid Transaction")
		}
	}

	//Just copy the last hash and return
	err := chain.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, err = item.ValueCopy(lastHash)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ := item.ValueCopy(lastBlockData)

		lastBlock := Deserialize(lastBlockData)

		lastHeight = lastBlock.Height

		return err
	})
	Handle(err)

	//Create a new block with last hash
	newBlock := CreateBlock(transactions, lastHash, lastHeight+1)

	//Put (hash, serialized block) into db
	//Associate lh with hash to easily get last hash
	err = chain.Database.Update(func(txn *badger.Txn) error {
		err := txn.Set(newBlock.Hash, newBlock.Serialize())
		Handle(err)
		err = txn.Set([]byte("lh"), newBlock.Hash)

		chain.LastHash = newBlock.Hash

		return err
	})
	Handle(err)

	return newBlock
}

func (chain *BlockChain) AddBlock(block *Block){
	var lastHash []byte
	var lastBlockData []byte
	err := chain.Database.Update(func(txn *badger.Txn) error{
		if _, err := txn.Get(block.Hash); err == nil{
			return nil
		}

		blockData := block.Serialize()
		err := txn.Set(block.Hash, blockData)
		Handle(err)

		item, err := txn.Get([]byte("lh"))
		Handle(err)
		lastHash, _ = item.ValueCopy(lastHash)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ = item.ValueCopy(lastBlockData)

		lastBlock := Deserialize(lastBlockData)

		if block.Height > lastBlock.Height {
			err = txn.Set([]byte("lh"), block.Hash)
			Handle(err)
			chain.LastHash = block.Hash
		}

		return nil
	})
	Handle(err)
}

func (chain *BlockChain) GetBlock(blockHash []byte) (Block, error) {
	var block Block
	var blockData []byte

	err := chain.Database.View(func(txn *badger.Txn) error{
		if item, err := txn.Get(blockHash); err != nil{
			return errors.New("Block not found")
		} else{
			blockData, _ = item.ValueCopy(blockData)
			block = *Deserialize(blockData)
		}
		return nil
	})

	if err != nil{
		return block, err // return empty block and err
	}

	return block, nil
}

//check if the copy of the block chain is the same
func (chain *BlockChain) GetBlockHashes() [][]byte{
	var blocks [][]byte

	iter := chain.Iterator()

	for{
		block := iter.Next()

		blocks = append(blocks, block.Hash)

		if len(block.PrevHash) == 0{
			break
		}
	}
	return blocks
}

func (chain *BlockChain) GetBestHeight() int{
	var lastBlock Block
	var lastHash[] byte
	var lastBlockData[] byte

	err := chain.Database.View(func(txn *badger.Txn) error{
		item, err := txn.Get([]byte("lh"))//get lasthash
		Handle(err)
		lastHash, err = item.ValueCopy(lastHash)

		item, err = txn.Get(lastHash)
		Handle(err)
		lastBlockData, _ = item.ValueCopy(lastBlockData)

		lastBlock = *Deserialize(lastBlockData)
		return nil
	})
	Handle(err)

	return lastBlock.Height
}

func (chain *BlockChain) Iterator() *BlockChainIterator {
	iter := &BlockChainIterator{chain.LastHash, chain.Database}

	return iter
}

//move iter to one previous block
func (iter *BlockChainIterator) Next() *Block {
	var block *Block
	var encodeBlock []byte
	err := iter.Database.View(func(txn *badger.Txn) error {
		item, err := txn.Get(iter.CurrentHash) //get block data using hash
		encodeBlock, err := item.ValueCopy(encodeBlock)
		block = Deserialize(encodeBlock) //decode
		return err
	})
	Handle(err)

	iter.CurrentHash = block.PrevHash //move one back
	return block
}


/*unspent transaction output*/
func (chain *BlockChain) FindUTXO() map[string]TxOutputs {
	UTXO := make(map[string]TxOutputs) //unspent transactions
	spentTXOs := make(map[string][]int)

	iter := chain.Iterator()

	for{
		block := iter.Next()

		for _, tx := range block.Transactions{
			txID := hex.EncodeToString(tx.ID)
		Outputs:
			for outIdx, out := range tx.Outputs{
				if spentTXOs[txID] != nil{
					for _, spentOut := range spentTXOs[txID]{
						if spentOut == outIdx{
							continue Outputs
						}
					}
				}
				outs := UTXO[txID]
				outs.Outputs = append(outs.Outputs, out)
				UTXO[txID] = outs // append another output
			}
			if tx.IsMoneybase() == false{
				for _, in := range tx.Inputs{
					inTxID := hex.EncodeToString(in.ID)
					spentTXOs[inTxID] = append(spentTXOs[inTxID], in.Out)
				}
			}
		}
		if len(block.PrevHash) == 0{
			break
		}
	}
	return UTXO
}



func (bc *BlockChain) FindTransaction(ID []byte) (Transaction, error) {
	iter := bc.Iterator()

	for {
		block := iter.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}
		if len(block.PrevHash) == 0 {
			break
		}
	}
	return Transaction{}, errors.New("Transaction does not exist")
}

func (bc *BlockChain) SignTransaction(tx *Transaction, privKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sign(privKey, prevTXs)
}

func (bc *BlockChain) VerifyTransaction(tx *Transaction) bool {
	if tx.IsMoneybase(){
		return true //check if transaction is money base
	}


	prevTXs := make(map[string]Transaction)

	for _, in := range tx.Inputs {
		prevTX, err := bc.FindTransaction(in.ID)
		Handle(err)
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
}

func retry(dir string, originalOpts badger.Options) (*badger.DB, error){
	lockPath := filepath.Join(dir, "LOCK")
	if err := os.Remove(lockPath); err != nil{
		return nil, fmt.Errorf(`removing "LOCK": %s`, err)
	}
	retryOPts := originalOpts
	retryOPts.Truncate = true
	db,err := badger.Open(retryOPts)
	return db, err
}

func openDB(dir string, opts badger.Options) (*badger.DB, error){
	if db, err := badger.Open(opts); err != nil{
		if strings.Contains(err.Error(), "LOCK"){// lock file exists -> db locked
			if db, err := retry(dir, opts); err == nil{
				log.Println("database unlocked")
				return db, nil
			}
			log.Println("could not unlock database")
		}
		return nil, err
	} else {
		return db, nil
	}

}