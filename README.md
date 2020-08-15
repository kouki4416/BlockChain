# Car share block-chain

This program allows you to create an accoount and store transactions inside the block-chain.

## Usage

There are following functionalities

  **createaccount**: Register a new account with personal information
  
  **listaccount**: List all account
  
  **getbalance**: View the balance of specified account
  
  **createblochchain**: Create a blockchain
  
  **send**: Make a send transaction
  
  **buytoken**: Add a new token(money) into block chain
  
  **startnode**: Start node and ready for reading bloch-chain from peer
  
 You can see the detail of these usage by running
  ```bash
  go run main.go
  ```
  
 For example you can do the following to make a transation
  ```bash
  go run main.go createaccount -name Master 
  go run main.go createaccount -name Ken -cartype Civic -hascar true
  go run main.go createaccount -name Jhon
  go run main.go createblockhain -address <pubkey hash produced after creating Master account>
  go run main.go buytoken -address <public key hash of Jhon> -amount <money ammount>
  go run main.go send -from <pubkey hash of Jhon> -to <pubkey hash of Ken> -amount <send amount>
  ```
  
