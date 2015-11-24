package main

import (
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
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
	Name string `toml:"name"`
	ID   int    `toml:"id"`
}
type config struct {
	Interval          time.Duration `toml:"interval"`
	VeracrossUsername string        `toml:"veracross_username"`
	VeracrossPassword string        `toml:"veracross_password"`
	VeracrossSchool   string        `toml:"veracross_school"`
	Reports           []report      `toml:"reports"`
	ReportsPath       string        `toml:"reports_path"`
}

func init() {
	flag.Parse()

	if *fVersion {
		fmt.Printf("query - version %s\n", Version)
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
	//save report
	err = saveReport(body, name)
	if err != nil {
		log.Println(err)
		return
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
	for _, report := range conf.Reports {
		wg.Add(1)
		go runReport(report.ID, report.Name)
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
	done := make(chan bool)
	ticker := time.NewTicker(time.Minute * conf.Interval).C
	for {
		go run(done)
		<-done
		<-ticker
	}
}
