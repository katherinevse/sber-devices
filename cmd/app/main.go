package main

import (
	"fmt"
	"log"
	"log/slog"
	"sber-devices/internal/client"
	"sber-devices/internal/logger"
	"sber-devices/internal/test_runner"
	"strconv"
	"sync"
	"time"
)

func main() {
	logger.Configure()

	baseURL := "http://193.168.227.93"
	finalURL := "http://193.168.227.93/passed"

	qtyOfThreads, err := strconv.Atoi("5")
	if err != nil {
		log.Fatalf("Failed to parse quantity of threads: %s\n", "5")
	}

	maxRPS, err := strconv.Atoi("3")
	if err != nil {
		log.Fatalf("Failed to parse RPS parameter: %s\n", "3")
	}
	limiter := time.Tick(getTimeLimit(maxRPS))
	runner(qtyOfThreads, baseURL, finalURL, limiter)
}

func getTimeLimit(rps int) time.Duration {
	if rps <= 0 {
		return time.Second
	}

	return time.Duration(1000/rps) * time.Millisecond
}

func runner(qtyOfThreads int, baseURL string, finalURL string, limiter <-chan time.Time) {
	wg := sync.WaitGroup{}
	wg.Add(qtyOfThreads)

	successRate := 0
	var successRateMutex sync.Mutex

	slog.Info("Test runner is working")
	for i := 0; i < qtyOfThreads; i++ {
		go func(n int) {
			defer wg.Done()
			testRunner := test_runner.NewTestRunner(client.New(), baseURL, finalURL, limiter)
			err := testRunner.RunTests()
			if err != nil {
				log.Printf("Process #%d: Test failed with error: %v", n, err)
				return
			}

			slog.Info(fmt.Sprintf("Process #%d: Test successfully passed", n))

			successRateMutex.Lock()
			defer successRateMutex.Unlock()
			successRate++
		}(i)
	}

	wg.Wait()

	log.Printf("Successfully passed %d tests of %d\n", successRate, qtyOfThreads)
}
