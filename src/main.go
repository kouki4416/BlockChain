package main

import (
	"./cli"
	"./wallet"
	"os"
)

func main() {
	defer os.Exit(0)
	cmd := cli.CommandLine{}
	cmd.Run()

	w := wallet.MakeWallet()
	w.Address()
}
