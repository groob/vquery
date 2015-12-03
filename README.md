Veracross Axiom Web Scraper

A command line utility to print/save Veracross reports without logging in to Axiom. 
This tool allows someone to save a report from Axiom's workspace at a specified interval(see config file sample).

# How it works:
Print json output to screen. 
```bash
vquery -config config.toml -report 12345
```

Not specifying a report parameter will save the reports specified in `config.toml` to a destination on disk. 
The utility will run continuously and refresh the report at the interval specified in the config(in minutes).
```bash
vquery -config config.toml
```

# Saving a CSV file
To save as a CSV, you can combine the json output with additional command like tools, like [jq](https://stedolan.github.io/jq/) and [json2csv](https://github.com/jehiah/json2csv)

Example:
The example below will save report #163998 as a csv file, using the `person_id`, `first_name`, `last_name` and `email_1` fields in the report. 
```bash
vquery -config config.toml -report 163998 | jq -c '.[]' | json2csv -k person_id,first_name,last_name,email_1 > people.csv
```
We can do further work like, sorting the results alphabetically by `last_name` first. 

```bash
vquery -config config.toml -report 163998 | jq -c '.[]' |json2csv -k person_id,first_name,last_name,email_1 | sort -k 3,3 -t,
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
