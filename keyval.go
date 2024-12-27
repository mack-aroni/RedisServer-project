package main

import "sync"

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

	kv.data[string(key)] = []byte(val)
	return nil
}

// Retrieves KV Pair (readlock)
func (kv *KV) Get(key []byte) ([]byte, bool) {
	kv.mu.RLock()
	defer kv.mu.RUnlock()

	val, ok := kv.data[string(key)]
	return val, ok
}
