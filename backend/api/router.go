package api

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"
	"time"

	"smartley-contracts/blockchain"
	"smartley-contracts/contracts"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

func init() {
	ExecutionEnvironments = make(map[string]*contracts.VMExecutionEnvironment)
}

type executeContractRequest struct {
	ContractAddress   string `json:"contract_address"`
	FunctionSignature string `json:"function_signature"`
}

var ExecutionEnvironments map[string]*contracts.VMExecutionEnvironment

func generateContractAddress() string {
	// Create a new random UUID
	newUUID, err := uuid.NewRandom()
	if err != nil {
		log.Fatalf("Failed to generate UUID: %v", err)
	}

	// Generate a hash of the UUID
	hash := sha256.Sum256([]byte(newUUID.String()))

	// Return the first 20 bytes of the hash as a hex-encoded string
	return hex.EncodeToString(hash[:20])
}

func createContract(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return
	}

	type soliditySource struct {
		Source            string `json:"source"`
		RicardianContract string `json:"ricardianContract"`
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "Missing Solidity source code in request body", http.StatusBadRequest)
		return
	}

	var sourceObj soliditySource
	err = json.Unmarshal(body, &sourceObj)
	if err != nil {
		http.Error(w, "Error parsing Solidity source code from JSON", http.StatusBadRequest)
		return
	}

	// Compile the Solidity contract code into bytecode and ABI
	abiBytes, bytecode, err := contracts.CompileSoliditySource(sourceObj.Source)
	if err != nil {
		log.Println("Error compiling Solidity code:", err)
		http.Error(w, "Error compiling Solidity code", http.StatusInternalServerError)
		return
	}

	// Remove metadata from the bytecode
	cleanedBytecode := removeMetadata(bytecode)

	log.Println("Cleaned Bytecode:", cleanedBytecode)

	// Create a new Contract instance
	contract := &contracts.Contract{
		SoliditySource: sourceObj.Source,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		Name:           "Contract Name",
		Source:         "Solidity",
		Bytecode:       cleanedBytecode,
		ABI:            abiBytes,
	}

	// Generate and set the Ricardian contract
	contract.RicardianContract = generateRicardianContract(contract)

	// Generate a contract address and set it
	contract.Address = generateContractAddress()

	// Set the contract ID to its address
	contract.ID = contract.Address

	// Save the contract to the database and deploy it to the blockchain
	_, err = contracts.CreateContract(contract, blockchain.BlockchainInstance)
	if err != nil {
		log.Println("Error creating and deploying the contract:", err)
		http.Error(w, "Error creating and deploying the contract", http.StatusInternalServerError)
		return
	}

	// Add the contract to the blockchain
	contractTransaction := &blockchain.Transaction{
		Sender:    "0",
		Recipient: contract.Address,
		Payload:   "",
		Contract:  []byte(contract.Bytecode),
	}

	blockchain.BlockchainInstance.AddTransaction(contractTransaction)
	blockchain.BlockchainInstance.AddBlock()

	// Create a new VMExecutionEnvironment for the contract
	env := contracts.NewVMExecutionEnvironment(contract)

	// Store the VMExecutionEnvironment in the ExecutionEnvironments map
	ExecutionEnvironments[contract.ID] = env

	// Respond with the created contract as JSON
	respJSON, err := json.Marshal(contract)
	if err != nil {
		log.Println("Error marshaling created contract:", err)
		http.Error(w, "Error marshaling created contract", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}

// removeMetadata searches for the Solidity contract metadata start sequence and removes it from the bytecode
func removeMetadata(bytecode string) string {
	metadataStart := "a165627a7a72305820"
	metadataIndex := strings.Index(bytecode, metadataStart)

	if metadataIndex != -1 {
		return bytecode[:metadataIndex]
	}

	return bytecode
}

func generateRicardianContract(contract *contracts.Contract) string {
	// Extract specific parts of the smart contract's Solidity source code
	// and convert them into human-readable format.
	// You can add your own custom logic here to extract relevant parts of the
	// Solidity source code and turn them into human-readable text.
	solidityCodeExample := contract.SoliditySource[:30] + "..."
	humanReadableCodeExample := "Example contract code: " + solidityCodeExample

	// Add dummy text and placeholder values to the Ricardian contract.
	dummyText := "This is a sample Ricardian contract generated automatically from the Solidity source code. The contract represents a simple agreement between parties and includes some example content from the Solidity source code."
	placeholderValue := "Placeholder value: 12345"

	ricardianContract := fmt.Sprintf(
		"Contract Name: %s\n"+
			"Contract ID: %s\n"+
			"Contract Address: %s\n"+
			"Created at: %s\n"+
			"Updated at: %s\n\n"+
			"%s\n\n"+
			"%s\n\n"+
			"%s",
		contract.Name,
		contract.ID,
		contract.Address,
		contract.CreatedAt.Format(time.RFC1123),
		contract.UpdatedAt.Format(time.RFC1123),
		dummyText,
		humanReadableCodeExample,
		placeholderValue,
	)

	return ricardianContract
}

func getContractByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractAddress := vars["id"] // Use contractAddress instead of contractID

	// Retrieve the VMExecutionEnvironment for the contract
	_, ok := ExecutionEnvironments[contractAddress]
	if !ok {
		http.Error(w, "VMExecutionEnvironment not found for contract", http.StatusNotFound)
		return
	}

	// Retrieve the contract object
	contract, err := contracts.GetContract(contractAddress) // Pass contractAddress instead of contractID
	if err != nil {
		http.Error(w, "Error retrieving contract", http.StatusInternalServerError)
		return
	}

	// Respond with the contract as JSON
	respJSON, err := json.Marshal(contract)
	if err != nil {
		log.Println("Error marshaling contract:", err)
		http.Error(w, "Error marshaling contract", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}

func rootHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]string{
		"message": "200 OK - Welcome to the root endpoint",
	}
	json.NewEncoder(w).Encode(response)
}

