package core

import (
	"encoding/binary"
	"fmt"

	badger "github.com/dgraph-io/badger/v4"
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

// Enqueue adds a message to the queue
func StoreEnqueue(queueName string, value []byte) error {
	return store.Update(func(txn *badger.Txn) error {
		// read tail
		var tail uint64
		item, err := txn.Get(metaKey(queueName, "tail"))
		if err == nil {
			val, _ := item.ValueCopy(nil)
			tail = bytesToUint64(val)
		}

		// write new entry
		if err := txn.Set(queueKey(queueName, tail), value); err != nil {
			return err
		}

		// increment tail
		return txn.Set(metaKey(queueName, "tail"), uint64ToBytes(tail+1))
	})
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
