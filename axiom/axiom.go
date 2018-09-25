package axiom

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
	"golang.org/x/net/html"
)

// Client is a axiom http client
type Client struct {
	Username string
	Password string
	School   string
	// authenticity_token
	// x-csrf-token
	token string
	// http client
	client *http.Client

	logger log.Logger
}

type Option func(*Client)

func WithLogger(logger log.Logger) Option {
	return func(c *Client) {
		c.logger = logger
	}
}

// NewClient returns a logged in Axiom Client
func NewClient(username, password, school string, opts ...Option) (*Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}
	client := &Client{
		Username: username,
		Password: password,
		School:   school,
		client:   &http.Client{Jar: jar},
		logger:   log.NewNopLogger(),
	}

	for _, optFn := range opts {
		optFn(client)
	}

	level.Debug(client.logger).Log("msg", "configured client", "school", school)
	return client, client.login()
}

// Report returns the json data for a report ID
func (c *Client) Report(reportID int) (*http.Response, error) {
	urlStr := fmt.Sprintf("https://axiom.veracross.com/%s/query/%d/result_data.json", c.School, reportID)
	req, err := http.NewRequest("POST", urlStr, nil)
	if err != nil {
		return nil, err
	}
	level.Debug(c.logger).Log(
		"msg", "getting report from veracross",
		"school", c.School,
		"report", reportID,
		"url", urlStr,
	)
	req.Header.Set("x-csrf-token", c.token)
	return c.client.Do(req)
}

func (c *Client) login() error {
	// visit login page and extract csrf token from html form
	urlStr := fmt.Sprintf("https://accounts.veracross.com/%s/axiom/login", c.School)
	resp, err := c.client.Get(urlStr)
	if err != nil {
		return errors.Wrapf(err, "get request for %s", urlStr)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("got %s for url %s", resp.Status, urlStr)
	}

	_, tok, err := htmlFormValues(resp)
	if err != nil {
		return errors.Wrap(err, "get token from login url")
	}
	resp.Body.Close()

	level.Debug(c.logger).Log(
		"msg", "extracted token from login page",
		"token", tok,
		"url", urlStr,
	)

	// visit authenticate page, and extract new csrf and authenticity tokens
	urlStr = fmt.Sprintf("https://accounts.veracross.com/%s/axiom/authenticate", c.School)
	v := url.Values{}
	v.Set("authenticity_token", tok)
	v.Add("username", c.Username)
	v.Add("password", c.Password)
	v.Encode()
	resp, err = c.client.PostForm(urlStr, v)
	if err != nil {
		return errors.Wrapf(err, "authenticate to %s", urlStr)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("got %s for url %s", resp.Status, urlStr)
	}
	aut, tok, err := htmlFormValues(resp)
	if err != nil {
		return err
	}
	resp.Body.Close()

	level.Debug(c.logger).Log(
		"msg", "got csrf and authenticity_token from authenticate form",
		"csrf_token", tok,
		"authenticity_token", aut,
		"urlStr", urlStr,
	)

	// submit form to session page, get back a usable token
	v = url.Values{}
	v.Set("authenticity_token", tok)
	v.Set("account", aut)
	v.Encode()
	urlStr = fmt.Sprintf("https://axiom.veracross.com/%s/session", c.School)
	resp, err = c.client.PostForm(urlStr, v)
	if err != nil {
		return errors.Wrapf(err, "post form to %s", urlStr)
	}
	if resp.StatusCode != http.StatusOK {
		return errors.Errorf("got %s for url %s", resp.Status, urlStr)
	}

	_, tok, err = htmlFormValues(resp)
	if err != nil {
		return err
	}
	resp.Body.Close()

	level.Debug(c.logger).Log(
		"msg", "completed login",
		"token", tok,
	)

	// set csrf token
	c.token = tok
	return nil
}

func htmlFormValues(resp *http.Response) (aut, csrf string, err error) {
	var f func(*html.Node)

	doc, err := html.Parse(resp.Body)
	if err != nil {
		return "", "", errors.Wrap(err, "parsing html body to extract token")
	}

	f = func(n *html.Node) {
		if n.Type == html.ElementNode && n.Data == "meta" {
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "csrf-token" {
					csrf = n.Attr[1].Val // used to be 0
				}
			}
		}
		if n.Type == html.ElementNode && n.Data == "input" {
			for _, a := range n.Attr {
				if a.Key == "name" && a.Val == "account" {
					aut = n.Attr[3].Val
				}
			}
		}
		for cd := n.FirstChild; cd != nil; cd = cd.NextSibling {
			f(cd)
		}
	}

	f(doc)
	// Check that the token was set
	if csrf == "" {
		return "", "", errors.New("csrf token not set")
	}
	return aut, csrf, nil
}
