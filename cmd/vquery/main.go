package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/groob/vquery/axiom"
)

var (
	// Version returns the tool version version
	Version   = "unreleased"
	fVersion  = flag.Bool("version", false, "display the version")
	fConfig   = flag.String("config", "", "configuration file to load")
	fReportID = flag.Int("report", 0, "use report ID")
	conf      config
	client    *axiom.Client
	wg        sync.WaitGroup
)

type report struct {
	Name        string      `toml:"name"`
	ID          int         `toml:"id"`
	Format      string      `toml:"format"`
	Keys        stringArray `toml:"keys"`
	PrintHeader bool        `toml:"print_header"`
}

type config struct {
	Interval          duration `toml:"interval"`
	VeracrossUsername string   `toml:"veracross_username"`
	VeracrossPassword string   `toml:"veracross_password"`
	VeracrossSchool   string   `toml:"veracross_school"`
	Reports           []report `toml:"reports"`
	ReportsPath       string   `toml:"reports_path"`
}

type duration struct {
	time.Duration
}

func (d *duration) UnmarshalText(text []byte) error {
	var err error
	d.Duration, err = time.ParseDuration(string(text))
	return err
}

func init() {
	flag.Parse()

	if *fVersion {
		fmt.Printf("vquery - version %s\n", Version)
		os.Exit(0)
	}

	if _, err := toml.DecodeFile(*fConfig, &conf); err != nil {
		log.Fatal(err)
	}

	c, err := axiom.NewClient(conf.VeracrossUsername, conf.VeracrossPassword,
		conf.VeracrossSchool)
	if err != nil {
		log.Fatal(err)
	}
	client = c
}

func saveReport(jsonData []byte, name string) error {
	jsonFile, err := os.Create(conf.ReportsPath + "/" + name + ".json")
	if err != nil {
		return err
	}
	defer jsonFile.Close()
	_, err = jsonFile.Write(jsonData)
	if err != nil {
		return err
	}
	return nil
}

// runs a report from Veracross and saves the JSON file localy
func runReport(reportID int, name string) {
	defer wg.Done()
	resp, err := client.Report(reportID)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	log.Printf("Saving axiom report %s", name)
	//save report
	err = saveReport(body, name)
	if err != nil {
		log.Println(err)
		return
	}
}

func saveCSVReport(r report) {
	defer wg.Done()
	if len(r.Keys) == 0 {
		log.Println("must specify keys for csv header")
		return
	}
	csvFile, err := os.Create(conf.ReportsPath + "/" + r.Name + ".csv")
	if err != nil {
		log.Println(err)
		return
	}
	defer csvFile.Close()
	log.Printf("Saving axiom report as csv file %s.csv", r.Name)
	runCSVReport(r, csvFile)
}

func runCSVReport(r report, output io.Writer) {
	var csvJSON []map[string]interface{}
	resp, err := client.Report(r.ID)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&csvJSON)
	if err != nil {
		log.Println(err)
		return
	}
	var expandedKeys [][]string
	for _, key := range r.Keys {
		expandedKeys = append(expandedKeys, strings.Split(key, "."))
	}
	var w *csv.Writer
	w = csv.NewWriter(output)
	if r.PrintHeader {
		w.Write(r.Keys)
		w.Flush()
	}
	for _, data := range csvJSON {
		var record []string
		for _, expandedKey := range expandedKeys {
			record = append(record, getValue(data, expandedKey))
		}
		w.Write(record)
		w.Flush()
	}
}

// prints a report to stdout
func printReport(id int) {
	resp, err := client.Report(id)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		log.Println(err)
	}
}

func run(done chan bool) {
	for _, r := range conf.Reports {
		wg.Add(1)
		if r.Format != "csv" {
			go runReport(r.ID, r.Name)
		} else {
			go saveCSVReport(r)
		}
	}
	wg.Wait()
	done <- true
}

func main() {
	// print a single report and exit
	if *fReportID != 0 {
		printReport(*fReportID)
		os.Exit(0)
	}
	//run as a daemon, saving reports from config
	// interval := time.Minute * conf.Interval
	interval := conf.Interval.Duration
	done := make(chan bool)
	ticker := time.NewTicker(interval).C
	for {
		go run(done)
		<-done
		log.Printf("Sleeping for %s", interval)
		<-ticker
	}
}
