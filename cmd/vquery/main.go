package main

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
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
	logger    log.Logger
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

	logger = log.NewLogfmtLogger(os.Stderr)

	if _, err := toml.DecodeFile(*fConfig, &conf); err != nil {
		level.Info(logger).Log(
			"msg", "decode config file",
			"err", err,
			"path", *fConfig,
		)
		os.Exit(1)
	}

	c, err := axiom.NewClient(
		conf.VeracrossUsername,
		conf.VeracrossPassword,
		conf.VeracrossSchool,
		axiom.WithLogger(logger),
	)
	if err != nil {
		level.Info(logger).Log(
			"msg", "create axiom client",
			"err", err,
			"school", conf.VeracrossUsername,
		)
		os.Exit(1)
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
func runReport(reportID int, name string, logger log.Logger) {
	defer wg.Done()
	resp, err := client.Report(reportID)
	if err != nil {
		level.Info(logger).Log("err", err)
		return
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		level.Info(logger).Log("err", err)
		return
	}
	level.Debug(logger).Log("msg", "saving axiom report", "report", name, "id", reportID)
	//save report
	err = saveReport(body, name)
	if err != nil {
		level.Info(logger).Log("err", err)
		return
	}
}

func saveCSVReport(r report, logger log.Logger) {
	defer wg.Done()
	if len(r.Keys) == 0 {
		level.Info(logger).Log("err", "must specify keys for csv header")
		return
	}
	csvFile, err := os.Create(conf.ReportsPath + "/" + r.Name + ".csv")
	if err != nil {
		level.Info(logger).Log("err", err)
		return
	}
	defer csvFile.Close()
	level.Debug(logger).Log("msg", "saving axiom report as csv file", "file", csvFile.Name())
	runCSVReport(r, csvFile, logger)
}

func runCSVReport(r report, output io.Writer, logger log.Logger) {
	var csvJSON []map[string]interface{}
	resp, err := client.Report(r.ID)
	if err != nil {
		level.Info(logger).Log("err", err)
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&csvJSON)
	if err != nil {
		level.Info(logger).Log("err", err)
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
func printReport(id int, logger log.Logger) {
	resp, err := client.Report(id)
	if err != nil {
		level.Info(logger).Log("err", err)
		return
	}
	defer resp.Body.Close()
	_, err = io.Copy(os.Stdout, resp.Body)
	if err != nil {
		level.Info(logger).Log("err", err)
	}
}

func run(done chan bool, logger log.Logger) {
	for _, r := range conf.Reports {
		wg.Add(1)
		if r.Format != "csv" {
			go runReport(r.ID, r.Name, logger)
		} else {
			go saveCSVReport(r, logger)
		}
	}
	wg.Wait()
	done <- true
}

func main() {
	// print a single report and exit
	if *fReportID != 0 {
		printReport(*fReportID, logger)
		os.Exit(0)
	}

	//run as a daemon, saving reports from config
	// interval := time.Minute * conf.Interval
	interval := conf.Interval.Duration
	done := make(chan bool)
	ticker := time.NewTicker(interval).C
	for {
		go run(done, logger)
		<-done
		level.Debug(logger).Log("msg", "sleep until ticker", "duration", interval)
		<-ticker
	}
}
