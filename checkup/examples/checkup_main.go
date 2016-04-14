package main

import "fmt"
import "github.com/josephlewis42/historia/checkup"

func main() {
	c := checkup.NewTCPCheckup([]string{"localhost:8000", "localhost:4001", "localhost:4002"})

	c.SetStateChangeHandler(func(host string, state bool) {
		fmt.Printf("State change alert host: %s is up? %t\n", host, state)
	})
	c.Start()

	select {}
}
