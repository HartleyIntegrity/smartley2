package contracts

import (
	"errors"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/google/uuid"
)

const solcPath = `C:\\Users\\PC\\smartley-contracts\\backend\\api\\solc.exe`

func CompileSolidityString(source string) (string, error) {
	tmpFile, err := ioutil.TempFile("", "solidity")
	if err != nil {
		return "", err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(source)); err != nil {
		return "", err
	}
	if err := tmpFile.Close(); err != nil {
		return "", err
	}

	return CompileSolidityFile(tmpFile.Name())
}

func CompileSolidityFile(filename string) (string, error) {
	out, err := exec.Command(solcPath, filename, "--bin").CombinedOutput()
	if err != nil {
		return "", err
	}

	output := string(out)
	log.Printf("Solidity compiler output: %s", output)

	if strings.Contains(output, "Error") {
		return "", errors.New(output)
	}

	return output, nil
}

func GenerateContractAddress() string {
	return uuid.New().String()
}
