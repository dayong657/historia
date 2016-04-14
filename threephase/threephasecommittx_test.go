package threephase

import (
	"errors"
	"log"
	"testing"
	"time"

	"github.com/josephlewis42/historia/storage"
)

var (
	testHosts = []string{"host1", "host2"}
)

func TestNewThreePhaseCommit(t *testing.T) {
	fakeComm := newFakeComm(testHosts)
	db := storage.NewInMemoryStorage()

	NewThreePhaseCommit(&fakeComm, db, &fakeComm)
}

func readN(n int, c <-chan string, timeoutms int) error {
	duration := time.Millisecond * 5000
	//log.Printf("Waiting for %d connections\n", n)

	for i := 1; i <= n; i++ {
		select {
		case _ = <-c:
			//log.Printf("Read %d of %d\n", i, n)
			continue
		case _ = <-time.After(duration):
			log.Printf("Timed out waiting for cxn %d of %d\n", i, n)

			return errors.New("Timed out processing transactions")
		}
	}

	return nil
}

func TestCommitTxNormal(t *testing.T) {
	singleHost := []string{"host", "two"}
	initc := make(chan string)
	precommitc := make(chan string)
	commitc := make(chan string)

	fakeComm := newFakeComm(singleHost)
	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	fakeComm.InitializeTransactionI = newHandlerCallback(nil, nil, initc)
	fakeComm.PreCommitI = newHandlerCallback(nil, nil, precommitc)
	fakeComm.DoCommitI = newHandlerCallback(nil, nil, commitc)

	go func() {
		result := tpc.CommitTx("tx", []byte{}, singleHost)

		if result == false {
			t.Fatal("Result was not success")
		}
	}()

	if err := readN(2, initc, 500); err != nil {
		t.Fatal(err)
	}

	if err := readN(2, precommitc, 500); err != nil {
		t.Fatal(err)
	}

	if err := readN(2, commitc, 500); err != nil {
		t.Fatal(err)
	}
}

func txTestHelper(t *testing.T, fakeComm fakeCommunicationHandler, hosts []string, expectedResult bool, channels map[string]chan string) {

	db := storage.NewInMemoryStorage()
	tpc := NewThreePhaseCommit(&fakeComm, db, &fakeComm)

	finished := make(chan bool)
	go func() {
		result := tpc.CommitTx("tx", []byte{}, hosts)

		if result != expectedResult {
			t.Errorf("Commit returned wrong result, expected %t got %t\n", expectedResult, result)
		}

		finished <- true
	}()

	for key, c := range channels {

		if err := readN(len(hosts), c, 500); err != nil {
			t.Errorf("Error when getting %s, err: %s\n", key, err)
		}
	}
	<-finished
}

func TestCommitTxInitError(t *testing.T) {
	abortc := make(chan string)
	fakeComm := newFakeComm(testHosts)

	fakeComm.InitializeTransactionI = newHandlerCallback([]string{"host1"}, nil, nil)
	fakeComm.AbortI = newHandlerCallback(nil, nil, abortc)

	txTestHelper(t,
		fakeComm,
		testHosts,
		false, // expected result
		map[string]chan string{"abort": abortc})
}

func TestCommitTxInitTimeout(t *testing.T) {
	abortc := make(chan string)
	fakeComm := newFakeComm(testHosts)

	fakeComm.InitializeTransactionI = newHandlerCallback(nil, []string{"host1"}, nil)
	fakeComm.AbortI = newHandlerCallback(nil, nil, abortc)

	txTestHelper(t,
		fakeComm,
		testHosts,
		false, // expected result
		map[string]chan string{"abort": abortc})
}

func TestCommitTxPrecommitTimeout(t *testing.T) {
	abortc := make(chan string)
	fakeComm := newFakeComm(testHosts)

	fakeComm.PreCommitI = newHandlerCallback(nil, []string{"host1"}, nil)
	fakeComm.AbortI = newHandlerCallback(nil, nil, abortc)

	txTestHelper(t,
		fakeComm,
		testHosts,
		false, // expected result
		map[string]chan string{"abort": abortc})
}

func TestCommitTxCommitTimeout(t *testing.T) {
	fakeComm := newFakeComm(testHosts)

	fakeComm.DoCommitI = newHandlerCallback(nil, []string{"host1"}, nil)

	txTestHelper(t,
		fakeComm,
		testHosts,
		false, // expected result
		map[string]chan string{})
}

//func (this *threePhaseInternal) CommitTx(transactionid string, data []byte, nodes []string) (success bool) {

/**
func TestTimeConsuming(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
}

type ThreePhaseCommit interface {
	Create(request []byte) (success bool)
	Read(request []byte) (results []byte, success bool)
	Update(request []byte) (success bool)
	Delete(request []byte) (success bool)

	CommitTx(transactionid string, data []byte, nodes []string) (success bool)

	// these methods are called by an external handler
	InitializeTransaction(transaction []byte) (ok bool)
	Abort(transactionID string) (ok bool)
	DoCommit(transactionID string) (ok bool)
	PreCommit(transactionID string) (ok bool)
	CheckCommit(transactionID string) (didcommit bool)
	LocalRead(request []byte) (results []byte, success bool)
}
**/
