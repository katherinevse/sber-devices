package testrunner

import (
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"sber-devices/internal/form"
	"strings"
	"sync"
	"time"
)

// TODO добавить горутины  тесты.
// TODO удалить зависимости с клиентом

func Runner(qtyOfThreads int, baseURL string, finalURL string, limiter <-chan time.Time) {
	wg := sync.WaitGroup{}
	wg.Add(qtyOfThreads)

	successRate := 0
	var successRateMutex sync.Mutex

	slog.Info("Test runner is working")
	for i := 0; i < qtyOfThreads; i++ {
		go func(n int) {
			defer wg.Done()

			jar, err := cookiejar.New(nil)
			if err != nil {
				log.Fatalf("Failed to create cookie jar: %v", err)
			}

			client := &http.Client{
				CheckRedirect: func(req *http.Request, via []*http.Request) error {
					return http.ErrUseLastResponse // отключить редиректы
				},
				Jar: jar,
			}

			err = RunTests(client, baseURL, finalURL, limiter)
			if err == nil {
				slog.Info(fmt.Sprintf("Process #%d: Test successfully passed", n))
				successRateMutex.Lock()
				successRate++
				successRateMutex.Unlock()
			} else {
				log.Printf("Process #%d: Test failed with error: %v", n, err)
			}
		}(i)
	}
	wg.Wait()
	log.Printf("Successfully passed %d tests of %d\n", successRate, qtyOfThreads)
}

// Runner выполняет тест переходя от вопроса к вопросу.
func RunTests(client *http.Client, baseURL string, finalURL string, limiter <-chan time.Time) error {
	currentURL, sid, err := navigateToQuestionPage(client, baseURL, limiter)
	if err != nil {
		return err
	}

	for {
		<-limiter

		resp, err := fetchPageWithSid(client, currentURL, sid, limiter)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.Request.URL.String() == finalURL {
			log.Println("All tests completed")
			break
		}

		//// TODO del me ------------------------------------------------------
		//b, err := io.ReadAll(resp.Body)
		//if err != nil {
		//	return err
		//}
		//fmt.Println(string(b))
		//// TODO del me ------------------------------------------------------

		// Обработка текущей страницы
		formData, err := form.ParseFormData(resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing question page: %v", err)
		}

		// Отправка формы и получение ответа
		resp, location, err := submitForm(client, currentURL, formData, sid, limiter)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
			log.Println("POST request successful")
		} else {
			log.Printf("POST request failed with status code: %d", resp.StatusCode)
		}

		//nextURL
		currentURL = baseURL + location
	}

	log.Println("Test successfully completed")
	return nil
}

// fetchPageWithSid выполняет GET запрос с cookie sid.
func fetchPageWithSid(client *http.Client, url, sid string, limiter <-chan time.Time) (*http.Response, error) {
	<-limiter
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %v", err)
	}
	if sid != "" {
		req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error getting question page: %v", err)
	}
	log.Printf("Received GET response: %s %s", resp.Status, resp.Request.URL)

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("expected status 200 or 302, but got %d", resp.StatusCode)
	}

	return resp, nil
}

// submitForm выполняет POST запрос с формой.
func submitForm(client *http.Client, url string, formData url.Values, sid string, limiter <-chan time.Time) (*http.Response, string, error) {
	<-limiter

	req, err := http.NewRequest("POST", url, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("error creating POST request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if sid != "" {
		req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error sending form: %v", err)
	}
	defer resp.Body.Close()

	location := resp.Header.Get("Location")
	log.Println("Added new Location:", location)

	log.Printf("Received response: %s %s", resp.Status, resp.Request.URL)

	if resp.StatusCode != http.StatusFound {
		return nil, "", fmt.Errorf("expected status 302, but got %d", resp.StatusCode)
	}

	return resp, location, nil
}

// navigateToQuestionPage navigates to the question page and extracts the sid.
func navigateToQuestionPage(client *http.Client, baseURL string, limiter <-chan time.Time) (string, string, error) {
	<-limiter
	startURL := baseURL + "/start"

	req, err := http.NewRequest("GET", startURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request to 'Start': %v", err)
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error pressing 'Start' button: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return "", "", fmt.Errorf("expected status 302, but got %d", resp.StatusCode)
	}

	if err := extractAndSaveCookies(resp, client); err != nil {
		return "", "", fmt.Errorf("error extracting cookies after 'Start': %v", err)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", "", fmt.Errorf("no Location header found during redirect")
	}

	// Extract sid from cookies
	sid := ""
	for _, cookie := range client.Jar.Cookies(resp.Request.URL) {
		if cookie.Name == "sid" {
			sid = cookie.Value
			break
		}
	}
	if sid == "" {
		return "", "", fmt.Errorf("failed to get sid from cookies")
	}

	currentURL := baseURL + location
	return currentURL, sid, nil
}

// extractAndSaveCookies сохраняет cookies из ответа.
func extractAndSaveCookies(resp *http.Response, client *http.Client) error {
	cookies := resp.Cookies()
	u, err := url.Parse(resp.Request.URL.String())
	if err != nil {
		return fmt.Errorf("error parsing URL: %v", err)
	}
	client.Jar.SetCookies(u, cookies)

	return nil
}
