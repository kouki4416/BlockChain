package blockChain

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
)

type Transaction struct {
	ID      []byte //hash
	Inputs  []TxInput
	Outputs []TxOutput
}

type TxOutput struct {
	Value  float64 //amount of money
	PubKey string //needed to unlock token(use name for phase1)
}

type TxInput struct {
	ID  []byte //transaction which the output is in(e.g. txn x)
	Out int //index of output
	Sig string //user name for phase1
}

/*create hash based on transaction byte data*/
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte
	encode := gob.NewEncoder(&encoded)
	err := encode.Encode(tx)
	Handle(err)

	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

/*base transaction to give $100*/
func MoneybaseTx(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Coins to %s", to)
	}

	txin := TxInput{[]byte{}, -1, data} // empty input
	txout := TxOutput{100, to} //give $100 beginning for simplicity

	tx := Transaction{nil, []TxInput{txin}, []TxOutput{txout}}
	tx.SetID()

	return &tx
}

func NewTransaction(from, to string, amount float64, chain *BlockChain) *Transaction {
	var inputs []TxInput
	var outputs []TxOutput

	acc, validOutputs := chain.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Panic("Error: not enough funds")
	}

	//
	for txid, outs := range validOutputs {
		txID, err := hex.DecodeString(txid)
		Handle(err)

		for _, out := range outs {
			input := TxInput{txID, out, from}
			inputs = append(inputs, input)
		}
	}

	//user sending money to address
	outputs = append(outputs, TxOutput{amount, to})

	//user receive change as output
	if acc > amount {
		outputs = append(outputs, TxOutput{acc - amount, from})
	}

	tx := Transaction{nil, inputs, outputs}
	tx.SetID()

	return &tx
}

/*check if txn is valid*/
func (tx *Transaction) IsMoneybase() bool {
	return len(tx.Inputs) == 1 && len(tx.Inputs[0].ID) == 0 && tx.Inputs[0].Out == -1
}

/*check if signature of input is the same as data passed*/
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

/*check if public key of the output is the same as data passed*/
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
