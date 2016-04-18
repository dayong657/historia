package threephase

import "errors"

type HandlerCallback func(tx []byte, dest string) (ok bool, err error)

func newHandlerCallback(notOkDests []string, timeoutDests []string, ret chan<- string) HandlerCallback {
	if notOkDests == nil {
		notOkDests = []string{}
	}

	if timeoutDests == nil {
		timeoutDests = []string{}
	}

	return func(tx []byte, dest string) (ok bool, err error) {
		//log.Printf("Processing tx: %s, dest: %s, ret is null? %t\n", string(tx), dest, ret == nil)
		if ret != nil {
			ret <- dest
		}

		for _, errDest := range notOkDests {
			if errDest == dest {
				return false, nil
			}
		}

		for _, timeoutDest := range timeoutDests {
			if timeoutDest == dest {
				// 	time.Sleep(time.Second * 3)
				return false, errors.New(dest + " fake timed out")
			}
		}

		return true, nil
	}
}

type HostGetter func() ([]string, error)

func newHostGetter(hosts []string, err error, ret chan bool) HostGetter {
	return func() ([]string, error) {
		if ret != nil {
			ret <- true
		}

		if err != nil {
			return nil, err
		}

		return hosts, nil
	}
}

type fakeCommunicationHandler struct {
	InitializeTransactionI HandlerCallback
	AbortI                 HandlerCallback
	DoCommitI              HandlerCallback
	PreCommitI             HandlerCallback
	CheckCommitI           HandlerCallback

	GetCreateSetI HostGetter
	GetReadSetI   HostGetter
	GetUpdateSetI HostGetter
	GetDeleteSetI HostGetter
}

func (f *fakeCommunicationHandler) InitializeTransaction(tx []byte, dest string) (ok bool, err error) {
	return f.InitializeTransactionI(tx, dest)
}

func (f *fakeCommunicationHandler) Abort(tx []byte, dest string) (ok bool, err error) {
	return f.AbortI(tx, dest)
}

func (f *fakeCommunicationHandler) CheckCommit(tx []byte, dest string) (ok bool, err error) {
	return f.CheckCommitI(tx, dest)
}

func (f *fakeCommunicationHandler) PreCommit(tx []byte, dest string) (ok bool, err error) {
	return f.PreCommitI(tx, dest)
}

func (f *fakeCommunicationHandler) DoCommit(tx []byte, dest string) (ok bool, err error) {
	return f.DoCommitI(tx, dest)
}

func (f *fakeCommunicationHandler) ReadData(tx []byte, dest string) (ok []byte, err error) {
	return []byte{}, nil
}

func (f *fakeCommunicationHandler) GetCreateSet() ([]string, error) {
	return f.GetCreateSetI()
}

func (f *fakeCommunicationHandler) GetReadSet() ([]string, error) {
	return f.GetReadSetI()
}

func (f *fakeCommunicationHandler) GetUpdateSet() ([]string, error) {
	return f.GetUpdateSetI()
}

func (f *fakeCommunicationHandler) GetDeleteSet() ([]string, error) {
	return f.GetDeleteSetI()
}

/**

	GetCreateSetI HostGetter
	GetReadSetI   HostGetter
	GetUpdateSetI HostGetter
	GetDeleteSetI HostGetter
**/

func newFakeComm(hosts []string) fakeCommunicationHandler {
	var this fakeCommunicationHandler
	this.InitializeTransactionI = newHandlerCallback([]string{}, []string{}, nil)
	this.AbortI = newHandlerCallback([]string{}, []string{}, nil)
	this.DoCommitI = newHandlerCallback([]string{}, []string{}, nil)
	this.PreCommitI = newHandlerCallback([]string{}, []string{}, nil)
	this.CheckCommitI = newHandlerCallback([]string{}, []string{}, nil)

	this.GetCreateSetI = newHostGetter(hosts, nil, nil)
	this.GetReadSetI = newHostGetter(hosts, nil, nil)
	this.GetUpdateSetI = newHostGetter(hosts, nil, nil)
	this.GetDeleteSetI = newHostGetter(hosts, nil, nil)

	return this
}
