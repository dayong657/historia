package storage

import (
	"encoding/json"
	"reflect"
	"testing"
)

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

func TestStats(t *testing.T) {
	// just make sure stats doesn't crash
	m := NewInMemoryStorage()
	m.Stats()

	m.Prepare([]byte{3}, []byte{1, 2})
	m.Commit([]byte{3})
	m.Stats()
}

func TestMemoryPrepareBad(t *testing.T) {
	m := NewInMemoryStorage()
	if m.Prepare([]byte{3}, []byte{1, 2}) != true {
		t.Fatal("Could not prepare a valid request")
	}

	if m.Prepare([]byte{3}, []byte{1, 2}) == true {
		t.Fatal("prepared something that was already prepared")
	}

}

func TestMemoryCommitBad(t *testing.T) {
	m := NewInMemoryStorage()
	if m.Commit([]byte{3}) == nil {
		t.Fatal("Committed a non-existant request")
	}

}

func TestMemoryMerge(t *testing.T) {
	expected := make(map[string]interface{})
	expected["a"] = "b"
	expected["b"] = "c"
	actual := make(map[string]interface{})

	inputResponses := [][]byte{
		[]byte(`{"a":"b"}`),
		[]byte(`{"b":"c"}`),
	}

	m := NewInMemoryStorage()

	result, _ := m.Merge([]byte{}, inputResponses)

	err := json.Unmarshal(result, &actual)
	if err != nil {
		t.Fatal("merge didn't produce valid JSON")
	}

	if !reflect.DeepEqual(expected, actual) {
		t.Fatal("didn't merge properly")
	}

	_, ok := m.Merge([]byte{}, [][]byte{[]byte("INVALID_JSON")})
	if ok {
		t.Fatal("merge accepted invalid JSON")
	}
}
