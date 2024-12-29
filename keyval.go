package main

import (
	"sync"
)

// KV Represents Data Interaction
type KV struct {
	mu   sync.RWMutex
	data map[string][]byte
}

// KV Constructor
func NewKV() *KV {
	return &KV{
		data: map[string][]byte{},
	}
}

// Creates New KV Pair (rw lock)
func (kv *KV) Set(key []byte, val []byte) error {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	kv.data[string(key)] = val
	return nil
}

// Retrieves KV Pair (readlock)
func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	val, ok := kv.data[string(key)]
	return val, ok
}

// Deletes KV Pair (rw lock)
func (kv *KV) Del(key []byte) bool {
	kv.mu.Lock()
	defer kv.mu.Unlock()

	_, exists := kv.data[string(key)]
	if exists {
		delete(kv.data, string(key))
	}

	return exists
}
