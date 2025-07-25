package core

import (
	"fmt"
	"os"

	"github.com/glycerine/zygomys/zygo"
)

type ZygResult struct {
	Value zygo.Sexp
	Error error
}

type VM struct {
	environment *zygo.Zlisp
}

func NewVM() *VM {
	env := zygo.NewZlisp()
	return &VM{
		environment: env,
	}
}

func (vm *VM) AddVariables(variables map[string]any) {
	if variables != nil {
		for k, v := range variables {
			vm.environment.AddGlobal(k, toSexp(vm.environment, v))
		}
	}
}

func (vm *VM) Execute(path string) (*ZygResult, error) {
	code, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("could not read file %s: %w", path, err)
	}

	err = vm.environment.LoadString(string(code))
	if err != nil {
		return nil, fmt.Errorf("error executing %s: %w", path, err)
	}

	out, err := vm.environment.Run()
	return &ZygResult{Value: out, Error: err}, nil
}
