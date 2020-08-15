package cli

import (
	"../blockChain"
	"../wallet"
	"../network"
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strconv"
	_ "text/template/parse"
)

type CommandLine struct{}

/*commands for user*/
func (cli *CommandLine) printUsage() {
	fmt.Println("Usage:")
	fmt.Println(" getbalance -address ADDRESS - get the balance for an address")
	fmt.Println(" createblockchain -address ADDRESS creates a blockchain and sends genesis reward to address")
	fmt.Println(" printchain - Prints the blocks in the chain")
	fmt.Println(" send -from FROM -to TO -amount AMOUNT -mine - Send amount of coins")
	fmt.Println(" createaccount -hascar true/false -cartype Type -name Name Creates a new account")
	fmt.Println(" listaccount - Lists the accounts in our account file")
	fmt.Println(" reindexUTXO - Rebuilds the UTXO set")
	fmt.Println(" startnode -miner ADDRESS - Start a node with ID specified in NODE_ID env")
	fmt.Println(" buytoken -address ADDRESS -amount AMOUNT - buy specified amount of token")

}

func (cli *CommandLine) listAccount(nodeID string) {
	wallets, _ := wallet.CreateWallets(nodeID)

	addresses := wallets.GetAllAddresses()
	//name:= wallets.GetALLNames()
	//
	//carType := wallets.GetAllCarType()
	for _, address := range addresses {
		fmt.Println(address)
		fmt.Println(wallets.Wallets[address].Name)

		fmt.Println(wallets.Wallets[address].CarType)
		fmt.Println(wallets.Wallets[address].HasCar)
	}

}
func (cli *CommandLine) createAccount(nodeID string, name, carType string, hasCar bool) {
	fmt.Printf("New name is %s\n", name)
	fmt.Printf("New name is %s\n", carType)
	wallets, _ := wallet.CreateWallets(nodeID)
	address := wallets.AddWallet(name, carType, hasCar)
	wallets.SaveFile(nodeID)

	fmt.Printf("New address is %s\n", address)

}

/*check if passed args are valid*/
func (cli *CommandLine) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		runtime.Goexit() //exit app with garbage collection of database
	}
}

func (cli *CommandLine) StartNode(nodeID, minerAddress string) {
	fmt.Printf("Starting Node %s\n", nodeID)

	if len(minerAddress) > 0 {
		if wallet.ValidateAddress(minerAddress) {
			fmt.Println("Mining is on. Address to receive rewards: ", minerAddress)
		} else {
			log.Panic("Wrong miner address!")
		}
	}
	network.StartServer(nodeID, minerAddress)
}

func (cli *CommandLine) printChain(nodeID string) {
	chain := blockChain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	iter := chain.Iterator()

	for {
		block := iter.Next()

		fmt.Printf("Prev. hash: %x\n", block.PrevHash)
		fmt.Printf("Hash: %x\n", block.Hash)
		pow := blockChain.NewProof(block)
		fmt.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Println(tx)
		}
		fmt.Println()
		//genesis block does not have previous hash
		if len(block.PrevHash) == 0 {
			break
		}
	}
}

func (cli *CommandLine) createBlockChain(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := blockChain.InitBlockChain(address, nodeID)
	chain.Database.Close()

	UTXOSet := blockChain.UTXOSet{chain}
	UTXOSet.Reindex()

	fmt.Println("Finished!")
}

func (cli *CommandLine) getBalance(address, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}
	chain := blockChain.ContinueBlockChain(nodeID)
	UTXOSet := blockChain.UTXOSet{chain}
	defer chain.Database.Close()

	var balance float64 = 0
	pubKeyHash := wallet.Base58Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUnspentTransactions(pubKeyHash)

	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of %s: %f\n", address, balance)
}

