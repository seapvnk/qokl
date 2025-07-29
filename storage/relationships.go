package storage

import (
	"encoding/json"
	"errors"
	"fmt"
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
