package threephase

import "github.com/josephlewis42/historia/storage"

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

func NewThreePhaseCommit(comm CommunicationHandler, db storage.Storage, ch NodeSet) ThreePhaseCommit {
	return newThreePhaseInternal(comm, db, ch)
}

func newThreePhaseInternal(comm CommunicationHandler, db storage.Storage, ch NodeSet) *threePhaseInternal {
	return &threePhaseInternal{
		comm:         comm,
		db:           db,
		ch:           ch,
		transactions: make(map[string]*ThreePhaseTransaction),
	}
}