func (cli *CommandLine) send(from, to string, amount float64, nodeID string, mineNow bool) {
	if !wallet.ValidateAddress(from) {
		log.Panic("Address is not Valid")
	}
	if !wallet.ValidateAddress(to) {
		log.Panic("Address is not Valid")
	}
	chain := blockChain.ContinueBlockChain(nodeID)
	UTXOSet := blockChain.UTXOSet{chain}
	defer chain.Database.Close()

	wallets, err := wallet.CreateWallets(nodeID)
	if err != nil{
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := blockChain.NewTransaction(&wallet, to, amount, &UTXOSet)
	if mineNow{
		log.Printf("Mining!!!")
		cbTx := blockChain.MoneybaseTx(from, "")
		txs := []*blockChain.Transaction{cbTx, tx}
		block := chain.MineBlock(txs)
		UTXOSet.Update(block)
	} else{
		network.SendTx(network.KnownNodes[0], tx)
		fmt.Println("send tx")
	}
	fmt.Println("success!")
}

func(cli *CommandLine) reindexUTXO(nodeID string){
	chain := blockChain.ContinueBlockChain(nodeID)
	defer chain.Database.Close()
	UTXOSet := blockChain.UTXOSet{chain}
	UTXOSet.Reindex()

	count := UTXOSet.CountTransactions()
	fmt.Printf("Done. there are %d transactions in the UTXO set. \n ", count)
}

func (cli *CommandLine) buyToken(address string, amount float64, nodeID string) {
	if !wallet.ValidateAddress(address) {
		log.Panic("Address is not Valid")
	}

	chain := blockChain.ContinueBlockChain(nodeID)
	UTXOSet := blockChain.UTXOSet{chain}
	defer chain.Database.Close()

	//wallets, err := wallet.CreateWallets(nodeID)
	//if err != nil{
	// log.Panic(err)
	//}
	//wallet := wallets.GetWallet(address)

	//tx := blockChain.NewTransaction(&wallet, to, amount, &UTXOSet)
	log.Printf("Mining!!!")
	cbTx := blockChain.BuyTransaction(address, "", amount)
	txs := []*blockChain.Transaction{cbTx}
	block := chain.MineBlock(txs)
	UTXOSet.Update(block)
	fmt.Println("success!")
}

func (cli *CommandLine) Run() {
	cli.validateArgs()

	nodeID := os.Getenv("NODE_ID")
	if nodeID == ""{
		fmt.Printf("NODE_ID env is not set")
		runtime.Goexit()
	}

	//set up flags for user arguments
	getBalanceCmd := flag.NewFlagSet("getbalance", flag.ExitOnError)
	createBlockchainCmd := flag.NewFlagSet("createblockchain", flag.ExitOnError)
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)
	createAccountCmd := flag.NewFlagSet("createaccount", flag.ExitOnError)
	listAccountCmd := flag.NewFlagSet("listaccount", flag.ExitOnError)
	reindexUTXOCmd := flag.NewFlagSet("reindexutxo", flag.ExitOnError)
	startNodeCmd := flag.NewFlagSet("startnode", flag.ExitOnError)
	buyTokenCmd := flag.NewFlagSet("buytoken", flag.ExitOnError)

	getBalanceAddress := getBalanceCmd.String("address", "", "The address to get balance for")
	createBlockchainAddress := createBlockchainCmd.String("address", "", "The address to send genesis block reward to")
	buyTokenAddress := buyTokenCmd.String("address", "", "The address add money")
	createAccountName := createAccountCmd.String("name", "erer", "The name of owner")
	createAccountCarType := createAccountCmd.String("cartype", "erer", "car type")
	createAccountHasCar := createAccountCmd.Bool("hascar", false, "if the owner has car")
	sendFrom := sendCmd.String("from", "", "Source wallet address")
	sendTo := sendCmd.String("to", "", "Destination wallet address")
	sendAmount := sendCmd.Float64("amount", 0, "Amount to send")
	buyTokenAmount := buyTokenCmd.Float64("amount", 0, "Amount to buy")
	sendMine := sendCmd.Bool("mine", false, "Mine immediately on the same node")
	startNodeMiner := startNodeCmd.String("miner", "", "Enable mining mode and send reward")

	switch os.Args[1] {
	case "reindexutxo":
		err := reindexUTXOCmd.Parse(os.Args[2:])
		if err != nil{
			log.Panic(err)
		}
	case "getbalance":
		err := getBalanceCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "startnode":
		err := startNodeCmd.Parse(os.Args[2:])
		if err != nil{
			log.Panic(err)
		}
	case "createblockchain":
		err := createBlockchainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "listaccount":
		err := listAccountCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "createaccount":
		err := createAccountCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "send":
		err := sendCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	case "buytoken":
		err := buyTokenCmd.Parse(os.Args[2:])
		if err != nil {
			log.Panic(err)
		}
	default:
		cli.printUsage()
		runtime.Goexit()
	}

	if getBalanceCmd.Parsed() {
		if *getBalanceAddress == "" {
			getBalanceCmd.Usage()
			runtime.Goexit()
		}
		cli.getBalance(*getBalanceAddress, nodeID)
	}

	if createBlockchainCmd.Parsed() {
		if *createBlockchainAddress == "" {
			createBlockchainCmd.Usage()
			runtime.Goexit()
		}
		cli.createBlockChain(*createBlockchainAddress, nodeID)
	}

	if printChainCmd.Parsed() {
		cli.printChain(nodeID)
	}
	if createAccountCmd.Parsed() {
		cli.createAccount(nodeID,*createAccountName, *createAccountCarType, *createAccountHasCar)
	}
	if listAccountCmd.Parsed() {
		cli.listAccount(nodeID)
	}
	if reindexUTXOCmd.Parsed(){
		cli.reindexUTXO(nodeID)
	}
	if sendCmd.Parsed() {
		if *sendFrom == "" || *sendTo == "" || *sendAmount <= 0 {
			sendCmd.Usage()
			runtime.Goexit()
		}

		cli.send(*sendFrom, *sendTo, *sendAmount, nodeID, *sendMine)
	}
	if startNodeCmd.Parsed(){
		nodeID := os.Getenv("NODE_ID")
		if nodeID == ""{
			startNodeCmd.Usage()
			runtime.Goexit()
		}
		cli.StartNode(nodeID, *startNodeMiner)
	}
	if buyTokenCmd.Parsed(){
		if *buyTokenAddress == "" || *buyTokenAmount <= 0 {
			buyTokenCmd.Usage()
			runtime.Goexit()
		}

		cli.buyToken(*buyTokenAddress, *buyTokenAmount, nodeID)
	}
}

