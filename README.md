Veracross Axiom Web Scraper

A command line utility to print/save Veracross reports without logging in to Axiom. 
This tool allows someone to save a report from Axiom's workspace at a specified interval(see config file sample).

# How it works:
Print json output to screen. 
```bash
vquery -config config.toml -report 12345
```

Not specifying a report parameter will save the reports specified in `config.toml` to a destination on disk. 
```bash
vquery -config config.toml
```


# Use as a Go Library

If you'd like to use the code as a Go library to extend this proggram, the documentation is available on [godoc](https://godoc.org/github.com/groob/vquery/axiom)

```Go
package main

func main() {
    // Create a new http client logged into axiom.
	client, err := axiom.NewClient(username, password, school)
	if err != nil {
		log.Fatal(err)
	}

    // the JSON output of the axiom report will be saved to 
    // the resp.Body
    resp, err := client.Report(reportID)
    if err != nil {
        //handle err
    }
	defer resp.Body.Close()
    // Print response to stdout
    io.Copy(os.Stdout, resp.Body)
}
```
