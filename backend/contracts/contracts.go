package contracts

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"regexp"
	"smartley-contracts/storage"
	"smartley-contracts/types"
	"time"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"golang.org/x/crypto/sha3"
)

type VMExecutionEnvironment struct {
	Stack          types.Stack
	Memory         types.Memory
	Storage        types.Storage
	ProgramCounter int
	Bytecode       []byte
	ABI            []byte // Add this field
}

// NewVMExecutionEnvironment creates a new VMExecutionEnvironment for the given contract.
func NewVMExecutionEnvironment(contract *Contract) *VMExecutionEnvironment {
	// Validate the bytecode string and remove any non-hexadecimal characters.
	re := regexp.MustCompile(`[^0-9a-fA-F]`)
	cleanedBytecode := re.ReplaceAllString(contract.Bytecode, "")

	// Check if the cleaned bytecode has an odd length and left-pad with a zero if necessary.
	if len(cleanedBytecode)%2 != 0 {
		cleanedBytecode = "0" + cleanedBytecode
	}

	bytecode, err := hex.DecodeString(cleanedBytecode)
	if err != nil {
		log.Fatalf("Failed to decode bytecode: %v", err)
	}

	abiBytes := []byte(contract.ABI)

	// Initialize the VMExecutionEnvironment with the appropriate values.
	env := &VMExecutionEnvironment{
		// Set the required fields based on the given contract.
		Bytecode: bytecode, // Use the decoded bytecode
		ABI:      abiBytes, // Convert ABI to []byte
	}

	return env
}

func (env *VMExecutionEnvironment) ExecuteWithArgs(functionSignature string, args []interface{}) (interface{}, error) {
	// Parse the ABI from the environment
	parsedABI, err := abi.JSON(bytes.NewReader(env.ABI))
	if err != nil {
		return nil, fmt.Errorf("failed to parse ABI: %v", err)
	}

	// Encode the arguments according to the contract ABI
	encodedArgs, err := parsedABI.Pack(functionSignature, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to encode arguments: %v", err)
	}

	// Call the Execute function with the prepared contract bytecode and encoded arguments
	result, err := env.Execute(env.Bytecode, functionSignature, encodedArgs)
	return result, err
}

func (env *VMExecutionEnvironment) Execute(contractBytecode []byte, functionSignature string, encodedArgs []byte) (interface{}, error) {
	// Find the function selector
	selector, err := findFunctionSelector(env.ABI, functionSignature)
	if err != nil {
		return nil, err
	}

	// Decode the function selector
	selectorBytes, err := hex.DecodeString(selector)
	if err != nil {
		return nil, fmt.Errorf("unable to decode selector: %v", err)
	}

	// Concatenate the selector and encoded arguments
	inputData := append(selectorBytes, encodedArgs...)

	fmt.Printf("Execute: inputData: %x\n", inputData)

	// Execute the contract bytecode with the given input data
	returnValue, err := env.ExecuteBytecode(contractBytecode, inputData)
	return returnValue, err
}

