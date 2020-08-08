package blockChain

import (
	"encoding/hex"
	"github.com/dgraph-io/badger"
	"log"
)

//prefix to divide data in badger database
var(
	utxoPrefix = []byte("utxo-") //differentiate utxo set from blockchain
	prefixLength = len(utxoPrefix)
)

type UTXOSet struct{
	Blockchain *BlockChain
}


func (u UTXOSet) Reindex(){
	db := u.Blockchain.Database
	u.DeleteByPrefix(utxoPrefix)

	UTXO := u.Blockchain.FindUTXO()

	err := db.Update(func(txn *badger.Txn) error{
		for txId, outs := range UTXO{
			key, err := hex.DecodeString(txId)
			if err != nil{
				return err
			}
			key = append(utxoPrefix, key ...)
			err = txn.Set(key, outs.Serialize())
			Handle(err)
		}

		return nil
	})
	Handle(err)
}

func (u *UTXOSet) Update(block *Block){
	db := u.Blockchain.Database

	err := db.Update(func(txn *badger.Txn) error{
		for _, tx := range block.Transactions{
			if tx.IsMoneybase() == false{
				for _, in := range tx.Inputs{
					updateOuts := TxOutputs{}
					inID := append(utxoPrefix, in.ID ... )
					item, err := txn.Get(inID)
					Handle(err)

					var v []byte
					v, err = item.ValueCopy(v)
					Handle(err)
					outs := DeserializeOutputs(v)

					for outIdx, out := range outs.Outputs{
						if outIdx != in.Out{ //check if attached to input to check if spent
							updateOuts.Outputs = append(updateOuts.Outputs, out)
						}
					}

					if len(updateOuts.Outputs) == 0{
						if err := txn.Delete(inID); err != nil{
							log.Panic(err)
						}
					} else{
						if err := txn.Set(inID, updateOuts.Serialize()); err != nil{
							log.Panic(err)
						}
					}
				}
			}

			newOutputs := TxOutputs{}
			for _, out := range tx.Outputs{
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			txID := append(utxoPrefix, tx.ID ... )
			if err := txn.Set(txID, newOutputs.Serialize()); err != nil{
				log.Panic(err)
			}
		}
		return nil
	})
	Handle(err)
}

//count how many unspent transactions
func (u UTXOSet) CountTransactions() int{
	db := u.Blockchain.Database
	counter := 0

	err := db.View(func(txn *badger.Txn) error{
		opts := badger.DefaultIteratorOptions
		it := txn.NewIterator(opts)
		defer it.Close()
		for it.Seek(utxoPrefix); it.ValidForPrefix(utxoPrefix); it.Next(){
			counter++
		}

		return nil
	})
	Handle(err)
	return counter
}

func (u *UTXOSet) DeleteByPrefix(prefix []byte){
	deleteKeys := func(keysForDelete [][]byte) error{
		if err := u.Blockchain.Database.Update(func(txn *badger.Txn) error {
			for _, key:= range keysForDelete{
				if err := txn.Delete(key); err != nil{
					return err
				}
			}
			return nil
		}); err != nil{
			return err
		}
		return nil
	}

	collectSize := 100000
	u.Blockchain.Database.View(func(txn *badger.Txn) error{
		opts:= badger.DefaultIteratorOptions
		opts.PrefetchValues = false
		it := txn.NewIterator(opts)
		defer it.Close()

		keysForDelete := make([][]byte, 0, collectSize)
		keysCollected := 0
		for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next(){
			key := it.Item().KeyCopy(nil)
			keysForDelete = append(keysForDelete, key)
			keysCollected++
			if keysCollected == collectSize{
				if err := deleteKeys(keysForDelete); err != nil{
					log.Panic(err)
				}
				keysForDelete = make([][]byte, 0, collectSize)
				keysCollected = 0
			}
		}
		if keysCollected > 0 {
			if err := deleteKeys(keysForDelete); err != nil{
				log.Panic(err)
			}
		}
		return nil
	})
}