func executeContractFunction(w http.ResponseWriter, r *http.Request) {
	fmt.Println("executeContractFunction called")
	vars := mux.Vars(r)
	contractAddress := vars["id"] // Use contractAddress instead of contractID

	// Retrieve the VMExecutionEnvironment for the contract
	env, ok := ExecutionEnvironments[contractAddress]
	if !ok {
		http.Error(w, "VMExecutionEnvironment not found for contract", http.StatusNotFound)
		return
	}

	// Read the JSON body containing the functionSignature
	var requestBody struct {
		FunctionSignature string   `json:"functionSignature"`
		Args              []string `json:"args,omitempty"`
	}
	err := json.NewDecoder(r.Body).Decode(&requestBody)
	if err != nil {
		http.Error(w, "Error decoding request body", http.StatusBadRequest)
		return
	}

	functionSignature := requestBody.FunctionSignature
	if functionSignature == "" {
		http.Error(w, "Missing 'functionSignature' in request body", http.StatusBadRequest)
		return
	}

	parsedAbi, err := abi.JSON(bytes.NewReader(env.ABI))
	if err != nil {
		http.Error(w, "Error parsing contract ABI", http.StatusInternalServerError)
		return
	}

	method, ok := parsedAbi.Methods[functionSignature]
	if !ok {
		http.Error(w, "Function signature not found in ABI", http.StatusBadRequest)
		return
	}

	args := make([]interface{}, len(requestBody.Args))
	for i, arg := range requestBody.Args {
		argType := method.Inputs[i].Type

		var typedArg interface{}
		switch argType.T {
		case abi.IntTy, abi.UintTy:
			bigIntValue, ok := new(big.Int).SetString(arg, 10)
			if !ok {
				http.Error(w, "Error converting string to big.Int", http.StatusBadRequest)
				return
			}
			typedArg = bigIntValue
		case abi.StringTy:
			typedArg = arg
		case abi.BytesTy:
			bytesArg, err := hexutil.Decode(arg)
			if err != nil {
				http.Error(w, "Error decoding hex string to bytes", http.StatusBadRequest)
				return
			}
			typedArg = bytesArg
		default:
			http.Error(w, "Unsupported argument type", http.StatusBadRequest)
			return
		}

		args[i] = typedArg
	}

	// Execute the contract function
	result, err := env.ExecuteWithArgs(functionSignature, args)

	if err != nil {
		log.Printf("Error executing contract function: %v\nFunction signature: %s\nArgs: %v\n", err, functionSignature, requestBody.Args)
		http.Error(w, "Error executing contract function", http.StatusInternalServerError)
		return
	}

	// Respond with the execution result
	respJSON, err := json.Marshal(map[string]interface{}{
		"result": result,
	})
	if err != nil {
		http.Error(w, "Error marshaling execution result", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(respJSON)
}

func routes(bc *blockchain.Blockchain) *mux.Router {
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/", rootHandler).Methods("GET")
	router.HandleFunc("/contracts", createContract).Methods("POST")
	router.HandleFunc("/contracts/{id}", getContractByID).Methods("GET")
	router.HandleFunc("/contracts/{id}/execute", executeContractFunction).Methods("POST")
	router.HandleFunc("/contracts", getContracts).Methods("GET")
	router.HandleFunc("/chain", getChainHandler).Methods("GET")
	router.HandleFunc("/transactions/new", createTransaction).Methods("POST")
	router.HandleFunc("/mine", mineHandler).Methods("GET")
	router.HandleFunc("/contracts/{id}/ricardian", getRicardianContractByID).Methods("GET")

	return router
}

// Add the following new functions

func getChainHandler(w http.ResponseWriter, r *http.Request) {
	allContracts, err := contracts.GetAllContracts()
	if err != nil {
		http.Error(w, "Error retrieving contracts", http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"chain":     bc.Chain(),
		"length":    len(bc.Chain()),
		"contracts": allContracts,
	}
	json.NewEncoder(w).Encode(data)
}

func getContracts(w http.ResponseWriter, r *http.Request) {
	allContracts, err := contracts.GetAllContracts()
	if err != nil {
		http.Error(w, "Error retrieving contracts", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(allContracts)
}

func createTransaction(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}

	var transaction blockchain.Transaction
	err = json.Unmarshal(body, &transaction)
	if err != nil {
		http.Error(w, "Error unmarshalling JSON", http.StatusBadRequest)
		return
	}

	bc.AddTransaction(&transaction)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(transaction)
}

func mineHandler(w http.ResponseWriter, r *http.Request) {
	bc.AddBlock()
	lastBlock := bc.LastBlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "New Block Forged",
		"block":   lastBlock,
	})
}

func getRicardianContractByID(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	contractAddress := vars["id"]

	// Use the existing GetContract function to retrieve the contract object
	contract, err := contracts.GetContract(contractAddress)
	if err != nil {
		http.Error(w, "Error retrieving contract", http.StatusInternalServerError)
		return
	}

	// Respond with the Ricardian contract as a plain text
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(contract.RicardianContract))
}