func (env *VMExecutionEnvironment) ExecuteBytecode(contractBytecode []byte, inputData []byte) (interface{}, error) {
	pc := 0
	for pc < len(contractBytecode) {
		opCode := contractBytecode[pc]
		pc++
		switch opCode {

		case 0x00: // STOP
			return nil, nil

		case 0x01: // ADD
			x, y := env.Stack.Pop(), env.Stack.Pop()
			env.Stack.Push(x + y)

		case 0x02: // MUL
			x, y := env.Stack.Pop(), env.Stack.Pop()
			env.Stack.Push(x * y)

		// ... implement other arithmetic and logical opcodes ...

		case 0x35: // CALLDATALOAD
			index := env.Stack.Pop()
			indexInt := int(index)
			if err := checkSliceBounds(inputData, indexInt, indexInt+32); err != nil {
				return nil, err
			}
			data := big.NewInt(0).SetBytes(inputData[indexInt : indexInt+32])
			env.Stack.Push(data.Int64())

		case 0x36: // CALLDATASIZE
			env.Stack.Push(int64(len(inputData)))

		case 0x37: // CALLDATACOPY
			memOffset := env.Stack.Pop()
			dataOffset := env.Stack.Pop()
			length := env.Stack.Pop()

			memOffsetInt := int(memOffset)
			dataOffsetInt := int(dataOffset)
			lengthInt := int(length)

			if err := checkSliceBounds(env.Memory, memOffsetInt, memOffsetInt+lengthInt); err != nil {
				return nil, err
			}
			if err := checkSliceBounds(inputData, dataOffsetInt, dataOffsetInt+lengthInt); err != nil {
				return nil, fmt.Errorf("error in CALLDATACOPY: %w, dataOffset: %d, length: %d, inputData length: %d", err, dataOffsetInt, lengthInt, len(inputData))
			}
			copy(env.Memory[memOffsetInt:memOffsetInt+lengthInt], inputData[dataOffsetInt:dataOffsetInt+lengthInt])

		case 0x51: // MLOAD
			offset := env.Stack.Pop()
			if err := checkSliceBounds(env.Memory, int(offset), int(offset+32)); err != nil {
				return nil, err
			}
			value := big.NewInt(0).SetBytes(env.Memory[offset : offset+32])
			env.Stack.Push(value.Int64())

		case 0x52: // MSTORE
			offset, value := env.Stack.Pop(), env.Stack.Pop()
			if err := checkSliceBounds(env.Memory, int(offset), int(offset+32)); err != nil {
				return nil, err
			}
			binary.BigEndian.PutUint64(env.Memory[offset:offset+32], uint64(value))

		// ... implement other memory and storage opcodes ...

		case 0x60: // PUSH1
			if err := checkSliceBounds(contractBytecode, pc, pc+1); err != nil {
				return nil, err
			}
			value := int64(contractBytecode[pc])
			env.Stack.Push(value)
			pc++

		// ... implement other push opcodes ...

		case 0x80: // DUP1
			value := env.Stack.Peek(0)
			env.Stack.Push(value)

		// ... implement other dup opcodes ...

		case 0x90: // SWAP1
			x, y := env.Stack.Pop(), env.Stack.Pop()
			env.Stack.Push(x)
			env.Stack.Push(y)

		// ... implement other swap opcodes ...

		// ... implement other opcodes ...

		default:
			return nil, fmt.Errorf("unknown opcode: 0x%x", opCode)
		}
	}

	return nil, fmt.Errorf("execution reached end of bytecode without encountering STOP or RETURN")
}

func checkSliceBounds(slice []byte, start int, end int) error {
	if start < 0 || start >= len(slice) || end < 0 || end > len(slice) {
		return fmt.Errorf("slice bounds out of range [%d:%d] with length %d", start, end, len(slice))
	}
	return nil
}

// Add other functions and types in the contracts package here

