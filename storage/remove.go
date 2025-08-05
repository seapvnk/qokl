package storage

import (
	"strings"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/v9/zygo"
	"github.com/seapvnk/qokl/parser"
)

// FnEntityDeleteAll return all entities that matches
// Lisp (deleteAll admin: (Fn [e] (and (> (hget %age) 22) (= (hget name) "Pedro"))))
func FnEntityDeleteAll(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
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

	count := int64(0)
	edb.View(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		query := makeTagQuery(tag.Name())
		for it.Seek(query); it.ValidForPrefix(query); it.Next() {
			item := it.Item()
			key := strings.Replace(string(item.Key()), string(query), "", int(1))
			err := deleteRowInQuery(env, key, predicate)
			if err == nil {
				count++
			}
		}
		return nil
	})

	return &zygo.SexpInt{Val: count}, nil
}

func deleteRowInQuery(env *zygo.Zlisp, key string, predicate *zygo.SexpFunction) error {
	entityHash := retrieveEntity(env, key)
	result, err := env.Apply(predicate, []zygo.Sexp{entityHash})
	if err == nil {
		result, isBool := result.(*zygo.SexpBool)
		if !isBool {
			return nil
		}

		if result.Val {
			return deleteEntity(key)
		}
	}

	return nil
}

// FnRelationship add tag to an entity
// Lisp (deleteEntity myEntity)
func FnDeleteEntity(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return parser.SignalWrongArgs()
	}

	objID := getEntityIDFromQuery(args[0])
	err := deleteEntity(objID)

	if err != nil {
		return parser.SignalErr(env, err)
	}

	return parser.SignalOk(env)
}

func deleteEntity(objID string) error {
	err := edb.Update(func(txn *badger.Txn) error {
		err := removeAllTags(txn, objID)
		err = removeAllRelationships(txn, objID)
		err = removeEntityFields(txn, objID)
		err = txn.Delete(makeEntityEntry(objID))
		return err
	})

	return err
}

func removeEntityFields(txn *badger.Txn, objID string) error {
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()
	query := makeEntityComponentQuery(objID)
	for it.Seek(query); it.ValidForPrefix(query); it.Next() {
		item := it.Item()
		err := txn.Delete(item.Key())
		if err != nil {
			return err
		}
	}

	return nil
}

func removeAllRelationships(txn *badger.Txn, objID string) error {
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()
	query := makeRelationshipTagQuery(objID)
	for it.Seek(query); it.ValidForPrefix(query); it.Next() {
		item := it.Item()
		rel := strings.Replace(string(item.Key()), string(query), "", int(1))
		err := removeRelationship(txn, rel, objID)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeRelationship(txn *badger.Txn, rel string, objID string) error {
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()
	query := makeRelationshipEntryOneSide(rel, objID)
	for it.Seek(query); it.ValidForPrefix(query); it.Next() {
		item := it.Item()
		targetID := strings.Replace(string(item.Key()), string(query), "", int(1))
		err := removeRelationshipWith(txn, rel, objID, targetID)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeRelationshipWith(txn *badger.Txn, rel string, e1 string, e2 string) error {
	var err error

	e1Side := makeRelationshipEntry(rel, e1, e2)
	if err = txn.Delete(e1Side); err != nil {
		return err
	}

	e2Side := makeRelationshipEntry(rel, e2, e1)
	if err = txn.Delete(e2Side); err != nil {
		return err
	}

	e1SideMeta := makeRelationshipMetaEntry(rel, e1, e2)
	if err = txn.Delete(e1SideMeta); err != nil {
		return err
	}

	e2SideMeta := makeRelationshipMetaEntry(rel, e2, e1)
	if err = txn.Delete(e2SideMeta); err != nil {
		return err
	}

	return nil
}

func removeAllTags(txn *badger.Txn, objID string) error {
	it := txn.NewIterator(badger.DefaultIteratorOptions)
	defer it.Close()
	query := makeTagEntryReverseEntity(objID)
	for it.Seek(query); it.ValidForPrefix(query); it.Next() {
		item := it.Item()
		tag := strings.Replace(string(item.Key()), string(query), "", int(1))
		err := removeTag(txn, tag, objID)
		if err != nil {
			return err
		}
	}

	return nil
}

func removeTag(txn *badger.Txn, tagName string, objID string) error {
	var err error

	tagReverseQuery := makeTagEntryReverse(tagName, objID)
	if err = txn.Delete(tagReverseQuery); err != nil {
		return err
	}

	tagQuery := makeTagEntry(tagName, objID)
	if err = txn.Delete(tagQuery); err != nil {
		return err
	}

	return nil
}
