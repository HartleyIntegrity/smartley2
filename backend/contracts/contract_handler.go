package contracts

type ContractHandler interface {
	DeployContract(contract *Contract) error
}
