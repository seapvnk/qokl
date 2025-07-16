package core

import (
	"errors"
	"time"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/zygo"
)

// Store/Cache module setup
func (vm *VM) UseCacheModule() *VM {
	vm.environment.AddFunction("setCache", fnStoreSetCache)
	vm.environment.AddFunction("getCache", fnStoreGetCache)
	vm.environment.AddFunction("deleteCache", fnStoreDeleteCache)

	return vm
}

// fnStoreSetCache set a value on a key in cache.
// Lisp: (setCache "my-key" "my-value" 10)
func fnStoreSetCache(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 3 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	key, ok := args[0].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("setCache: first arg must be string")
	}

	value, ok := args[1].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("setCache: second arg must be string")
	}

	ttl, ok := args[2].(*zygo.SexpInt)
	if !ok {
		return zygo.SexpNull, errors.New("setCache: third arg must be int")
	}

	entry := badger.NewEntry([]byte("cache."+key.S), []byte(value.S))
	if ttl.Val > 0 {
		entry.WithTTL(time.Second * time.Duration(ttl.Val))
	}

	err := store.Update(func(txn *badger.Txn) error {
		return txn.SetEntry(entry)
	})

	return zygo.SexpNull, err
}

// fnStoreGetCache retrieves a cached value, returns errror if not found or expired
// Lisp: (getCache "my-key")
func fnStoreGetCache(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	key, ok := args[0].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("getCache: first arg must be string")
	}
	var val string
	err := store.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("cache." + string(key.S)))
		if err != nil {
			return errors.New("key not found or expired")
		}

		var valCopy []byte
		valCopy, err = item.ValueCopy(nil)
		val = string(valCopy)
		return err
	})
	return toSexp(env, val), err
}

// DeleteCache removes a key from the cache
// fnStoreDeleteCache removes a key from the cache
// Lisp: (deleteCache "my-key")
func fnStoreDeleteCache(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	key, ok := args[0].(*zygo.SexpStr)
	if !ok {
		return zygo.SexpNull, errors.New("deleteCache: first arg must be string")
	}

	err := store.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("cache." + key.S))
	})

	return zygo.SexpNull, err
}
