package core

import (
	"log"

	badger "github.com/dgraph-io/badger/v4"
)

var store *badger.DB

func OpenStore() {
	var err error
	db, err := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if err != nil {
		log.Fatal(err)
	}

	store = db
}

// Store module setup
func (vm *VM) UseStoreModule() *VM {
	vm.environment.AddFunction("dispatch", fnDispatch)

	return vm.UseCacheModule()
}

func CloseStore() {
	store.Close()
}
