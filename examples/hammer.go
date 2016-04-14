package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

var (
	numSeconds = flag.Int("seconds", 5, "run for x seconds (default 5)")
	threads    = flag.Int("threads", 1, "number of concurrent requests to make")
)

func main() {
	fmt.Println("Usage: hammer <flags> host:port")
	fmt.Println("Example: hammer --threads 20 localhost:8000")

	flag.Parse()

	if len(flag.Args()) != 1 {
		fmt.Println("Error, you must specify one host.")
		return
	}

	counter := make(chan int, 1000)
	failed := make(chan int, 10000)

	for i := 0; i < *threads; i++ {
		go hammer(flag.Args()[0], counter, failed)
	}

	success := 0
	failedNum := 0
	t := time.After(time.Duration(*numSeconds) * time.Second)
	for {
		select {
		case <-t:
			fmt.Printf("Finished, made %d successful %d failed requests in %d seconds using %d threads\n", success, failedNum, *numSeconds, *threads)
			return
		case <-counter:
			success++
		case <-failed:
			failedNum++
		}
	}

}

func hammer(host string, nailed chan<- int, failed chan<- int) {
	pid := strconv.Itoa(os.Getpid())
	count := 0
	for {
		count++
		resp, err := http.Get("http://" + host + "/log/hammer_" + pid + "nail_number_" + strconv.Itoa(count))
		if err == nil && resp.StatusCode == 200 {
			resp.Body.Close()

			nailed <- 1
		} else {
			failed <- 1
		}
	}
}
