package storage

import (
	"encoding/json"
	"errors"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/v9/zygo"
	"github.com/google/uuid"
	"github.com/seapvnk/qokl/parser"
)

// FnAddTags add tag to an entity
// Lisp (addTag admin: myEntity)
func FnAddTag(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	objID := getEntityIDFromQuery(args[1])
	err := edb.Update(func(txn *badger.Txn) error {
		err := addTags(txn, env, objID, args[0])
		return err
	})

	if err != nil {
		return parser.SignalErr(env, zygo.WrongNargs)
	}

	return parser.SignalOk(env)
}

// addTag create an entry tag for an entity
func addTags(txn *badger.Txn, env *zygo.Zlisp, objID string, tagArg zygo.Sexp) error {
	if !entityExists(txn, objID) {
		return errors.New("entity does not exists")
	}

	switch tagArg := tagArg.(type) {
	case *zygo.SexpSymbol:
		tagName := tagArg.Name()
		txn.Set(makeTagEntry(tagName, objID), []byte("1"))
	case *zygo.SexpPair:
		pair := tagArg
		ok := true
		for ok {
			var sym *zygo.SexpSymbol
			sym, ok = pair.Head.(*zygo.SexpSymbol)
			if ok {
				tagName := sym.Name()
				txn.Set(makeTagEntry(tagName, objID), []byte("1"))
				txn.Set(makeTagEntryReverse(tagName, objID), []byte("1"))
			}

			pair, ok = pair.Tail.(*zygo.SexpPair)
		}
	}

	return nil
}

// FnEntityInsert insert an entity at database
// Lisp (insert %(admin user) name: "Pedro" age: 23)
func FnEntityInsert(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) < 3 || len(args)%2 == 0 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	var obj map[string]interface{}
	objID := uuid.NewString()

	err := edb.Update(func(txn *badger.Txn) error {
		// insert entry
		err := txn.Set(makeEntityEntry(objID), []byte("1"))
		if err != nil {
			return err
		}

		// add tags
		if errNotFound := addTags(txn, env, objID, args[0]); errNotFound != nil {
			return err
		}

		// store object keys
		obj = setComponents(txn, env, objID, args)

		return err
	})

	if err != nil {
		return zygo.SexpNull, err
	}

	obj["id"] = objID
	return parser.ToSexp(env, obj), nil
}

// setComponents insert/update components keys for a obj
func setComponents(txn *badger.Txn, env *zygo.Zlisp, objID string, args []zygo.Sexp) map[string]interface{} {
	obj := make(map[string]interface{})

	for i := 1; i < len(args)-1; i += 2 {
		key := args[i]
		value := args[i+1]
		keySym, ok := key.(*zygo.SexpSymbol)
		if ok {
			goVal, parserError := parser.SexpToGo(value)
			if parserError != nil {
				continue
			}

			data, err := json.Marshal(StoredValue{
				Value: goVal,
			})

			if err != nil {
				continue
			}

			err = txn.Set(makeEntityComponentEntry(keySym.Name(), objID), data)
			if err == nil {
				obj[keySym.Name()] = goVal
			}
		}
	}

	return obj
}
