package axiom

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

var head = `
<!DOCTYPE html>
<html>
<head>
    <meta charset="utf-8" />
    <meta http-equiv="X-UA-Compatible" content="IE=edge">
    <meta name="ROBOTS" content="NOINDEX, NOFOLLOW, NOARCHIVE">


    <title>Axiom</title>
        <script src="//js.honeybadger.io/v0.2/honeybadger.min.js" type="text/javascript"></script>
        <script type="text/javascript">
          Honeybadger.configure({
            api_key: '52caa67695779873c7b23f37fc030600',
            environment: 'production',
          });
        </script>

    <link href="/assets/vendor-d27a97e18f050015cf59626af5448073.css" media="all" rel="stylesheet" />
    <link href="/assets/session-c3fce1a5c5162587ef068a1fa32100e7.css" media="all" rel="stylesheet" />

    <script src="/assets/vendor-eda8ad85c03e1c02a33dad00591a3fac.js"></script>
    <script src="/assets/session-79723aa99a35430e45d3943b6cb5b5e0.js"></script>

    <meta content="authenticity_token" name="csrf-param" />
<meta name="csrf-token" content="Pl0opv2ylPnZWrWtSon9siyOL5hjhpZWPFm5nsQtikY=" />

    <script type="text/javascript">
        $.ajaxSetup({
            headers: {
                'X-CSRF-Token': $('meta[name="csrf-token"]').attr('content'),
            },
        });
    </script>
</head>
`

// returns an HTML document with sample axiom login html
var loginHandler = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, head)
}

// returns hello world
var helloHandler = func(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "Hello World")
}

// mock http server
func mockServer(f http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(f))
}

// test that NewClient returns http.Client with a cookiejar
func TestNewClientHTTP(t *testing.T) {
	username := "foo"
	password := "bar"
	school := "baz"
	client, err := NewClient(username, password, school)
	if err != nil {
		t.Fatal("Expected to get a new client", "got", err)
	}

	if client.client == nil {
		t.Fatal("nil http client returned")
	}
	// TODO: Test cookiejar
}

type recordingTransport struct {
	req *http.Request
}

func (t *recordingTransport) RoundTrip(req *http.Request) (resp *http.Response, err error) {
	t.req = req
	return nil, errors.New("dummy impl")
}

func TestSubmit(t *testing.T) {
	server := mockServer(loginHandler)
	defer server.Close()
	username := "foo"
	password := "bar"
	school := "baz"
	tr := &recordingTransport{}
	client, err := newClient(username, password, school)
	if err != nil {
		t.Fatal("Expected to get a new client", "got", err)
	}
	client.URL, _ = url.Parse(server.URL)
	client.client = &http.Client{Transport: tr}

	client.submit()
	if name := tr.req.FormValue("login_name"); name != username {
		t.Error("Expected", username, "got", name)
	}
}

//test that a new Client is returned correctly
func TestNewClient(t *testing.T) {
	username := "foo"
	password := "bar"
	school := "baz"
	urlStr := fmt.Sprintf("https://axiom.veracross.com/%s/", school)
	schoolURL, err := url.Parse(urlStr)
	if err != nil {
		t.Fatal(err)
	}
	//token := "Pl0opv2ylPnZWrWtSon9siyOL5hjhpZWPFm5nsQtikY="

	client, err := newClient(username, password, school)
	if err != nil {
		t.Fatal("Expected to get a new client", "got", err)
	}

	t.Log("Testing the Client struct")
	if client.Username != username {
		t.Fatal("Expected", username, "got", client.Username)
	}
	if client.Password != password {
		t.Fatal("Expected", password, "got", client.Password)
	}
	if client.School != school {
		t.Fatal("Expected", school, "got", client.School)
	}
	if client.URL.Path != schoolURL.Path {
		t.Fatal("Expected", schoolURL.Path, "got", client.URL.Path)
	}

	// loggedInClient, err := NewClient(username, password, school)
	// if err != nil {
	// 	t.Fatal("Expected to get a new client", "got", err)
	// }
	// if loggedInClient.Token != token {
	// 	t.Error("Expected", token, "got", loggedInClient.Token)
	// }
	// test to see if url.Parse in newClient() returns an error.
	invalidURLSchool := `WRTH #R$TH#%$T //foo?.xxx://`
	_, err = newClient(username, password, invalidURLSchool)
	if err == nil {
		t.Fatal("Expected newClient to return an error, invalidURL")
	}
	_, err = NewClient(username, password, invalidURLSchool)
	if err == nil {
		t.Fatal("Expected newClient to return an error, invalidURL")
	}
}

// Test loginToken function
func TestLoginToken(t *testing.T) {
	server := mockServer(loginHandler)
	defer server.Close()
	token := "Pl0opv2ylPnZWrWtSon9siyOL5hjhpZWPFm5nsQtikY="
	loginURL := fmt.Sprintf(server.URL + "/login")
	username := "foo"
	password := "bar"
	school := "baz"

	client, err := newClient(username, password, school)
	if err != nil {
		t.Fatal(err)
	}
	err = client.loginToken(loginURL)
	if err != nil {
		t.Fatal(err)
	}

	err = client.loginToken(loginURL)
	if err != nil {
		t.Fatal(err)
	}

	if client.Token != token {
		t.Fatal("expected", token, "got", client.Token)
	}
	err = client.loginToken("fff")
	if err == nil {
		t.Fatal(err)
	}
}

func TestSetToken(t *testing.T) {
	server := mockServer(loginHandler)
	defer server.Close()
	token := "Pl0opv2ylPnZWrWtSon9siyOL5hjhpZWPFm5nsQtikY="

	client := &Client{
		Username: "foo",
		Password: "bar",
		School:   "baz",
	}

	t.Log("testing setToken()")
	resp, _ := http.Get(server.URL)
	err := client.setToken(resp)
	if err != nil {
		t.Fatal(err)
	}
	if client.Token != token {
		t.Fatal("Expected", token, "got", client.Token)
	}

	helloServer := mockServer(helloHandler)
	defer helloServer.Close()

	client.Token = ""
	resp, _ = http.Get(helloServer.URL)
	err = client.setToken(resp)
	if err == nil {
		t.Error("if there is no token set, setToken should return an error, got nothing")
	}
}
