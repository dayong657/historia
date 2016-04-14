package threephase

import (
	"encoding/json"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/josephlewis42/historia/storage"
)

type Phase int

const (
	PhaseUncertain = 0
	PhasePrepared  = 1
	PhaseCommitted = 2
	PhaseAborted   = 3
)

var (
	PhaseTimeout = time.Second * 1
)

type ThreePhaseTransaction struct {
	Peers         []string
	Data          string
	TransactionID string
	status        Phase
}

type threePhaseInternal struct {
	comm             CommunicationHandler
	db               storage.Storage
	ch               NodeSet
	transactions     map[string]*ThreePhaseTransaction
	transactionslock sync.RWMutex
}

func (this *threePhaseInternal) Create(request []byte) (success bool) {
	transactionID := strconv.Itoa(int(time.Now().UnixNano()))
	nodes, err := this.ch.GetCreateSet()
	if err != nil {
		return false
	}

	return this.CommitTx(transactionID, request, nodes)
}

func (this *threePhaseInternal) Read(request []byte) (results []byte, success bool) {
	//transactionID := string(time.Now().UnixNano())
	//nodes, err := this.ch.GetReadSet()
	nodes, err := this.ch.GetReadSet()
	if err != nil {
		return nil, false
	}

	temporary := [][]byte{}

	for _, node := range nodes {
		result, err := this.comm.ReadData(request, node)

		if err != nil {
			log.Printf("Read: Got error while reading data from node: %s %s\n", node, err)
			return nil, false
		}

		temporary = append(temporary, result)
	}

	return this.db.Merge(request, temporary)
}

func (this *threePhaseInternal) LocalRead(request []byte) (results []byte, success bool) {
	return this.db.Read(request)
}

func (this *threePhaseInternal) Update(request []byte) (success bool) {
	transactionID := string(time.Now().UnixNano())
	nodes, err := this.ch.GetUpdateSet()
	if err != nil {
		return false
	}

	return this.CommitTx(transactionID, request, nodes)
}

func (this *threePhaseInternal) Delete(request []byte) (success bool) {
	transactionID := string(time.Now().UnixNano())
	nodes, err := this.ch.GetDeleteSet()
	if err != nil {
		return false
	}

	return this.CommitTx(transactionID, request, nodes)
}

func (this *threePhaseInternal) CommitTx(transactionid string, data []byte, nodes []string) (success bool) {
	log.Printf("Starting transaction %s\n", transactionid)
	transaction := ThreePhaseTransaction{
		Peers:         nodes,
		Data:          string(data),
		TransactionID: transactionid,
	}

	data, err := json.Marshal(transaction)

	if err != nil {
		log.Printf("CommitTx: error, could not marshal json %s\n", err)
		return false
	}

	log.Printf("Starting initial for transaction %s\n", transactionid)
	// initial
	if !allOkay(this.comm.InitializeTransaction, data, nodes) {
		log.Printf("Timed out waiting for init for transaction %s\n", transactionid)
		allOkay(this.comm.Abort, []byte(transactionid), nodes)
		return false
	}

	// precommit
	log.Printf("Starting precommit for transaction %s\n", transactionid)
	if !allOkay(this.comm.PreCommit, []byte(transactionid), nodes) {
		log.Printf("Timed out waiting for precommit for transaction %s\n", transactionid)

		allOkay(this.comm.Abort, []byte(transactionid), nodes)
		return false
	}

	// commit
	log.Printf("Starting commit for transaction %s\n", transactionid)
	return allOkay(this.comm.DoCommit, []byte(transactionid), nodes)
}

func allOkay(callback func(request []byte, destination string) (ok bool, err error), data []byte, nodes []string) bool {
	for _, node := range nodes {
		ok, err := callback(data, node)
		if err != nil {
			log.Printf("Threephase::allokay, got error: %s from node %s", err, node)
			return false
		}

		if !ok {
			log.Printf("Threephase::allokay, got not ok from node %s", node)
			return false
		}
	}

	return true
}

func okayCheck(callback func(request []byte, destination string) (ok bool, err error), data []byte, nodes []string) (numOkay, numNotOkay, numErr int) {

	for _, node := range nodes {
		ok, err := callback(data, node)
		if err != nil {
			numErr += 1
		}

		if ok {
			numOkay += 1
		} else {
			numNotOkay += 1
		}
	}

	return numOkay, numNotOkay, numErr
}

