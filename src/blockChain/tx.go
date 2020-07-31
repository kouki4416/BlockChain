package blockChain

type TxOutput struct {
	Value  float64 //amount of money
	PubKey string  //needed to unlock token(use name for phase1)
}

type TxInput struct {
	ID  []byte //transaction which the output is in(e.g. txn x)
	Out int    //index of output
	Sig string //user name for phase1
}

/*check if signature of input is the same as data passed*/
func (in *TxInput) CanUnlock(data string) bool {
	return in.Sig == data
}

/*check if public key of the output is the same as data passed*/
func (out *TxOutput) CanBeUnlocked(data string) bool {
	return out.PubKey == data
}
