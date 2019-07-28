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

// Returns (uid, username)
type UsernameFunc func() (string, string)

func GetUsername() (uid string, username string) {
	// get calling user ID and name
	u, err := user.Current()
	if err == nil {
		return u.Uid, u.Username
	}

	// just return the user ID
	return fmt.Sprintf("%d", os.Getuid()), ""
}

type TimeFunc func() string

func GetTime() string {
	return time.Now().UTC().Format(time.RFC3339)
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

func PopulateEntry(e *LogEntry, r io.Reader, uf UsernameFunc, tf TimeFunc) *LogError {
	// set the time
	e.Time = tf()

	// populate the uid and username
	uid, username := uf()
	e.UserID = uid
	e.Username = username

	// just use the full arguments list minus the program name
	e.Arguments = os.Args[1:]

	// read stdin
	body, err := ioutil.ReadAll(r)
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

	// build the log entry
	entry := LogEntry{}
	err = PopulateEntry(&entry, os.Stdin, GetUsername, GetTime)
	if err != nil {
		return err
	}

	// write out JSON
	err = EncodeJSON(f, entry)
	if err != nil {
		return err
	}

	// emit success metrics here if you want!
	//
	// metrics.Increment("sendmail-shim.success", 1, map[string]string{"uid": entry.UserID})

	return nil
}

func main() {
	err := EmitLog()
	if err != nil {
		// emit failure metrics here if you want!
		//
		// metrics.Increment("sendmail-shim.error", 1, map[string]string{"reason": err.Tag})

		log.Fatal(err.Err)
	}
}
