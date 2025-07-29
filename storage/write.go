package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/google/uuid"
	"github.com/seapvnk/qokl/parser"
)

// FnRelationship add tag to an entity
// Lisp (relationship myEntity yourEntity are: %friends %(for 10 years)) // are for both sides, belongs <-, has ->
func FnRelationship(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) < 4 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	e1 := getEntityIDFromQuery(args[0])
	e2 := getEntityIDFromQuery(args[1])
	entities := []string{e1, e2}

	// parse rel type
	relTypeSexpSym, relTypeOk := args[2].(*zygo.SexpSymbol)
	if !relTypeOk {
		return parser.SignalErr(env, errors.New("incorrect relationship type, it should be a symbol"))
	}
	relType := relTypeSexpSym.Name()

	// parse rel
	relSexpSym, relOk := args[3].(*zygo.SexpSymbol)
	if !relOk {
		return parser.SignalErr(env, errors.New("incorrect relationship def, it should be a symbol"))
	}
	rel := relSexpSym.Name()

	// parse rel data
	var relData zygo.Sexp
	relData = zygo.SexpNull
	if len(args) >= 5 {
		relData = args[4]
	}

	err := edb.Update(func(txn *badger.Txn) error {
		return addRelationship(txn, entities, relType, rel, relData)
	})

	if err != nil {
		return parser.SignalErr(env, err)
	}

	return parser.SignalOk(env)
}

// addRelationship add relationship between two entities
func addRelationship(txn *badger.Txn, entityIDs []string, relType string, rel string, relData zygo.Sexp) error {
	e1, e2 := entityIDs[0], entityIDs[1]
	var (
		entry1     *badger.Entry
		entry1Meta *badger.Entry
		entry2     *badger.Entry
		entry2Meta *badger.Entry
	)

	var (
		data []byte
		err  error
	)
	switch value := relData.(type) {
	case *zygo.SexpSentinel:
		data, err = json.Marshal(StoredValue{
			Value: nil,
		})

		if err != nil {
			return err
		}
	default:
		goVal, parserError := parser.SexpToGo(value)
		if parserError != nil {
			return parserError
		}

		data, err = json.Marshal(StoredValue{
			Value: goVal,
		})

		if err != nil {
			return err
		}
	}

	switch relType {
	case "belongs":
		entry1 = badger.NewEntry(makeRelationshipEntry(rel, e1, e2), []byte("belongs"))
		entry2 = badger.NewEntry(makeRelationshipEntry(rel, e2, e1), []byte("has"))
		entry1Meta = badger.NewEntry(makeRelationshipMetaEntry(rel, e1, e2), data)
		entry2Meta = badger.NewEntry(makeRelationshipMetaEntry(rel, e2, e1), data)
	case "has":
		entry2 = badger.NewEntry(makeRelationshipEntry(rel, e2, e1), []byte("belongs"))
		entry1 = badger.NewEntry(makeRelationshipEntry(rel, e1, e2), []byte("has"))
		entry2Meta = badger.NewEntry(makeRelationshipMetaEntry(rel, e2, e1), data)
		entry1Meta = badger.NewEntry(makeRelationshipMetaEntry(rel, e1, e2), data)
	case "are":
		entry1 = badger.NewEntry(makeRelationshipEntry(rel, e1, e2), []byte("are"))
		entry2 = badger.NewEntry(makeRelationshipEntry(rel, e2, e1), []byte("are"))
		entry1Meta = badger.NewEntry(makeRelationshipMetaEntry(rel, e1, e2), data)
		entry2Meta = badger.NewEntry(makeRelationshipMetaEntry(rel, e2, e1), data)
	default:
		return fmt.Errorf("undefined relationship type: %s", rel)
	}

	txn.SetEntry(entry1)
	txn.SetEntry(entry2)
	txn.SetEntry(entry1Meta)
	txn.SetEntry(entry2Meta)

	txn.Set(makeRelationshipTagEntry(rel, e1), []byte("1"))
	txn.Set(makeRelationshipTagEntry(rel, e2), []byte("1"))

	return nil
}

// FnEntityDeleteAll return all entities that matches
// Lisp (updateAll admin: (fn [e] (begin ...) e) (fn [e] (and (> (hget %age) 22) (= (hget name) "Pedro"))))
func FnEntityUpdateAll(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 3 {
		return parser.SignalWrongArgs()
	}

	tag, tagOk := args[0].(*zygo.SexpSymbol)
	if !tagOk {
		return parser.SignalWrongArgs()
	}

	mapfn, mapfnOk := args[1].(*zygo.SexpFunction)
	if !mapfnOk {
		return parser.SignalWrongArgs()
	}

	predicate, predicateOk := args[2].(*zygo.SexpFunction)
	if !predicateOk {
		return parser.SignalWrongArgs()
	}

	count := int64(0)
	edb.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		query := makeTagQuery(tag.Name())
		for it.Seek(query); it.ValidForPrefix(query); it.Next() {
			item := it.Item()
			key := strings.Replace(string(item.Key()), string(query), "", int(1))
			updated, errUpdate := updateRowInQuery(env, txn, key, mapfn, predicate)

			if errUpdate != nil {
				return errUpdate
			}
			if updated {
				count++
			}
		}
		return nil
	})

	return &zygo.SexpInt{Val: count}, nil
}

func updateRowInQuery(env *zygo.Zlisp, txn *badger.Txn, key string, mapfn *zygo.SexpFunction, predicate *zygo.SexpFunction) (bool, error) {
	entityHash := retrieveEntity(env, key)
	result, err := env.Apply(predicate, []zygo.Sexp{entityHash})
	if err == nil {
		result, isBool := result.(*zygo.SexpBool)
		if !isBool {
			return false, nil
		}

		if result.Val {
			applyHash, errApply := env.Apply(mapfn, []zygo.Sexp{entityHash})
			if errApply != nil {
				return false, errApply
			}

			resultHash, isHash := applyHash.(*zygo.SexpHash)
			if !isHash {
				return false, errors.New("update function should always return a hash")
			}

			var setParams []zygo.Sexp
			setParams = append(setParams, zygo.SexpNull)
			numPairs := zygo.HashCountKeys(resultHash)
			for i := 0; i < numPairs; i++ {
				pair, err := resultHash.HashPairi(i)
				if err != nil {
					continue
				}

				hashkey, ok := pair.Head.(*zygo.SexpSymbol)
				if !ok {
					continue
				}

				if hashkey.Name() == "id" {
					continue
				}

				hashval, okVal := pair.Tail.(*zygo.SexpPair)
				if !okVal {
					continue
				}

				setParams = append(setParams, hashkey)
				setParams = append(setParams, hashval.Head)
			}

			setComponents(txn, env, key, setParams)
		}
	}

	return true, nil
}

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
