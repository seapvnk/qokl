package storage

import (
	"encoding/json"
	"errors"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/seapvnk/qokl/parser"
)

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
