package storage

import (
	"encoding/json"
	"errors"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/seapvnk/qokl/parser"
)

// FnEntityGet an entity at database
// Lisp (entity myEntity)
func FnEntityGet(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return parser.SignalWrongArgs()
	}

	objID := getEntityIDFromQuery(args[0])
	entityHash := retrieveEntity(env, objID)

	return entityHash, nil
}

// FnEntityRelationships fetch every which meet criterea
// Lisp (relationshipsOf myEntity are: %friends)
func FnEntityRelationships(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 3 {
		return parser.SignalWrongArgs()
	}

	objID := getEntityIDFromQuery(args[0])
	relType, relTypeOk := args[1].(*zygo.SexpSymbol)
	if !relTypeOk {
		return parser.SignalErr(env, errors.New("relation type must be a symbol"))
	}

	rel, relOk := args[2].(*zygo.SexpSymbol)
	if !relOk {
		return parser.SignalErr(env, errors.New("rel must be a symbol"))
	}

	rows := &zygo.SexpArray{}

	edb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		query := makeRelationshipEntryOneSide(rel.Name(), objID)
		for it.Seek(query); it.ValidForPrefix(query); it.Next() {
			item := it.Item()
			key := strings.Replace(string(item.Key()), string(query), "", int(1))
			skip := false
			item.Value(func(v []byte) error {
				skip = string(v) != relType.Name()
				return nil
			})

			if skip {
				continue
			}

			relMeta := getRelationshipMeta(env, txn, rel.Name(), objID, key)
			rows.Val = append(rows.Val, relMeta)
		}
		return nil
	})

	return rows, nil
}

func getRelationshipMeta(env *zygo.Zlisp, txn *badger.Txn, relName string, e1 string, e2 string) zygo.Sexp {
	item, err := txn.Get(makeRelationshipMetaEntry(relName, e1, e2))
	if err != nil {
		return zygo.SexpNull
	}

	var relationshipMeta zygo.Sexp

	err = item.Value(func(v []byte) error {
		var itemValue StoredValue
		err := json.Unmarshal(v, &itemValue)
		if err != nil {
			return err
		}
		relationshipMeta = parser.ToSexp(env, itemValue.Value)
		return nil
	})

	if err != nil {
		return zygo.SexpNull
	}

	return relationshipMeta
}

// FnEntitySelect return all entities that matches
// Lisp (select admin: (Fn [e] (and (> (hget %age) 22) (= (hget name) "Pedro"))))
func FnEntitySelect(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 2 {
		return parser.SignalWrongArgs()
	}

	tag, tagOk := args[0].(*zygo.SexpSymbol)
	if !tagOk {
		return parser.SignalWrongArgs()
	}

	predicate, predicateOk := args[1].(*zygo.SexpFunction)
	if !predicateOk {
		return parser.SignalWrongArgs()
	}

	rows := &zygo.SexpArray{}
	edb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		query := makeTagQuery(tag.Name())
		for it.Seek(query); it.ValidForPrefix(query); it.Next() {
			item := it.Item()
			key := strings.Replace(string(item.Key()), string(query), "", int(1))
			applyQueryOnRow(env, rows, key, predicate)
		}
		return nil
	})

	return rows, nil
}

func applyQueryOnRow(env *zygo.Zlisp, rows *zygo.SexpArray, key string, predicate *zygo.SexpFunction) {
	entityHash := retrieveEntity(env, key)
	result, err := env.Apply(predicate, []zygo.Sexp{entityHash})
	if err == nil {
		result, isBool := result.(*zygo.SexpBool)
		if !isBool {
			return
		}

		if result.Val {
			rows.Val = append(rows.Val, entityHash)
		}
	}
}

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
