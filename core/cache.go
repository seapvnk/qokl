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
// Lisp: (setCache %mykey 10 (msgpack(hash(msg: "value"))))
func fnStoreSetCache(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 3 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	key, ok := args[0].(*zygo.SexpSymbol)
	if !ok {
		return zygo.SexpNull, errors.New("setCache: first arg must be symbol")
	}

	ttl, ok := args[1].(*zygo.SexpInt)
	if !ok {
		return zygo.SexpNull, errors.New("setCache: third arg must be int")
	}

	value, ok := args[2].(*zygo.SexpRaw)
	if !ok {
		return zygo.SexpNull, errors.New("setCache: second arg must be raw bytes")
	}

	err := store.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte("cache."+key.Name()), value.Val)
		if ttl.Val > 0 {
			entry.WithTTL(time.Second * time.Duration(ttl.Val))
		}

		return txn.SetEntry(entry)
	})

	return zygo.SexpNull, err
}

// fnStoreGetCache retrieves a cached value, returns nil if not found or expired
// Lisp: (getCache %mykey)
func fnStoreGetCache(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 1 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	key, ok := args[0].(*zygo.SexpSymbol)
	if !ok {
		return zygo.SexpNull, errors.New("getCache: first arg must be symbol")
	}
	var val []byte
	err := store.View(func(txn *badger.Txn) error {
		item, err := txn.Get([]byte("cache." + key.Name()))
		if err != nil {
			return errors.New("key not found or expired")
		}

		val, err = item.ValueCopy(nil)
		return err
	})

	if err != nil {
		return zygo.SexpNull, nil
	}

	return toSexp(env, val), nil
}

// fnStoreDeleteCache removes a key from the cache
// Lisp: (deleteCache %mykey)
func fnStoreDeleteCache(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	key, ok := args[0].(*zygo.SexpSymbol)
	if !ok {
		return zygo.SexpNull, errors.New("deleteCache: first arg must be symbol")
	}

	err := store.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte("cache." + key.Name()))
	})

	return zygo.SexpNull, err
}
