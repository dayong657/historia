package checkup

import "fmt"

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
