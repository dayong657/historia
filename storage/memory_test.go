package storage

import "testing"

var (
	testKey   = []byte("test")
	testValue = []byte("value")
)

func TestMemoryRead(t *testing.T) {
	m := NewInMemoryStorage()

	data, ok := m.Read(testKey)
	if ok {
		t.Errorf("Memory should have been empty for key: 'test', instead got value: %s\n", string(data))
	}

	m.Prepare(testKey, testValue)
	m.Commit(testKey)

	data, ok = m.Read(testKey)
	if !ok {
		t.Errorf("Memory should have been stored for key: 'test', instead it was empty\n")
	}
}

func TestMemoryCommit(t *testing.T) {
	m := NewInMemoryStorage()

	data, ok := m.Read(testKey)
	if ok {
		t.Errorf("Memory should have been empty for key: 'test', instead got value: %s\n", string(data))
	}

	m.Prepare(testKey, testValue)
	m.Commit(testKey)

	data, ok = m.Read(testKey)
	if !ok {
		t.Errorf("Memory should have been stored for key: 'test', instead it was empty\n")
	}
}

func TestMemoryAbort(t *testing.T) {
	m := NewInMemoryStorage()

	data, ok := m.Read(testKey)
	if ok {
		t.Errorf("Memory should have been empty for key: 'test', instead got value: %s\n", string(data))
	}

	m.Prepare(testKey, testValue)
	m.Abort(testKey)

	data, ok = m.Read(testKey)
	if ok {
		t.Errorf("Memory should have been empty after abort for key: 'test', instead got value: %s\n", string(data))
	}

	ok = m.Abort(testValue)
	if ok {
		t.Errorf("Memory should not be able to abort a non-existing transaction\n")
	}
}

func TestMemoryMerge(t *testing.T) {

}

/**
type Storage interface {
	Read(request []byte) (value []byte, ok bool)
	Commit(transactionID []byte) error
	Prepare(transactionID, value []byte) bool
	Abort(transactionID []byte) bool
	Merge(request []byte, response [][]byte) (result []byte, ok bool)
}

*/
