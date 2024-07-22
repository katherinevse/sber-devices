package test_runner

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
	SetCookies(u *url.URL, cookies []*http.Cookie)
	Cookies(u *url.URL) []*http.Cookie
}

type TestRunner struct {
	client   HTTPClient
	baseURL  string
	finalURL string
	limiter  <-chan time.Time
}

// NewTestRunner создает новый экземпляр testRunner.
func NewTestRunner(client HTTPClient, baseURL, finalURL string, limiter <-chan time.Time) *TestRunner {
	return &TestRunner{
		client:   client,
		baseURL:  baseURL,
		finalURL: finalURL,
		limiter:  limiter,
	}
}

// RunTests выполняет тесты, переходя от вопроса к вопросу.
func (tr *TestRunner) RunTests() error {
	currentURL, sid, err := tr.navigateToQuestionPage()
	if err != nil {
		return err
	}

	for {
		<-tr.limiter

		resp, err := tr.fetchPageWithSid(currentURL, sid)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.Request.URL.String() == tr.finalURL {
			log.Println("All tests completed")
			break
		}

		formData, err := ParseFormData(resp.Body)
		if err != nil {
			return fmt.Errorf("error parsing question page: %v", err)
		}

		resp, location, err := tr.submitForm(currentURL, formData, sid)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusFound {
			log.Println("POST request successful")
		} else {
			log.Printf("POST request failed with status code: %d", resp.StatusCode)
		}

		currentURL = tr.baseURL + location
	}

	log.Println("Test successfully completed")
	return nil
}

// fetchPageWithSid выполняет GET запрос с cookie sid.
func (tr *TestRunner) fetchPageWithSid(url, sid string) (*http.Response, error) {
	<-tr.limiter
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("error creating GET request: %v", err)
	}
	if sid != "" {
		req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	}

	resp, err := tr.client.Do(req)
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
func (tr *TestRunner) submitForm(url string, formData url.Values, sid string) (*http.Response, string, error) {
	<-tr.limiter

	req, err := http.NewRequest("POST", url, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, "", fmt.Errorf("error creating POST request: %v", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	if sid != "" {
		req.AddCookie(&http.Cookie{Name: "sid", Value: sid})
	}

	resp, err := tr.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("error sending test_runner: %v", err)
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
func (tr *TestRunner) navigateToQuestionPage() (string, string, error) {
	<-tr.limiter
	startURL := tr.baseURL + "/start"

	req, err := http.NewRequest("GET", startURL, nil)
	if err != nil {
		return "", "", fmt.Errorf("error creating request to 'Start': %v", err)
	}

	resp, err := tr.client.Do(req)
	if err != nil {
		return "", "", fmt.Errorf("error pressing 'Start' button: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusFound {
		return "", "", fmt.Errorf("expected status 302, but got %d", resp.StatusCode)
	}

	if err := tr.extractAndSaveCookies(resp); err != nil {
		return "", "", fmt.Errorf("error extracting cookies after 'Start': %v", err)
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return "", "", fmt.Errorf("no Location header found during redirect")
	}

	// Extract sid from cookies
	sid := ""
	for _, cookie := range tr.client.Cookies(resp.Request.URL) {
		if cookie.Name == "sid" {
			sid = cookie.Value
			break
		}
	}
	if sid == "" {
		return "", "", fmt.Errorf("failed to get sid from cookies")
	}

	currentURL := tr.baseURL + location
	return currentURL, sid, nil
}

// extractAndSaveCookies сохраняет cookies из ответа.
func (tr *TestRunner) extractAndSaveCookies(resp *http.Response) error {
	cookies := resp.Cookies()
	u, err := url.Parse(resp.Request.URL.String())
	if err != nil {
		return fmt.Errorf("error parsing URL: %v", err)
	}
	tr.client.SetCookies(u, cookies)

	return nil
}
