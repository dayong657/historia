package storage

import (
	"encoding/json"
	"errors"
)

func NewInMemoryStorage() Storage {
	return &inMemoryStorage{backend: make(map[string][]byte),
		precommit: make(map[string][]byte)}
}

type inMemoryStorage struct {
	backend   map[string][]byte
	precommit map[string][]byte
}

func (store *inMemoryStorage) Read(key []byte) (value []byte, ok bool) {
	value, ok = store.backend[string(key)]
	return value, ok
}

func (store *inMemoryStorage) Commit(key []byte) error {

	value, found := store.precommit[string(key)]
	if !found {
		return errors.New("Error, no transaction exists for " + string(key))
	}

	store.backend[string(key)] = value
	delete(store.precommit, string(key))
	return nil
}

func (store *inMemoryStorage) Prepare(key, value []byte) bool {

	// make sure the transaction isn't already processing
	_, found := store.precommit[string(key)]
	if found {
		return false
	}

	store.precommit[string(key)] = value
	return true
}

func (store *inMemoryStorage) Abort(transactionID []byte) bool {
	_, found := store.precommit[string(transactionID)]
	if !found {
		return false
	}

	delete(store.precommit, string(transactionID))
	return true
}

func (store *inMemoryStorage) Merge(request []byte, response [][]byte) (result []byte, ok bool) {
	results := make(map[string]interface{})

	for _, data := range response {
		src := make(map[string]interface{})
		err := json.Unmarshal(data, &src)

		if err != nil {
			return nil, false
		}

		for k, v := range src {
			results[k] = v
		}
	}

	result, err := json.Marshal(results)
	return result, err != nil
}

func (store *inMemoryStorage) Stats() string {
	output := "In Memory Storage Statistics\n"

	for k, v := range store.backend {
		output += k + "\t" + string(v) + "\n"
	}

	return output
}
