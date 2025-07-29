package storage

import (
	"errors"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/seapvnk/qokl/parser"
)

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
