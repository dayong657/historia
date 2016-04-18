package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/josephlewis42/historia/cohort"
	"github.com/josephlewis42/historia/storage"
	"github.com/josephlewis42/historia/threephase"
)

func NewThreePhaseHTTP(thishost int, hosts []string, db storage.Storage) {
	cohort := cohort.NewDefaultCohort(thishost, hosts)

	var tpi threePhaseHTTPImplementation
	tpi.db = db
	tpi.hosts = hosts
	tpi.myhost = hosts[thishost]
	tpi.chrt = cohort

	tpi.tpc = threephase.NewThreePhaseCommit(tpi, db, &cohort)

	r := mux.NewRouter()

	r.HandleFunc("/3pc/init/{id}", threePhaseInit("init", tpi.tpc.InitializeTransaction)).Methods("GET")
	r.HandleFunc("/3pc/abort/{id}", threePhaseCall("abort", tpi.tpc.Abort)).Methods("GET")
	r.HandleFunc("/3pc/commit/{id}", threePhaseCall("pre", tpi.tpc.DoCommit)).Methods("GET")
	r.HandleFunc("/3pc/precommit/{id}", threePhaseCall("commit", tpi.tpc.PreCommit)).Methods("GET")
	r.HandleFunc("/3pc/check/{id}", threePhaseCall("check", tpi.tpc.CheckCommit)).Methods("GET")

	r.HandleFunc("/log/{value}", tpi.clientCreate).Methods("GET")
	r.HandleFunc("/stats", tpi.statistics).Methods("GET")
	r.HandleFunc("/", tpi.root)

	log.Printf("Starting on %s\n", tpi.myhost)
	http.Handle("/", r)
	http.ListenAndServe(tpi.myhost, nil)
}

/**
	InitializeTransaction(transaction []byte) (ok bool)
	Abort(transactionID string) (ok bool)
	DoCommit(transactionID string) (ok bool)
	PreCommit(transactionID string) (ok bool)
	CheckCommit(transactionID string) (didcommit bool)
**/

func threePhaseInit(name string, wrapped func(tx []byte) bool) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handling %s\n", r.URL)
		vars := mux.Vars(r)
		id, ok := vars["id"]

		if !ok {
			log.Printf("%s Error, id wasn't found.\n", r.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		decoded, err := base64.StdEncoding.DecodeString(id)
		if err != nil {
			log.Printf("Error decoding Base64: %s\n", err)
		}
		result := wrapped(decoded)

		if result == true {
			w.WriteHeader(200)
			w.Write([]byte("Success"))
		} else {
			w.WriteHeader(400)
			w.Write([]byte("Failure"))
		}
	}

}

func threePhaseCall(name string, wrapped func(tx string) bool) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		log.Printf("handling %s\n", r.URL)

		vars := mux.Vars(r)
		id, ok := vars["id"]

		if !ok {
			log.Printf("%s Error, id wasn't found.", r.URL)
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		result := wrapped(id)

		if result == true {
			w.WriteHeader(200)
			w.Write([]byte("Success"))
		} else {
			w.WriteHeader(400)
			w.Write([]byte("Failure"))
		}
	}
}

type threePhaseHTTPImplementation struct {
	db     storage.Storage
	hosts  []string
	myhost string
	tpc    threephase.ThreePhaseCommit
	chrt   cohort.Cohort
}

// Satisfies the callback interface for 3PC
func (this threePhaseHTTPImplementation) Create(tx []byte, destination string) (ok bool, err error) {
	encoded := base64.StdEncoding.EncodeToString(tx)

	resp, err := http.Get("http://" + destination + "/3pc/init/" + url.QueryEscape(encoded))
	resp.Body.Close()
	return resp.StatusCode == 200, err
}

// Satisfies the callback interface for 3PC
func (this threePhaseHTTPImplementation) InitializeTransaction(tx []byte, destination string) (ok bool, err error) {
	encoded := base64.StdEncoding.EncodeToString(tx)

	resp, err := http.Get("http://" + destination + "/3pc/init/" + url.QueryEscape(encoded))
	if resp != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		return resp.StatusCode == 200, err
	}
	return false, err
}

// Satisfies the callback interface for 3PC
func (this threePhaseHTTPImplementation) Abort(tx []byte, destination string) (ok bool, err error) {
	resp, err := http.Get("http://" + destination + "/3pc/abort/" + string(tx))
	if resp != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		return resp.StatusCode == 200, err
	}
	return false, err
}

// Satisfies the callback interface for 3PC
func (this threePhaseHTTPImplementation) DoCommit(tx []byte, destination string) (ok bool, err error) {
	resp, err := http.Get("http://" + destination + "/3pc/commit/" + string(tx))
	if resp != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		return resp.StatusCode == 200, err
	}
	return false, err
}

// Satisfies the callback interface for 3PC
func (this threePhaseHTTPImplementation) PreCommit(tx []byte, destination string) (ok bool, err error) {
	resp, err := http.Get("http://" + destination + "/3pc/precommit/" + string(tx))
	if resp != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		return resp.StatusCode == 200, err
	}
	return false, err
}

// Satisfies the callback interface for 3PC
func (this threePhaseHTTPImplementation) CheckCommit(tx []byte, destination string) (didcommit bool, err error) {
	resp, err := http.Get("http://" + destination + "/3pc/check/" + string(tx))
	if resp != nil {
		if resp.Body != nil {
			resp.Body.Close()
		}
		return resp.StatusCode == 200, err
	}
	return false, err
}

// Satisfies the callback interface for 3PC
func (this threePhaseHTTPImplementation) ReadData(tx []byte, destination string) (result []byte, err error) {
	// TODO allow queries in the future
	return []byte{}, errors.New("Reading data is not implemented yet")

	//resp, err := http.Get("http://" + destination + "/3pc/read/" + url.QueryEscape(string(tx)))
	//return resp.StatusCode == 200, err
}

func (this threePhaseHTTPImplementation) clientCreate(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	value, _ := vars["value"]
	log.Printf("inserting %s\n", value)

	result := this.tpc.Create([]byte(value))

	if result == true {
		w.WriteHeader(200)
		w.Write([]byte("Success"))
	} else {
		w.WriteHeader(400)
		w.Write([]byte("Failure"))
	}

}

func (this threePhaseHTTPImplementation) statistics(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	w.Write([]byte(this.db.Stats()))
}

func (this threePhaseHTTPImplementation) root(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)

	w.Write([]byte("====== Connections ======\n\n"))

	for _, host := range this.chrt.GetAliveSet() {
		w.Write([]byte("* " + host + "\n"))
	}

	w.Write([]byte("====== Database ======\n\n"))
	w.Write([]byte(this.db.Stats()))

}

func main() {
	addresses := []string{}
	if len(os.Args) < 3 {
		fmt.Printf("Usage: %s <nodenum> <host:port> [<host:port>]+\n", os.Args[0])
		return
	}

	itemnum, err := strconv.Atoi(os.Args[1])
	if err != nil || itemnum < 0 || itemnum > len(os.Args)-2 {
		fmt.Printf("Illegal node number, it must be in the range 1-num of nodes\n")
		return
	}
	itemnum -= 1

	log.Printf("We are node %d of %d\n", itemnum+1, len(os.Args)-2)
	for i := 2; i < len(os.Args); i++ {
		log.Printf("Node %d is at http://%s\n", i-1, os.Args[i])
		addresses = append(addresses, os.Args[i])
	}

	NewThreePhaseHTTP(itemnum, addresses, storage.NewInMemoryStorage())

}
