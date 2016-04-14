package storage

type Storage interface {
	Read(request []byte) (value []byte, ok bool)
	Commit(transactionID []byte) error
	Prepare(transactionID, value []byte) bool
	Abort(transactionID []byte) bool
	Merge(request []byte, response [][]byte) (result []byte, ok bool)
	Stats() string
}
