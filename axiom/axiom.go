package axiom

import (
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"golang.org/x/net/html"
)

// Client is a axiom http client
type Client struct {
	Username string
	Password string
	School   string
	// authenticity_token
	// x-csrf-token
	Token string
	// http client
	client *http.Client
}

// NewClient returns a logged in Axiom user with cookies and x-csrf token.
func NewClient(username, password, school string) (*Client, error) {
	var client *Client
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client = &Client{
		Username: username,
		Password: password,
		School:   school,
		client:   &http.Client{Jar: jar},
	}
	return client, client.login()
}

// Do sends an API request and returns the API response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// set x-csrf-token header
	req.Header.Set("x-csrf-token", c.Token)
	return c.client.Do(req)
}

func (c *Client) Report(reportID int) (*http.Response, error) {
	jsonReportURL := fmt.Sprintf("https://axiom.veracross.com/%v/query/%v/result_data.json", c.School, reportID)
	req, err := http.NewRequest("POST", jsonReportURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	return c.Do(req)
}

// logs in to Axiom
func (c *Client) login() error {
	// open Axiom login page to set cookies and x-csrf-token
	loginURL := fmt.Sprintf("https://axiom.veracross.com/%v/login", c.School)
	req, err := http.NewRequest("GET", loginURL, nil)
	if err != nil {
		return err
	}
	resp, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	err = c.setToken(resp)
	if err != nil {
		return err
	}

	// Log in
	v := url.Values{}
	v.Set("authenticity_token", c.Token)
	v.Add("login_name", c.Username)
	v.Add("password", c.Password)
	v.Encode()
	// Post form
	resp, err = c.client.PostForm(loginURL, v)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	schoolURL := fmt.Sprintf("https://axiom.veracross.com/%v/", c.School)
	req, err = http.NewRequest("GET", schoolURL, nil)
	resp, err = c.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// set the logged in x-csrf-token or return error
	return c.setToken(resp)
}

// setToken parses an http.Response for x-csrf-token
func (c *Client) setToken(resp *http.Response) error {
	// parse html for x-csrf-token
	var f func(*html.Node)

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}
	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "csrf-token" {
					token := n.Attr[0].Val
					// set Token
					c.Token = token
				}
			}
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			f(c)
		}
	}

	f(doc)
	return nil
}
