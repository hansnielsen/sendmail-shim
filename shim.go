package main

import (
	"encoding/json"
	"fmt"
	"io"
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

type LogError struct {
	Err error
	Tag string
}

func OpenLogFile(path string) (*os.File, *LogError) {
	// write out the JSON object on a line by itself
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, &LogError{
			fmt.Errorf("couldn't open log file: %v", err),
			"open-log-file",
		}
	}
	return f, nil
}

func PopulateEntry(e *LogEntry) *LogError {
	// set the time
	e.Time = time.Now().UTC().Format(time.RFC3339)

	// get calling user ID and name
	u, err := user.Current()
	if err == nil {
		e.UserID = u.Uid
		e.Username = u.Username
	} else {
		// just fill in the user ID
		e.UserID = fmt.Sprintf("%d", os.Getuid())
	}

	// just use the full arguments list minus the program name
	e.Arguments = os.Args[1:]

	// read stdin
	body, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		return &LogError{
			fmt.Errorf("couldn't read stdin: %v", err),
			"stdin-failed",
		}
	}
	e.Body = string(body)

	return nil
}

func EncodeJSON(f io.Writer, e LogEntry) *LogError {
	j := json.NewEncoder(f)
	err := j.Encode(e)
	if err != nil {
		return &LogError{
			fmt.Errorf("couldn't encode JSON: %v", err),
			"json-encoding",
		}
	}
	return nil
}

func EmitLog() *LogError {
	// open the log file
	f, err := OpenLogFile(LogFilePath)
	if err != nil {
		return err
	}
	defer f.Close()

	// build JSON
	entry := LogEntry{}
	err = PopulateEntry(&entry)
	if err != nil {
		return err
	}

	// write out JSON
	err = EncodeJSON(f, entry)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	err := EmitLog()
	if err != nil {
		log.Fatal(err.Err)
	}
}
