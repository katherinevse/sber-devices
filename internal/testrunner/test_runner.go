package testrunner

import (
	"net/http"
	"time"
)

// Runner будет выполнять тест, проходя по всем вопросам и заполняя формы
func Runner(client http.Client, startURL string, finalURL string, limiter <-chan time.Time) error {
	nextURL := startURL
}
