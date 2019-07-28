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
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_APPEND|os.O_CREATE, 0644)
	if err != nil {
		return nil, &LogError{
			fmt.Errorf("couldn't open log file: %v", err),
			"open-log-file",
		}
	}
	return f, nil
}

type EmailLogger struct {
	Args    []string
	Body    io.Reader
	User    UsernameFunc
	Time    TimeFunc
	Writer  io.Writer
	LogPath string // used if there's no Writer
}

func (l *EmailLogger) Populate(e *LogEntry) *LogError {
	// set the time
	e.Time = l.Time()

	// populate the uid and username
	uid, username := l.User()
	e.UserID = uid
	e.Username = username

	// just use the full arguments list minus the program name
	e.Arguments = l.Args

	// read stdin
	body, err := ioutil.ReadAll(l.Body)
	if err != nil {
		return &LogError{
			fmt.Errorf("couldn't read stdin: %v", err),
			"stdin-failed",
		}
	}
	e.Body = string(body)

	return nil
}

func (l *EmailLogger) EncodeJSON(e LogEntry) *LogError {
	j := json.NewEncoder(l.Writer)
	err := j.Encode(e)
	if err != nil {
		return &LogError{
			fmt.Errorf("couldn't encode JSON: %v", err),
			"json-encoding",
		}
	}
	return nil
}

func (l *EmailLogger) Emit() *LogError {
	if l.Writer == nil {
		// open the log file if there's no writer
		f, err := OpenLogFile(l.LogPath)
		if err != nil {
			return err
		}
		l.Writer = f
		defer func() {
			l.Writer = nil
			f.Close()
		}()
	}

	// build the log entry
	entry := LogEntry{}
	err := l.Populate(&entry)
	if err != nil {
		return err
	}

	// write out JSON
	err = l.EncodeJSON(entry)
	if err != nil {
		return err
	}

	// emit success metrics here if you want!
	//
	// metrics.Increment("sendmail-shim.success", 1, map[string]string{"uid": entry.UserID})

	return nil
}

func main() {
	l := EmailLogger{
		LogPath: "/var/log/sendmail-shim.log.json",
		Args:    os.Args[1:],
		Body:    os.Stdin,
		User:    GetUsername,
		Time:    GetTime,
	}
	err := l.Emit()
	if err != nil {
		// emit failure metrics here if you want!
		//
		// metrics.Increment("sendmail-shim.error", 1, map[string]string{"reason": err.Tag})

		log.Fatal(err.Err)
	}
}
