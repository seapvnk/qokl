package core

import (
	"encoding/json"
	"log"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/google/uuid"
)

/*
* # Entity
*
* ## entity in storage:
* entity.entityid // pk
* components.component.entityid // value of component
* rel.relname.entityid1.entityid2 // relationship data
* tag.tagname.entityid // tag flag for entity
*
* ## entity api:
* (insert %(admin user) name: "Pedro" age: 23)
* (tag admin: myEntity) // can be entity id or entity hash (with id key inside)
* (fetch admin: myEntity) // can be entity id or entity hash (with id key inside), get exact entity by id
* (relationship myEntity yourEntity are: %friends %(for 10 years)) // are for both sides, belongs <-, has ->
* (relationOf myEntity yourEntity) // fetch all relationships between these two
* (relationsOf myEntity %friends are: %(for 10 years) has: %(meet years ago)) // fetch every which meet criteraa
* (select admin: (fn [name age] (and (> age 22) (= name "Pedro")))) // args are injected by entities key in admin tag
* (delete admin: (fn [name age] (and (> age 22) (= name "Pedro"))))
* (update admin:
        (fn [newEntity]
            (begin (hset newEntity %age (+ 1 (hget %age)))
                   (newEntity))
        (fn [name age]
            (and (> age 22) (= name "Pedro")))))
*/

var edb *badger.DB

func OpenDB() {
	var err error
	db, err := badger.Open(badger.DefaultOptions("./.storage"))
	if err != nil {
		log.Fatal(err)
	}

	edb = db
}

// Entity module setup
func (vm *VM) UseEntityModule() *VM {
	vm.environment.AddFunction("insert", fnEntityInsert)

	return vm
}

func CloseDB() {
	edb.Close()
}

type StoredValue struct {
	Value any `json:"value"`
}

func makeEntityEntry(entityID string) []byte {
	return []byte("entities." + entityID)
}

func makeTagEntry(tagName string, entityID string) []byte {
	return []byte("tags." + entityID + "." + tagName)
}

func makeEntityComponentEntry(componentName string, entityID string) []byte {
	return []byte("components." + entityID + "." + componentName)
}

// fnEntityInsert insert an entity at database
// Lisp (insert %(admin user) name: "Pedro" age: 23)
func fnEntityInsert(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) < 3 || len(args)%2 == 0 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	objID := uuid.NewString()
	obj := make(map[string]interface{})

	err := edb.Update(func(txn *badger.Txn) error {
		// insert entry
		err := txn.Set(makeEntityEntry(objID), []byte("1"))
		if err != nil {
			return err
		}

		// add tags
		tagArg := args[0]
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

		// store object keys
		for i := 1; i < len(args)-1; i += 2 {
			key := args[i]
			value := args[i+1]
			keySym, ok := key.(*zygo.SexpSymbol)
			if ok {
				goVal, parserError := SexpToGo(value)
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

		return err
	})

	if err != nil {
		return zygo.SexpNull, err
	}

	obj["id"] = objID
	return toSexp(env, obj), nil
}
