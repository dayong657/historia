package main

//
// import (
// 	"encoding/json"
// 	"fmt"
// 	"io"
// 	"io/ioutil"
// 	"log"
// 	"net/http"
// 	"os"
// 	"strconv"
// 	"sync"
// 	"time"
//
// 	"github.com/gorilla/mux"
// 	"github.com/josephlewis42/historia/checkup"
// 	"github.com/josephlewis42/historia/cohort"
// 	"github.com/josephlewis42/historia/storage"
// 	"github.com/josephlewis42/historia/threephase"
// )
//
// var (
// 	ckup   checkup.Checkup
// 	rwmode cohort.RWMode
//
// 	database         = storage.NewInMemoryStorage()
// 	transactions     = make(map[string]threephase.ThreePhaseTransaction)
// 	transactionslock sync.RWMutex
// )
//
// /**
// func MyHandler(w http.ResponseWriter, r *http.Request) {
// 	err := r.ParseForm()
//
// 	if err != nil {
// 		// Handle error
// 	}
//
// 	decoder := schema.NewDecoder()
// 	// r.PostForm is a map of our POST form values
// 	err := decoder.Decode(person, r.PostForm)
//
// 	if err != nil {
// 		// Handle error
// 	}
//
// 	// Do something with person.Name or person.Phone
// }
// **/
//
// func InsertDataHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	key := vars["key"]
//
// 	log.Println("Got posted data for key: " + key)
// 	body, _ := ioutil.ReadAll(io.LimitReader(r.Body, 4096))
// 	log.Println("data: " + string(body))
//
// 	// TODO do the master stuff
// }
//
// func ReadDataHandler(w http.ResponseWriter, r *http.Request) {
// 	vars := mux.Vars(r)
// 	key := vars["key"]
// 	from := vars["from"]
// 	to := vars["to"]
//
// 	log.Printf("Got data request for key: %s between: %s and %s\n", key, from, to)
//
// 	// TODO do the read and merge
// }
//
// func ThreePhaseCall(wrapped func(tx threephase.ThreePhaseTransaction, w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
//
// 		id, ok := vars["key"]
//
// 		if !ok {
// 			log.Printf("Error, key (%s) to %s was not an integer\n", vars["key"], r.URL)
// 			w.WriteHeader(http.StatusBadRequest)
// 			return
// 		}
//
// 		transactionslock.RLock()
//
// 		wrapped([]byte(id), w, r)
// 	}
// }
//
// func CanCommit(w http.ResponseWriter, r *http.Request) {
// 	// unmarshall transaction
// 	body := r.Body
// 	decoder := json.NewDecoder(body)
// 	var tx threephase.ThreePhaseTransaction
// 	err := decoder.Decode(&tx)
// 	if err != nil {
// 		log.Printf("ERROR, bad JSON for threephase transaction: %s\n", body)
//
// 		// send abort
// 		w.WriteHeader(http.StatusConflict)
// 	}
//
// 	transactionslock.Lock()
// 	defer transactionslock.Unlock()
//
// 	// make sure it isn't in map
// 	_, found := transactions[tx.Key]
// 	if found {
// 		log.Printf("ERROR, already have entry for transaction: %d\n", tx.Timestamp)
// 		w.WriteHeader(http.StatusConflict)
// 	}
// 	transactions[tx.Key] = tx
//
// 	// grab space in precommit line
// 	ok := database.Precommit([]byte(tx.Key), tx.Data)
// 	if ok {
// 		// yes
// 		w.WriteHeader(http.StatusOK)
// 	} else {
// 		// no
// 		w.WriteHeader(http.StatusConflict)
// 	}
//
// }
//
// func PreCommit(tx threephase.ThreePhaseTransaction, w http.ResponseWriter, r *http.Request) {
//
// 	// yes
// 	w.WriteHeader(http.StatusOK)
//
// 	// yes
// 	w.WriteHeader(http.StatusConflict)
//
// }
//
// func DoCommit(tx threephase.ThreePhaseTransaction, w http.ResponseWriter, r *http.Request) {
//
// 	// ack
// 	w.WriteHeader(http.StatusOK)
// }
//
// func DoAbort(tx threephase.ThreePhaseTransaction, w http.ResponseWriter, r *http.Request) {
//
// }
//
// func WasCommit(tx threephase.ThreePhaseTransaction, w http.ResponseWriter, r *http.Request) {
//
// }
//
// func main() {
// 	addresses := []string{}
// 	if len(os.Args) < 3 {
// 		fmt.Printf("Usage: %s <nodenum> <host:port> [<host:port>]+\n", os.Args[0])
// 		return
// 	}
//
// 	itemnum, err := strconv.Atoi(os.Args[1])
// 	if err != nil || itemnum < 0 || itemnum > len(os.Args)-2 {
// 		fmt.Printf("Illegal node number, it must be in the range 1-num of nodes\n")
// 		return
// 	}
// 	itemnum -= 1
//
// 	log.Printf("We are node %d of %d\n", itemnum+1, len(os.Args)-2)
// 	for i := 2; i < len(os.Args); i++ {
// 		log.Printf("Node %d is at http://%s\n", i-1, os.Args[i])
// 		addresses = append(addresses, os.Args[i])
// 	}
//
// 	log.Println("START CHECKUP SEQUENCE")
// 	// start up the checkup
// 	ckup = checkup.NewTCPCheckup(addresses)
// 	ckup.SetTimeout(1 * time.Second)
// 	ckup.SetPingInterval(2 * time.Second)
// 	ckup.SetStateChangeHandler(func(host string, state bool) {
// 		if state == true {
// 			log.Printf("Recovery detected at: %s\n", host)
// 		} else {
// 			log.Printf("Crash detected at: %s\n", host)
// 		}
// 	})
// 	ckup.Start()
// 	defer ckup.Stop()
// 	log.Println("END CHECKUP SEQUENCE")
//
// 	rwmode = cohort.NewReadMajorityWriteMajority(len(addresses))
//
// 	log.Println("STARTING SERVER")
//
// 	r := mux.NewRouter()
//
// 	r.HandleFunc("/data/{key}", InsertDataHandler).Methods("POST")
// 	r.HandleFunc("/data/{key}/{from:[0-9]+}/{to:[0-9]+}", ReadDataHandler).Methods("GET")
//
// 	r.HandleFunc("/v1/3pc/init", CanCommit).Methods("POST").Name("v1can")
// 	r.HandleFunc("/v1/3pc/commit/{id}", ThreePhaseCall(PreCommit)).Methods("GET").Name("v1pre")
// 	r.HandleFunc("/v1/3pc/precommit/{id}", ThreePhaseCall(DoCommit)).Methods("GET").Name("v1commit")
// 	r.HandleFunc("/v1/3pc/abort/{id}", ThreePhaseCall(DoAbort)).Methods("GET").Name("v1abort")
// 	r.HandleFunc("/v1/3pc/check/{id}", ThreePhaseCall(WasCommit)).Methods("GET").Name("v1check")
//
// 	http.Handle("/", r)
// 	http.ListenAndServe(addresses[itemnum], nil)
//
// }
//
// /**
// 	Abort(transactionID []byte, destination string) (ok bool, err error)
// 	DoCommit(transactionID []byte, destination string) (ok bool, err error)
// 	PreCommit(transactionID []byte, destination string) (ok bool, err error)
// 	CheckCommit(transactionID []byte, destination string) (didcommit bool, err error)
//
// 	ReadData(request []byte, destination string) (result []byte, err error)
// **/
