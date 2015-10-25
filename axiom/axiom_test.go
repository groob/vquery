package axiom

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"testing"
)

var client = &Client{
	Username: "Foo",
	Password: "Bar",
	School:   "Baz",
}

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
<meta content="Pl0opv2ylPnZWrWtSon9siyOL5hjhpZWPFm5nsQtikY=" name="csrf-token" />

    <script type="text/javascript">
        $.ajaxSetup({
            headers: {
                'X-CSRF-Token': $('meta[name="csrf-token"]').attr('content'),
            },
        });
    </script>
</head>
`

func mockServer() *httptest.Server {
	f := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprintln(w, head)
	}
	return httptest.NewServer(http.HandlerFunc(f))
}

func TestSetToken(t *testing.T) {
	server := mockServer()
	defer server.Close()

	resp, err := http.Get(server.URL)
	if err != nil {
		t.Fatal("mock server should respond to get request", err)
	}
	defer resp.Body.Close()

	err = client.setToken(resp)
	if err != nil {
		t.Fatal("Should be able to process html")
	}

	if client.Token != "Pl0opv2ylPnZWrWtSon9siyOL5hjhpZWPFm5nsQtikY=" {
		t.Fatal("Token not set")
		fmt.Println(client)
	}
}

func TestDo(t *testing.T) {
	server := mockServer()
	defer server.Close()

	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatal("Cooke jar not set", err)
	}
	client.Token = "Pl0opv2ylPnZWrWtSon9siyOL5hjhpZWPFm5nsQtikY="
	client.client = &http.Client{Jar: jar}
	req, err := http.NewRequest("GET", server.URL, nil)
	if err != nil {
		t.Fatal("Should be able to set http NewRequest", err)
	}
	resp, err := client.Do(req)
	if err != nil {
		t.Fatal("Should be able to mock http request", err)
	}
	defer resp.Body.Close()
}
