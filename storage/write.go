package storage

import (
	"encoding/json"
	"errors"

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

	entities := args[:1]

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
	relData := parser.ToSexp(env, true)
	
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
func addRelationship(txn *badger.Txn, entityIDs []zygo.Sexp, relType string, rel string, relData zygo.Sexp) error {
	e1Sexp, e1Ok := entityIDs[0].(*zygo.SexpStr)
	if !e1Ok {
		return errors.New("type error in add relationship")
	}

	e2Sexp, e2Ok := entityIDs[1].(*zygo.SexpStr)
	if !e2Ok {
		return errors.New("type error in add relationship")
	}

	e1, e2 := e1Sexp.S, e2Sexp.S
	relBytes := []byte(rel)
	var (
		entry1 *badger.Entry
		entry1Meta *badger.Entry
		entry2 *badger.Entry
		entry2Meta *badger.Entry
	)

	value, valueOk := relData.(*zygo.SexpHash)
	if !valueOk {
		return errors.New("error parsing relationship data")
	}

	goVal, parserError := parser.SexpToGo(value)
	if parserError != nil {
		return parserError
	}

	data, err := json.Marshal(StoredValue{
		Value: goVal,
	})

	if err != nil {
		return err
	}

	switch rel {
	case "belongs":
		entry1 = badger.NewEntry(makeRelationshipEntry("belongs", e1, e2), relBytes)
		entry2 = badger.NewEntry(makeRelationshipEntry("has", e1, e2), relBytes)
		entry1Meta = badger.NewEntry(makeRelationshipMetaEntry("belongs", e1, e2), data)
		entry2Meta = badger.NewEntry(makeRelationshipMetaEntry("has", e1, e2), data)
	case "has":
		entry1 = badger.NewEntry(makeRelationshipEntry("has", e1, e2), relBytes)
		entry2 = badger.NewEntry(makeRelationshipEntry("belongs", e1, e2), relBytes)
		entry1Meta = badger.NewEntry(makeRelationshipMetaEntry("has", e1, e2), data)
		entry2Meta = badger.NewEntry(makeRelationshipMetaEntry("belongs", e1, e2), data)
	case "are":
		entry1 = badger.NewEntry(makeRelationshipEntry("are", e1, e2), relBytes)
		entry2 = badger.NewEntry(makeRelationshipEntry("are", e1, e2), relBytes)
		entry1Meta = badger.NewEntry(makeRelationshipMetaEntry("are", e1, e2), data)
		entry2Meta = badger.NewEntry(makeRelationshipMetaEntry("are", e1, e2), data)
	default:
		return errors.New("undefined relationship type")
	}

	txn.SetEntry(entry1)
	txn.SetEntry(entry2)
	txn.SetEntry(entry1Meta)
	txn.SetEntry(entry2Meta)

	return nil
}

// FnAddTags add tag to an entity
// Lisp (addTag admin: myEntity)
func FnAddTag(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	objIDSexp, ok := args[1].(*zygo.SexpStr)
	if !ok {
		return parser.SignalErr(env, zygo.WrongNargs)
	}

	objID := objIDSexp.S
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