func (this *threePhaseInternal) InitializeTransaction(encodedTransaction []byte) (ok bool) {
	var tx ThreePhaseTransaction

	err := json.Unmarshal(encodedTransaction, &tx)

	if err != nil {
		log.Printf("InitializeTransaction: error decoding json %s, %s\n", err, string(encodedTransaction))
		return false
	}

	transactionid := tx.TransactionID

	this.transactionslock.Lock()
	defer this.transactionslock.Unlock()

	tx.status = PhaseUncertain

	// make sure the transaction hasn't already started
	_, found := this.transactions[transactionid]
	if found {
		log.Printf("InitializeTransaction, already have entry for transaction: %s\n", transactionid)
		return false
	}

	// Make sure the database wants to accept the transaction
	ok = this.db.Prepare([]byte(transactionid), []byte(tx.Data))
	if !ok {
		log.Printf("InitializeTransaction, database would not precommit")
		return false
	}

	this.transactions[transactionid] = &tx
	//go this.terminationProtocol(transactionid)
	// TODO check if anyone got precommit

	return true
}

func (this *threePhaseInternal) Abort(transactionID string) (ok bool) {
	this.transactionslock.Lock()
	defer this.transactionslock.Unlock()

	item, found := this.transactions[transactionID]

	if !found {
		log.Printf("Abort: the transaction with the ID %s couldn't be found\n", transactionID)
		return false
	}

	if item.status == PhaseCommitted {
		log.Printf("Abort: transaction %s was already comitted\n", transactionID)
		return false
	}

	// abort the data
	this.db.Abort([]byte(transactionID))
	item.status = PhaseAborted

	go this.autoCleanup(transactionID)
	return true
}

func (this *threePhaseInternal) DoCommit(transactionID string) (ok bool) {
	this.transactionslock.Lock()
	defer this.transactionslock.Unlock()

	item, found := this.transactions[transactionID]

	if !found {
		log.Printf("Commit: the transaction with the ID %s couldn't be found\n", transactionID)
		return false
	}

	if item.status != PhasePrepared {
		log.Printf("Commit: transaction %s isn't in the prepared phase, its phase is %d\n", transactionID, item.status)
		return false
	}

	// commit the data
	this.db.Commit([]byte(transactionID))
	item.status = PhaseCommitted

	go this.autoCleanup(transactionID)
	return true
}

// autoCleanup automatically removes a transaction after a given amount of time so the map doesn't grow too large
func (this *threePhaseInternal) autoCleanup(transactionID string) {
	time.Sleep(PhaseTimeout * 100)
	this.transactionslock.Lock()
	defer this.transactionslock.Unlock()

	delete(this.transactions, transactionID)

	log.Printf("AutoCleanup: Transaction %s was deleted\n", transactionID)
}

func (this *threePhaseInternal) PreCommit(transactionID string) (ok bool) {
	this.transactionslock.Lock()
	defer this.transactionslock.Unlock()

	item, found := this.transactions[transactionID]

	if !found {
		log.Printf("item not fund for precommit %s\n", transactionID)
		return false
	}

	if item.status != PhaseUncertain {
		log.Printf("item in wrong status for precommit %s, got %d\n", transactionID, item.status)
		return false
	}

	item.status = PhasePrepared
	this.transactions[transactionID] = item

	// auto-commit after a certain amount of time
	go this.autoCommit(transactionID)

	return true
}

func (this *threePhaseInternal) autoCommit(transactionID string) bool {
	time.Sleep(PhaseTimeout * 2)
	status, _ := this.getTransactionStatus(transactionID)
	if status != PhasePrepared {
		return false
	}

	log.Printf("AutoCommit: Transaction %s didn't complete yet, recovering.\n", transactionID)

	this.DoCommit(transactionID)
	return true
}

func (this *threePhaseInternal) getTransactionStatus(transactionID string) (phase Phase, found bool) {
	this.transactionslock.RLock()
	defer this.transactionslock.RUnlock()

	item, found := this.transactions[transactionID]

	if !found {
		return PhaseUncertain, false
	}

	return item.status, true
}

func (this *threePhaseInternal) CheckCommit(transactionID string) (didcommit bool) {
	status, _ := this.getTransactionStatus(transactionID)
	return status == PhaseCommitted || status == PhasePrepared
}

func (this *threePhaseInternal) getPeers(transactionID string) []string {
	this.transactionslock.RLock()
	defer this.transactionslock.RUnlock()
	item, found := this.transactions[transactionID]

	if !found {
		return nil
	}

	return item.Peers

}

func (this *threePhaseInternal) terminationProtocol(transactionID string) {
	for {
		time.Sleep(PhaseTimeout * 2)
		status, _ := this.getTransactionStatus(transactionID)
		peers := this.getPeers(transactionID)

		if status != PhasePrepared || peers == nil {
			return
		}

		numOkay, numNotOkay, numErr := okayCheck(this.comm.CheckCommit, []byte(transactionID), peers)

		if numOkay > 0 {
			log.Printf("Termination Protocol one host was okay\n")
			this.DoCommit(transactionID)
			return
		}

		if numErr > 1 {
			log.Printf("Termination Protocol Error, > 1 host down: ok: %d !ok: %d err: %d\n", numOkay, numNotOkay, numErr)
		}

		if numNotOkay == len(peers)-1 {
			log.Printf("Termination Protocol no hosts were okay\n")
			this.Abort(transactionID)
			return
		}
	}
}