type Contract struct {
	ID                string    `json:"id"`
	SoliditySource    string    `json:"solidity_source"`
	Address           string    `json:"address"`
	ABI               []byte    `json:"abi"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
	Name              string    `json:"name"`
	Source            string    `json:"source"`
	Bytecode          string    `json:"bytecode"`
	RicardianContract string    `json:"ricardianContract,omitempty"`
}

// findFunctionSelector finds the function selector for a given function signature in the ABI.
func findFunctionSelector(abiBytes []byte, functionSignature string) (string, error) {
	var abi []map[string]interface{}
	err := json.Unmarshal(abiBytes, &abi)
	if err != nil {
		return "", fmt.Errorf("failed to unmarshal ABI: %v", err)
	}

	functionFound := false
	var selector string
	for _, item := range abi {
		if item["type"].(string) == "function" && item["name"].(string) == functionSignature {
			functionFound = true
			inputs := item["inputs"].([]interface{})
			signature := functionSignature + "("
			for i, input := range inputs {
				inputMap := input.(map[string]interface{})
				signature += inputMap["type"].(string)
				if i < len(inputs)-1 {
					signature += ","
				}
			}
			signature += ")"
			selector = Keccak256(signature)[:8]
			break
		}
	}

	if !functionFound {
		return "", fmt.Errorf("function not found in ABI: %s", functionSignature)
	}

	return selector, nil
}

// Keccak256 returns a keccak256 hash of the given string.
func Keccak256(s string) string {
	hasher := sha3.NewLegacyKeccak256()
	hasher.Write([]byte(s))
	return hex.EncodeToString(hasher.Sum(nil))
}

func GetAllContracts() ([]*Contract, error) {
	var contracts []*Contract
	err := storage.DB.All(&contracts)
	return contracts, err
}

func GetContract(id string) (*Contract, error) {
	var contract Contract
	err := storage.DB.One("ID", id, &contract)
	return &contract, err
}

type CompiledContract struct {
	ABI      interface{} `json:"abi"`
	Bytecode string      `json:"bytecode"`
}

func CompileSoliditySource(soliditySource string) ([]byte, string, error) {
	apiUrl := "http://localhost:4000/compile"

	reqBody, err := json.Marshal(map[string]string{
		"source": soliditySource,
	})
	if err != nil {
		return nil, "", err
	}

	resp, err := http.Post(apiUrl, "application/json", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	respBody, err := ioutil.ReadAll(resp.Body)
	fmt.Println("Response JSON:")
	fmt.Println(string(respBody))
	if err != nil {
		return nil, "", err
	}

	if resp.StatusCode != http.StatusOK {
		var errorResp map[string]string
		json.Unmarshal(respBody, &errorResp)
		return nil, "", errors.New(errorResp["error"])
	}

	var compiledOutput map[string]map[string]interface{}
	err = json.Unmarshal(respBody, &compiledOutput)
	if err != nil {
		return nil, "", err
	}

	var abi interface{}
	var bytecode string

	if simpleStorage, ok := compiledOutput["SimpleStorage"]; ok {
		if abiData, ok := simpleStorage["abi"]; ok {
			abi = abiData
		}
		if evmData, ok := simpleStorage["evm"].(map[string]interface{}); ok {
			if bytecodeData, ok := evmData["bytecode"].(map[string]interface{}); ok {
				bytecode = bytecodeData["object"].(string)
			}
		}
	}

	if abi == nil || len(bytecode) == 0 {
		return nil, "", errors.New("failed to extract ABI and bytecode from compiled contract")
	}

	abiBytes, err := json.Marshal(abi)
	if err != nil {
		return nil, "", errors.New("failed to marshal ABI into byte slice")
	}

	return abiBytes, bytecode, nil
}

func CreateContract(contract *Contract, handler ContractHandler) (interface{}, error) {
	// Compile the Solidity source code to get ABI and bytecode
	abiBytes, bytecode, err := CompileSoliditySource(contract.SoliditySource)
	if err != nil {
		return nil, err
	}

	// Set the contract ABI and bytecode
	contract.ABI = abiBytes
	contract.Bytecode = bytecode

	err = storage.DB.Save(contract)
	if err != nil {
		return nil, err
	}

	err = handler.DeployContract(contract)
	if err != nil {
		return nil, err
	}

	return contract, nil
}

func UpdateContract(contract *Contract) (interface{}, error) {
	err := storage.DB.Update(contract)
	if err != nil {
		return nil, err
	}
	return contract, nil
}

func DeleteContract(id string) (interface{}, error) {
	var contract Contract
	err := storage.DB.One("ID", id, &contract)
	if err != nil {
		return nil, err
	}

	err = storage.DB.DeleteStruct(&contract)
	if err != nil {
		return nil, err
	}
	return contract, nil
}
