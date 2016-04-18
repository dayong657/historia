package checkup

import (
	"fmt"
	"testing"
	"time"
)

func ExampleNewTCPCheckup() {
	hosts := []string{"localhost:8080"}

	// creates a new checkup that looks for
	ckup := NewTCPCheckup(hosts)

	fmt.Printf("Number of alive hosts: %d\n", len(ckup.GetAliveHosts())) // 0
	fmt.Printf("Number of dead hosts: %d\n", len(ckup.GetDeadHosts()))   // 1

	// Output:
	// Number of alive hosts: 0
	// Number of dead hosts: 1

}

func ExampleStopCheckup() {
	hosts := []string{}
	ckup := NewTCPCheckup(hosts)

	// start the server, start should have no issues
	fmt.Println(ckup.Start() == nil)

	// stop,
	fmt.Println(ckup.Stop() == nil) // true

	// we should get an error if we start again
	fmt.Println(ckup.Stop() == AlreadyStoppedError) // true

	// Output:
	// true
	// true
	// true
}

func ExampleStartCheckup() {
	hosts := []string{}
	ckup := NewTCPCheckup(hosts)

	// start the server, start should have no issues
	fmt.Println(ckup.Start() == nil)

	// try to start again, we get an error
	fmt.Println(ckup.Start() == AlreadyRunningError) // true

	// stop,
	fmt.Println(ckup.Stop() == nil) // true

	// Output:
	// true
	// true
	// true
}

func TestConstructorNoFail(t *testing.T) {
	NewTCPCheckup([]string{})
	NewUDPCheckup([]string{})
}

func TestGettersAndSetters(t *testing.T) {
	ckup := NewTCPCheckup([]string{})

	newTimeout := -1 * time.Second
	ckup.SetTimeout(newTimeout)
	if ckup.GetTimeout() != newTimeout {
		t.Error("Could not set timeout")
	}

	newPing := -2 * time.Second
	ckup.SetPingInterval(newPing)
	if ckup.GetPingInterval() != newPing {
		t.Error("Could not set new ping")
	}
}

func TestUpdateState(t *testing.T) {
	ckup := NewTCPCheckup([]string{"a"}).(*checkupInternal)

	if len(ckup.GetDeadHosts()) != 1 {
		t.Fatal("A host was initially alive when it shouldn't have been")
	}

	if len(ckup.GetAliveHosts()) != 0 {
		t.Fatal("A host was initially dead when it shouldn't have been")
	}

	// make sure we don't update if nothing changed
	ckup.updateState("a", false, false)

	if len(ckup.GetDeadHosts()) != 1 {
		t.Fatal("State change shouldn't have happened")
	}

	ckup.updateState("a", false, true)

	if len(ckup.GetDeadHosts()) != 0 {
		t.Fatal("State change didn't happen")
	}

	if len(ckup.GetAliveHosts()) != 1 {
		t.Fatal("node disappeared")
	}
}

func TestCallback(t *testing.T) {
	ckup := NewTCPCheckup([]string{"a"}).(*checkupInternal)
	c := make(chan bool)
	ckup.SetStateChangeHandler(func(host string, state bool) {
		fmt.Printf("state ended")
		if state != true || host != "a" {
			t.Fatal("State change didn't work")
		}
		close(c)
	})

	fmt.Println("set handler")

	ckup.updateState("a", false, true)

	// wait
	select {
	case <-c:
		return
	case <-time.After(1 * time.Second):
		t.Fatal("Callback never called")
	}
}
