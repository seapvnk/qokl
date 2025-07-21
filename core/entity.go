package core

import (
	"encoding/json"
	"log"
	"path/filepath"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
	"github.com/google/uuid"
)

/*
* # Entity
*
* ## entity in storage:
* entities.entityid // pk
* components.entityid.component // value of component
* rels.relname.entityid1.entityid2 // relationship data
* tags.entityid.tagname // tag flag for entity
*
* ## entity api:
* (insert %(admin user) name: "Pedro" age: 23)
* (tag admin: myEntity) // can be entity id or entity hash (with id key inside)
* (entity myEntity) // can be entity id or entity hash (with id key inside), get exact entity by id
* (relationship myEntity yourEntity are: %friends %(for 10 years)) // are for both sides, belongs <-, has ->
* (relationOf myEntity yourEntity) // fetch all relationships between these two
* (relationsOf myEntity %friends are: %(for 10 years) has: %(meet years ago)) // fetch every which meet criteraa
* (select admin: (fn [e] (and (> (hget %age) 22) (= (hget %name) "Pedro"))))
* (delete admin: (fn [e] (and (> (hget %age) 22) (= (hget %name) "Pedro"))))
* (update admin:
        (fn [e]
            (begin (hset e %age (+ 1 (e %age)))
                   (newEntity))
        (fn [e]
            (and (> (hget e %age) 22) (= (hget e %name) "Pedro")))))
*/

var edb *badger.DB

func OpenDB(baseDir string) {
	var err error
	storagePath := filepath.Join(baseDir, "/.storage")
	absStoragePath, errFile := filepath.Abs(storagePath)
	if errFile != nil {
		log.Fatal(errFile)
	}

	db, err := badger.Open(badger.DefaultOptions(absStoragePath))
	if err != nil {
		log.Fatal(err)
	}

	edb = db
}

// Entity module setup
func (vm *VM) UseEntityModule() *VM {
	vm.environment.AddFunction("insert", fnEntityInsert)
	vm.environment.AddFunction("entity", fnEntityGet)
	vm.environment.AddFunction("select", fnEntitySelect)

	return vm
}

func CloseDB() {
	edb.Close()
}

type StoredValue struct {
	Value any `json:"value"`
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

// fnEntityGet an entity at database
// Lisp (entity myEntity)
func fnEntityGet(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	objID := getEntityIDFromQuery(args[0])
	entityHash := retrieveEntity(env, objID)

	return entityHash, nil
}

// fnEntitySelect return all entities that matches
// Lisp (select admin: (fn [e] (and (> (hget %age) 22) (= (hget name) "Pedro"))))
func fnEntitySelect(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	tag, tagOk := args[0].(*zygo.SexpSymbol)
	if !tagOk {
		return zygo.SexpNull, zygo.WrongNargs
	}

	predicate, predicateOk := args[1].(*zygo.SexpFunction)
	if !predicateOk {
		return zygo.SexpNull, zygo.WrongNargs
	}

	rows := &zygo.SexpArray{}

	edb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		query := makeTagQuery(tag.Name())
		for it.Seek(query); it.ValidForPrefix(query); it.Next() {
			item := it.Item()
			key := strings.Replace(string(item.Key()), string(query), "", int(1))

			entityHash := retrieveEntity(env, key)
			result, err := env.Apply(predicate, []zygo.Sexp{entityHash})
			if err == nil {
				result, isBool := result.(*zygo.SexpBool)
				if isBool && result.Val {
					rows.Val = append(rows.Val, entityHash)
				}
			} else {
				log.Print(err.Error())
			}
		}
		return nil
	})

	return rows, nil
}
