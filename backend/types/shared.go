package types

type Stack []int64

func (s *Stack) Push(value int64) {
	*s = append(*s, value)
}

func (s *Stack) Pop() int64 {
	length := len(*s)
	if length == 0 {
		panic("Cannot pop from an empty stack") // or handle the case as needed
	}

	value := (*s)[length-1]
	*s = (*s)[:length-1]
	return value
}

func (s *Stack) Len() int {
	return len(*s)
}

type Memory []byte

type Storage map[string]interface{}

func (s Storage) GetABI() []byte {
	abi, ok := s["abi"]
	if ok {
		return abi.([]byte)
	}
	return nil
}

func (s Storage) SetABI(abi []byte) {
	s["abi"] = abi
}

type ExecutionEnvironment interface {
	Execute() error
}

func (s Storage) GetBytecode() []byte {
	bytecode, ok := s["bytecode"]
	if ok {
		return bytecode.([]byte)
	}
	return nil
}

func (s Storage) SetBytecode(bytecode []byte) {
	s["bytecode"] = bytecode
}

func (s *Stack) Peek(n int) int64 {
	length := len(*s)
	if length < n+1 {
		panic("stack underflow")
	}
	return (*s)[length-n-1]
}
