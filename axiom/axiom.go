package axiom

import (
	"errors"
	"fmt"
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

	// school URL
	URL *url.URL
}

// NewClient returns a logged in Axiom Client
func NewClient(username, password, school string) (*Client, error) {
	client, err := newClient(username, password, school)
	if err != nil {
		return nil, err
	}
	return client, client.login()
}

// submit sends the login form.
func (c *Client) submit() error {
	urlStr := fmt.Sprintf("%s/login", c.URL)
	v := url.Values{}
	v.Set("authenticity_token", c.Token)
	v.Add("login_name", c.Username)
	v.Add("password", c.Password)
	v.Encode()
	// Post form
	_, err := c.client.PostForm(urlStr, v)
	if err != nil {
		return err
	}
	return nil
}

// Do sends an API request and returns the API response.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	// set x-csrf-token header
	req.Header.Set("x-csrf-token", c.Token)
	return c.client.Do(req)
}

// Report returns the json data for a report ID
func (c *Client) Report(reportID int) (*http.Response, error) {
	jsonReportURL := fmt.Sprintf("%s/query/%v/result_data.json", c.URL, reportID)
	req, err := http.NewRequest("POST", jsonReportURL, nil)
	if err != nil {
		return nil, err
	}
	return c.Do(req)
}

func (c *Client) login() error {
	urlStr := fmt.Sprintf("%s/login", c.URL)
	err := c.loginToken(urlStr)
	if err != nil {
		return err
	}
	err = c.submit()
	if err != nil {
		return err
	}
	urlStr = fmt.Sprintf("%s/#/homepage/main", c.URL)
	return c.loginToken(urlStr)
}

// newClient returns a client
func newClient(username, password, school string) (*Client, error) {
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
	urlStr := fmt.Sprintf("https://axiom.veracross.com/%s/", client.School)
	schoolURL, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	client.URL = schoolURL
	return client, nil
}

// set x-csrf token for url
func (c *Client) loginToken(urlStr string) error {
	resp, err := c.client.Get(urlStr)
	if err != nil {
		return err
	}
	return c.setToken(resp)
}

// setToken parses an http.Response for x-csrf-token
func (c *Client) setToken(resp *http.Response) error {
	// parse html for x-csrf-token
	var f func(*html.Node)
	var token string

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return err
	}

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "csrf-token" {
					token = n.Attr[1].Val // used to be 0
					// set Token
					c.Token = token
				}
			}
		}
		for cd := n.FirstChild; cd != nil; cd = cd.NextSibling {
			f(cd)
		}
	}

	f(doc)
	// Check that the token was set
	if token == "" {
		return errors.New("csrf-token not set")
	}
	return nil
}
