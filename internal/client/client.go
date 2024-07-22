package client

import (
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
)

type Client struct {
	client *http.Client
}

func New() *Client {
	jar, err := cookiejar.New(nil)
	if err != nil {
		log.Fatalf("Error creating cookie jar: %v", err)
	}

	return &Client{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
			Jar: jar,
		},
	}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.client.Do(req)
}

// SetCookies устанавливает cookies для URL.
func (c *Client) SetCookies(u *url.URL, cookies []*http.Cookie) {
	c.client.Jar.SetCookies(u, cookies)
}

// Cookies возвращает cookies для URL.
func (c *Client) Cookies(u *url.URL) []*http.Cookie {
	return c.client.Jar.Cookies(u)
}
