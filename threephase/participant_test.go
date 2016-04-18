package threephase

import (
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/josephlewis42/historia/storage"
)

var (
	transactionId = "transactionid"
	transaction   = ThreePhaseTransaction{
		Peers:         testHosts,
		Data:          "data",
		TransactionID: transactionId,
	}

	encodedTransaction = mustMarshal(transaction)
)

// utility function for marshaling transactions or panicing
func mustMarshal(transaction ThreePhaseTransaction) []byte {
	data, err := json.Marshal(transaction)

	if err != nil {
		panic("Could not marshal transaction")
	}
	return data
}

func TestInitTransaction(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	if tpc.InitializeTransaction([]byte{'a'}) {
		t.Error("Initialized an invalid transaction")
	}

	if !tpc.InitializeTransaction(encodedTransaction) {
		t.Error("Could not initialize a valid transaction")
	}

	if tpc.InitializeTransaction(encodedTransaction) {
		t.Error("Initialized an existing transaction")
	}
}

func TestGetPeersTransaction(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm).(*threePhaseInternal)

	if tpc.getPeers("a") != nil {
		t.Error("Got peers for a non-existant transaction")
	}

	if !tpc.InitializeTransaction(encodedTransaction) {
		t.Fatal("Could not initialize the prep transaction")
	}

	if !reflect.DeepEqual(tpc.getPeers(transactionId), transaction.Peers) {
		t.Error("Peer sets from getpeers differed")
	}
}

func TestAbortTransaction(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	if tpc.Abort("") {
		t.Error("Aborted a non-existant transaction")
	}

	if !tpc.InitializeTransaction(encodedTransaction) {
		t.Fatal("Could not get a transaction to a non-abortable  initstate")
	}

	if !tpc.PreCommit(transactionId) {
		t.Fatal("Could not get a transaction to a non-abortable precommit state")
	}

	if !tpc.DoCommit(transactionId) {
		t.Fatal("Could not get a transaction to a non-abortable commit state")
	}

	if tpc.Abort(transactionId) {
		t.Error("Aborted a comitted transaction")
	}

	if tpc.CheckCommit(transactionId) != true {
		t.Error("the transaction changed from comitted to non-comitted")
	}

}

func TestAbortTransactionSuccess(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	if !tpc.InitializeTransaction(encodedTransaction) {
		t.Fatal("Could not get a transaction to a non-abortable  initstate")
	}

	if !tpc.Abort(transactionId) {
		t.Error("Could not abort the transaction")
	}
}

func TestPrecommitTransaction(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	if tpc.PreCommit("") {
		t.Error("PreCommitted a non-existant transaction")
	}

	if !tpc.InitializeTransaction(encodedTransaction) {
		t.Fatal("Could not get a transaction to an initstate")
	}

	if !tpc.PreCommit(transactionId) {
		t.Fatal("Transaction could not be precomitted")
	}

	if !tpc.DoCommit(transactionId) {
		t.Fatal("Could not commit the transaction")
	}

	if tpc.PreCommit(transactionId) {
		t.Error("precomitted a comitted transaction")
	}

	if tpc.CheckCommit(transactionId) != true {
		t.Error("the transaction changed from comitted to precomitted")
	}
}

func TestAutoCommit(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := newThreePhaseInternal(&fakeComm, db, &fakeComm)

	if tpc.autoCommit("foo") {
		t.Error("autocommitted a non-existant transaction")
	}

	if !tpc.InitializeTransaction(encodedTransaction) ||
		!tpc.PreCommit(transactionId) {
		t.Fatal("Could not get a transaction to an initstate")
	}

	time.Sleep(6)

	if tpc.CheckCommit(transactionId) != true {
		t.Error("the transaction was not auto-committed")
	}
}

func TestCommitInvalid(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	if tpc.DoCommit("") {
		t.Error("Committed a non-existant transaction")
	}

}

/**
func TestPrecommitTransaction(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)



/**
func TestInitTransaction(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	if tpc.InitializeTransaction([]byte{'a'}) {
		t.Error("Initialized an invalid transaction")
	}

	if !tpc.InitializeTransaction(encodedTransaction) {
		t.Error("Could not initialize a valid transaction")
	}

	if tpc.InitializeTransaction(encodedTransaction) {
		t.Error("Initialized an existing transaction")
	}
}**/
