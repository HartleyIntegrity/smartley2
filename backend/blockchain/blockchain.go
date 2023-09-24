package blockchain

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"log"
	"smartley-contracts/contracts"
	"smartley-contracts/types"
	"strconv"
	"time"
)

type Transaction struct {
	Sender            string        `json:"sender"`
	Recipient         string        `json:"recipient"`
	Payload           string        `json:"payload"`
	Contract          []byte        // Existing field
	FunctionSignature string        // Existing field
	ABI               []byte        // New field: ABI data
	Arguments         []interface{} // New field: function arguments
}

type Block struct {
	Index        int            `json:"index"`
	Timestamp    string         `json:"timestamp"`
	Transactions []*Transaction `json:"transactions"`
	Proof        int            `json:"proof"`
	PreviousHash string         `json:"previous_hash"`
}

type Blockchain struct {
	chain               []*Block
	currentTransactions []*Transaction
	State               map[string]types.Storage // Maps contract addresses to their storage states
}

func (b *Blockchain) GetCurrentTransactions() []*Transaction {
	return b.currentTransactions
}

func (b *Blockchain) Chain() []*Block {
	return b.chain
}

func NewBlockchain() *Blockchain {
	log.Println("Creating new Blockchain instance")
	b := &Blockchain{
		chain:               make([]*Block, 0),
		currentTransactions: make([]*Transaction, 0),
		State:               make(map[string]types.Storage),
	}

	log.Println("Adding genesis block")
	b.AddBlock()
	log.Println("Genesis block added")

	return b
}

func (b *Blockchain) AddTransaction(transaction *Transaction) {
	b.currentTransactions = append(b.currentTransactions, transaction)
}

func (b *Blockchain) AddBlock() {
	log.Println("Adding block to the chain")

	var previousHash string

	if len(b.chain) == 0 {
		previousHash = "1"
	} else {
		lastBlock := b.chain[len(b.chain)-1]
		previousHash = b.Hash(lastBlock)
	}

	block := &Block{
		Index:        len(b.chain) + 1,
		Timestamp:    strconv.FormatInt(time.Now().Unix(), 10),
		Transactions: b.currentTransactions,
		Proof:        0,
		PreviousHash: previousHash,
	}

	log.Println("Calculating proof of work")
	proof := b.ProofOfWork()
	block.Proof = proof
	log.Println("Proof of work calculated:", proof)

	log.Println("Appending block to the chain")
	b.chain = append(b.chain, block)
	log.Println("Block added to the chain")

	// Execute transactions
	for _, tx := range b.currentTransactions {
		b.HandleTransaction(tx)
	}

	// Clear the current transaction pool
	b.currentTransactions = []*Transaction{}
}

func (b *Blockchain) Hash(block *Block) string {
	hash := sha256.New()

	hash.Write([]byte(block.PreviousHash +
		strconv.Itoa(block.Proof) +
		block.Timestamp,
	))

	return hex.EncodeToString(hash.Sum(nil))
}

func GenerateContractAddress(sender string, numContracts int) string {
	h := sha256.New()
	h.Write([]byte(sender + strconv.Itoa(numContracts)))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)[:20] // Truncate the address for simplicity
}

type BlockchainWrapper struct {
	*Blockchain
	State map[string]types.Storage
}

func (bw *BlockchainWrapper) ExecuteFunction(contractAddress, functionSignature string, args []interface{}) error {
	storage, ok := bw.State[contractAddress]
	if !ok {
		return fmt.Errorf("contract not found at address: %s", contractAddress)
	}

	env := &contracts.VMExecutionEnvironment{
		Stack:          make(types.Stack, 0),
		Memory:         make(types.Memory, 0),
		Storage:        storage,
		ProgramCounter: 0,
		Bytecode:       storage.GetBytecode(),
		ABI:            storage.GetABI(),
	}
	_, err := env.ExecuteWithArgs(functionSignature, args)
	if err == nil {
		bw.State[contractAddress] = storage
	}
	return err
}

var BlockchainInstance *BlockchainWrapper

func (bw *BlockchainWrapper) Init() {
	BlockchainInstance = bw
	BlockchainInstance.State = make(map[string]types.Storage)
}

var _ contracts.ContractHandler = (*BlockchainWrapper)(nil)

func (bw *BlockchainWrapper) DeployContract(contract *contracts.Contract) error {
	// Compile the Solidity contract code into bytecode
	compiledContract, err := contracts.CompileSolidityString(contract.SoliditySource)
	if err != nil {
		return err
	}

	// Generate a contract address
	contractAddress := contracts.GenerateContractAddress()

	// Create a transaction with the contract bytecode as the payload and add it to the current transaction pool
	txn := &Transaction{
		Sender:    "0",
		Recipient: contractAddress,
		Payload:   compiledContract,
	}
	bw.AddTransaction(txn)

	// Update the contract object with the contract address
	contract.Address = contractAddress

	return nil
}

func (bw *Blockchain) HandleTransaction(tx *Transaction) {
	if len(tx.Contract) > 0 {
		// Deploy a new smart contract
		contractAddress := GenerateContractAddress(tx.Sender, len(bw.State))
		storage := make(types.Storage)
		env := &contracts.VMExecutionEnvironment{
			Stack:          make(types.Stack, 0),
			Memory:         make(types.Memory, 0),
			Storage:        storage,
			ProgramCounter: 0,
			Bytecode:       tx.Contract,
			ABI:            tx.ABI, // Include the ABI from the transaction
		}
		_, err := env.ExecuteWithArgs("", nil) // Use ExecuteWithArgs with empty function signature and no arguments
		if err == nil {
			bw.State[contractAddress] = storage
		}
	} else if len(tx.FunctionSignature) > 0 {
		// Execute a smart contract function
		storage, ok := bw.State[tx.Recipient]
		if !ok {
			log.Printf("Contract not found at address: %s\n", tx.Recipient)
			return
		}

		env := &contracts.VMExecutionEnvironment{
			Stack:          make(types.Stack, 0),
			Memory:         make(types.Memory, 0),
			Storage:        storage,
			ProgramCounter: 0,
			Bytecode:       storage.GetBytecode(),
			ABI:            storage.GetABI(), // Use the GetABI method
		}
		_, err := env.ExecuteWithArgs(tx.FunctionSignature, tx.Arguments) // Use ExecuteWithArgs with function signature and arguments from the transaction
		if err == nil {
			bw.State[tx.Recipient] = storage
		}
	}
}
