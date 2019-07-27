package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

var LogFilePath = "/var/log/sendmail-shim.log.json"

type LogEntry struct {
	Arguments []string `json:"arguments"`
	Body      string   `json:"body"`
}

func main() {
	// read stdin
	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		log.Fatalf("couldn't read stdin: %v", err)
	}

	// build JSON
	entry := LogEntry{
		Arguments: os.Args,
		Body:      string(body),
	}

	// write out the JSON object on a line by itself
	f, err := os.OpenFile(LogFilePath, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		log.Fatalf("couldn't open log file: %v", err)
	}
	defer f.Close()

	j := json.NewEncoder(f)
	err = j.Encode(entry)
	if err != nil {
		log.Fatalf("couldn't encode JSON: %v", err)
	}
}
