package blockChain

import "bytes"
import "../wallet"

type TxOutput struct {
	Value      float64 //amount of money
	PubKeyHash []byte  //needed to unlock token(use name for phase1)
}

type TxInput struct {
	ID        []byte //transaction which the output is in(e.g. txn x)
	Out       int    //index of output
	Signature []byte //user name for phase1
	PubKey    []byte
}

func NewTXOutput(value float64, address string) *TxOutput {
	txo := &TxOutput{value, nil}
	txo.Lock([]byte(address))

	return txo
}
func (in *TxInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wallet.PublicKeyHash(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

func (out *TxOutput) Lock(address []byte) {
	pubKeyHash := wallet.Base58Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

func (out *TxOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}
