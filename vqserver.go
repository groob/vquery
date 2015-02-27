package main

import (
	"bytes"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"os/exec"
)

func main() {
	rtr := mux.NewRouter()
	rtr.HandleFunc("/v1/{id:[0-9]+}", vquery).Methods("GET")

	http.Handle("/", rtr)

	log.Println("Listening...")
	http.ListenAndServe(":3000", nil)
}

func runVQuery(id string) (string, error) {
	// Execs to ruby and runs the vquery script
	log.Printf("Executing %v...", id)
	vqueryCmd := exec.Command("ruby", "vquery/vquery.rb", id)
	var out, stderr bytes.Buffer
	vqueryCmd.Stdout = &out
	vqueryCmd.Stderr = &stderr
	err := vqueryCmd.Run()
	if err != nil {
		log.Printf("%v", stderr.String())
		return "", err
	}
	return out.String(), nil
}

func vquery(w http.ResponseWriter, r *http.Request) {
	// Serves the results of vquery over HTTP
	params := mux.Vars(r)
	id := params["id"]
	json_output, err := runVQuery(id)
	if err != nil {
		log.Println(err)
		w.Write([]byte("There was an error. Check the logs"))
	} else {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(json_output))
	}
}
