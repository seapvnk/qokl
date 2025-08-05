package core

import (
	"encoding/binary"
	"errors"
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
	"github.com/glycerine/zygomys/v9/zygo"
)

func queueKey(name string, index uint64) []byte {
	return []byte(fmt.Sprintf("queue.%s.%020d", name, index))
}

func metaKey(name, label string) []byte {
	return []byte(fmt.Sprintf("queue.%s.meta.%s", name, label))
}

func uint64ToBytes(n uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, n)
	return b
}

func bytesToUint64(b []byte) uint64 {
	return binary.BigEndian.Uint64(b)
}

// fnDispatch adds a message to a queue
// Lisp: (dispatch aQueue: (msgpack(hash key1: "value" key2: "value")))
func fnDispatch(env *zygo.Zlisp, name string, args []zygo.Sexp) (zygo.Sexp, error) {
	if len(args) != 2 {
		return zygo.SexpNull, zygo.WrongNargs
	}

	queueName, ok := args[0].(*zygo.SexpSymbol)
	if !ok {
		return zygo.SexpNull, errors.New("dispatch: first arg must be symbol")
	}

	value, ok := args[1].(*zygo.SexpRaw)
	if !ok {
		return zygo.SexpNull, errors.New("dispatch: second arg must serialized hash, use json function")
	}

	err := store.Update(func(txn *badger.Txn) error {
		// read tail
		var tail uint64
		item, err := txn.Get(metaKey(queueName.Name(), "tail"))
		if err == nil {
			val, _ := item.ValueCopy(nil)
			tail = bytesToUint64(val)
		}

		// write new entry
		if err := txn.Set(queueKey(queueName.Name(), tail), value.Val); err != nil {
			return err
		}

		// increment tail
		return txn.Set(metaKey(queueName.Name(), "tail"), uint64ToBytes(tail+1))
	})

	return zygo.SexpNull, err
}

// Dequeue gets and deletes the oldest message
func StoreDequeue(queueName string) ([]byte, error) {
	var value []byte

	err := store.Update(func(txn *badger.Txn) error {
		// read head
		var head uint64
		item, err := txn.Get(metaKey(queueName, "head"))
		if err == nil {
			val, _ := item.ValueCopy(nil)
			head = bytesToUint64(val)
		}

		key := queueKey(queueName, head)
		item, err = txn.Get(key)
		if err == badger.ErrKeyNotFound {
			return fmt.Errorf("queue %s is empty", queueName)
		} else if err != nil {
			return err
		}

		value, err = item.ValueCopy(nil)
		if err != nil {
			return err
		}

		if err := txn.Delete(key); err != nil {
			return err
		}

		return txn.Set(metaKey(queueName, "head"), uint64ToBytes(head+1))
	})

	return value, err
}
