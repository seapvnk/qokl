package storage

import (
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/seapvnk/qokl/parser"
)

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
			appendQueryToRow(env, rows, key, predicate)
		}
		return nil
	})

	return rows, nil
}

func appendQueryToRow(env *zygo.Zlisp, rows *zygo.SexpArray, key string, predicate *zygo.SexpFunction) {
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
