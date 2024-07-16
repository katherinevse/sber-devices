package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

func main() {

	startURL := os.Getenv("START_PAGE")
	finalURL := os.Getenv("FINAL_PAGE")

	//горутины
	qtyOfThreads, err := strconv.Atoi(os.Getenv("СOUNT_THREADS"))
	if err != nil {
		log.Fatalf("Failed to parse quantity of threads: %v\n", err)
	}

	//RPS
	maxRPS, err := strconv.Atoi(os.Getenv("MAX_RPS"))
	if err != nil {
		log.Fatalf("Failed to parse RPS parameter: %v\n", err)
	}
	limiter := time.Tick(getTimeLimit(maxRPS))

	var wg sync.WaitGroup
	wg.Add(qtyOfThreads)

	for i := 0; i < qtyOfThreads; i++ {
		go func(clientID int) {
			defer wg.Done()

			client := &http.Client{}
			err := Runner(client, startURL, finalURL, limiter)
			if err != nil {
				log.Printf("Client %d failed: %v", clientID, err)
			} else {
				log.Printf("Client %d successfully completed the testrunner", clientID)
			}
		}(i)
	}

	wg.Wait()
	fmt.Println("All tests completed")

}

// getTimeLimit возвращает интервал времени между запросами для ограничения RPS
func getTimeLimit(rps int) time.Duration {
	if rps <= 0 {
		return time.Second
	}
	return time.Duration(1000/rps) * time.Millisecond
}
