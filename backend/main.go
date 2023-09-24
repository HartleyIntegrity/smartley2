package main

import (
	"log"
	"smartley-contracts/api"
	"smartley-contracts/blockchain"
	"smartley-contracts/contracts"
	"smartley-contracts/storage"
)

func main() {
	api.ExecutionEnvironments = make(map[string]*contracts.VMExecutionEnvironment)

	// Initialize the storage
	storage.Init()
	defer storage.DB.Close()

	initBlockchain()

	api.Start()
}

func initBlockchain() {
	log.Println("Initializing blockchain...")
	bc := blockchain.NewBlockchain()
	log.Println("Blockchain initialized")

	log.Println("Adding sample transaction...")
	bc.AddTransaction(&blockchain.Transaction{
		Sender:    "Alice",
		Recipient: "Bob",
		Payload:   "Sample transaction",
	})
	log.Println("Sample transaction added")

	log.Println("Mining a new block...")
	bc.AddBlock()
	log.Println("New block mined")
}
