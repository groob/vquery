Veracross Axiom Web Scraper Library

For the ESWeb version, see [v0.0.2](https://github.com/groob/vquery/releases/tag/v0.0.2) and v0.0.2 [blog post](http://groob.io/posts/my-private-api/)

Package axiom allows scraping any Veracross Query.

# Example

```go
	// Create a axiom client (a logged in session)
	client, err := axiom.NewAxiomClient(username, password, school)
	if err != nil {
		log.Fatal(err)
	}

	// create a http request
	
	req, err := http.NewRequest("POST", "https://axiom.veracross.com/school/query/123456/result_data.json", nil)
	
	// Do request
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	
	// do something with resp, which will be a json stream
	jsonBody , err := ioutil.ReadAll(resp.Body)

```
