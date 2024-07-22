package main

import (
	"log"
	"log/slog"
	"os"
	"sber-devices/internal/testrunner"
	"strconv"
	"time"
)

//TODO архитектуру проекта сделать
// TODO добавить горутины.
// TODO удалить зависимости с клиентом

func main() {
	loggerConfig()

	firstURL := "http://193.168.227.93"
	finalURL := "http://193.168.227.93/passed"

	//qtyOfThreads, err := strconv.Atoi("5")
	//if err != nil {
	//	log.Fatalf("Failed to parse quantity of threads: %s\n", "5")
	//}

	maxRPS, err := strconv.Atoi("3")
	if err != nil {
		log.Fatalf("Failed to parse RPS parameter: %s\n", "3")
	}
	limiter := time.Tick(getTimeLimit(maxRPS))

	testrunner.RunTests(firstURL, finalURL, limiter)
}

func getTimeLimit(rps int) time.Duration {
	if rps <= 0 {
		return time.Second
	}
	return time.Duration(1000/rps) * time.Millisecond
}

func loggerConfig() {
	level := slog.LevelInfo
	err := level.UnmarshalText([]byte(os.Getenv("LOG_LEVEL")))
	if err != nil {
		slog.Info("Undefined log level")
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	handler := slog.NewJSONHandler(os.Stdout, opts)
	logger := slog.New(handler)
	slog.SetDefault(logger)
}
