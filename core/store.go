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

func CloseStore() {
	store.Close()
}
