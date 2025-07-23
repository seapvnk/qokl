package storage

import (
	"encoding/json"
	"errors"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/seapvnk/qokl/parser"
)

func entityExists(txn *badger.Txn, objID string) bool {
	item, err := txn.Get(makeEntityEntry(objID))
	if err != nil {
		return false
	}

	err = item.Value(func(val []byte) error {
		if string(val) != "1" {
			return errors.New("")
		}

		return nil
	})

	return err == nil
}

func getEntityIDFromQuery(arg zygo.Sexp) string {
	var objID string

	// extract obj id
	switch val := arg.(type) {
	case *zygo.SexpStr:
		objID = val.S
	case *zygo.SexpHash:
		for _, pairList := range val.Map {
			for _, pair := range pairList {
				kstr := pair.Head.(*zygo.SexpSymbol)
				if kstr.Name() == "id" {
					value, ok := pair.Tail.(*zygo.SexpStr)
					if !ok {
						return ""
					}
					objID = value.S
					break
				}
			}
		}
	}

	return objID
}

func retrieveEntity(env *zygo.Zlisp, objID string) *zygo.SexpHash {
	// build entity hash
	entityHash := zygo.SexpHash{
		Map: make(map[int][]*zygo.SexpPair),
	}

	keysFound := 0
	edb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		query := makeEntityComponentQuery(objID)
		for it.Seek(query); it.ValidForPrefix(query); it.Next() {
			item := it.Item()
			key := strings.Replace(string(item.Key()), string(query), "", int(1))
			item.Value(func(v []byte) error {
				var itemValue StoredValue
				err := json.Unmarshal(v, &itemValue)
				if err != nil {
					return nil
				}
				keySexp := env.MakeSymbol(key)
				valSexp := parser.ToSexp(env, itemValue.Value)
				entityHash.HashSet(keySexp, valSexp)
				keysFound++
				return nil
			})
		}

		return nil
	})

	if keysFound != 0 {
		entityHash.HashSet(env.MakeSymbol("id"), parser.ToSexp(env, objID))
	}

	return &entityHash
}
