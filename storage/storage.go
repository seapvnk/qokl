package storage

import (
	"log"
	"path/filepath"
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
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
* (entity myEntity) // get entity by id
* (remove myEntity) // delete entity by id
* (relationship myEntity yourEntity are: %friends %(for 10 years)) // are for both sides, belongs <-, has ->
* (relationOf myEntity yourEntity) // fetch all relationships between these two
* (relationsOf myEntity %friends are: %(for 10 years) has: %(meet years ago)) // fetch every which meet criteraa
* (select admin: (Fn [e] (and (> (hget %age) 22) (= (hget %name) "Pedro"))))
* (delete admin: (Fn [e] (and (> (hget %age) 22) (= (hget %name) "Pedro"))))
* (update admin:
        (Fn [e]
            (begin (hset e %age (+ 1 (e %age)))
                   (newEntity))
        (Fn [e]
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

func CloseDB() {
	edb.Close()
}

type StoredValue struct {
	Value any `json:"value"`
}

// FnEntityGet an entity at database
// Lisp (entity myEntity)
func FnEntityGet(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	objID := getEntityIDFromQuery(args[0])
	entityHash := retrieveEntity(env, objID)

	return entityHash, nil
}

// FnEntitySelect return all entities that matches
// Lisp (select admin: (Fn [e] (and (> (hget %age) 22) (= (hget name) "Pedro"))))
func FnEntitySelect(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
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
