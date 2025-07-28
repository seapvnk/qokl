package storage

import (
	"log"
	"path/filepath"

	badger "github.com/dgraph-io/badger/v4"
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

func OpenDB(baseDir string) string {
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
	return absStoragePath
}

func CloseDB() {
	edb.Close()
}

type StoredValue struct {
	Value any `json:"value"`
}
