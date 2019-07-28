package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"os/user"
	"time"
)

var LogFilePath = "/var/log/sendmail-shim.log.json"

type LogEntry struct {
	Time      string   `json:"time"`
	UserID    string   `json:"uid"`
	Username  string   `json:"username,omitempty"`
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
		Arguments: os.Args[1:],
		Body:      string(body),
		Time:      time.Now().UTC().Format(time.RFC3339),
	}

	// get calling user ID and name
	u, err := user.Current()
	if err == nil {
		entry.UserID = u.Uid
		entry.Username = u.Username
	} else {
		// just fill in the user ID
		entry.UserID = fmt.Sprintf("%d", os.Getuid())
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
