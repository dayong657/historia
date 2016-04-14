package threephase

type CommunicationHandler interface {
	InitializeTransaction(tx []byte, destination string) (ok bool, err error)
	Abort(transactionID []byte, destination string) (ok bool, err error)
	DoCommit(transactionID []byte, destination string) (ok bool, err error)
	PreCommit(transactionID []byte, destination string) (ok bool, err error)
	CheckCommit(transactionID []byte, destination string) (didcommit bool, err error)

	ReadData(request []byte, destination string) (result []byte, err error)
}

type NodeSet interface {
	GetCreateSet() ([]string, error)
	GetReadSet() ([]string, error)
	GetUpdateSet() ([]string, error)
	GetDeleteSet() ([]string, error)
}
