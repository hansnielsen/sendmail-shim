package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"testing"
)

func TestOpenLogFile(t *testing.T) {
	// test file open failure
	f, lerr := OpenLogFile("")
	if f != nil {
		t.Error("shouldn't open an empty filename")
	}
	if lerr.Tag != "open-log-file" {
		t.Errorf("invalid error tag %q", lerr.Tag)
	}
	if lerr.Err == nil {
		t.Error("should have an error")
	}

	dir, err := ioutil.TempDir("", "log-file-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)

	// test creating a file
	n := path.Join(dir, "test.log")
	f, lerr = OpenLogFile(n)
	if lerr != nil {
		t.Fatal(lerr.Err)
	}
	if f == nil {
		t.Fatal("expected a file")
	}
	_, err = f.WriteString("foo\n")
	f.Close()

	// test writing to an existing file
	f, lerr = OpenLogFile(n)
	if lerr != nil {
		t.Fatal(lerr.Err)
	}
	if f == nil {
		t.Fatal("expected a file")
	}
	_, err = f.WriteString("bar\n")
	f.Close()

	// verify that the file has the right stuff in it
	contents, err := ioutil.ReadFile(n)
	if err != nil {
		t.Fatal(err)
	}
	expected := "foo\nbar\n"
	if string(contents) != expected {
		t.Errorf("expected %q, got %q", expected, contents)
	}
}

type ErrorWriter struct{}

func (_ ErrorWriter) Write(_ []byte) (int, error) {
	return 0, fmt.Errorf("oh no")
}

func TestEncodeJSON(t *testing.T) {
	// test JSON encoding errors
	l1 := EmailLogger{
		Writer: ErrorWriter{},
	}
	lerr := l1.EncodeJSON(LogEntry{})
	if lerr == nil {
		t.Fatal("expected an error")
	}
	if lerr.Tag != "json-encoding" {
		t.Fatalf("unexpected tag %q", lerr.Tag)
	}

	// test expected encoding
	e := LogEntry{
		Time:      "2009-11-10T23:00:00Z",
		UserID:    "123",
		Username:  "foo",
		Arguments: []string{"yay", "asdf"},
		Body:      "stuff",
	}
	expected := "{\"time\":\"2009-11-10T23:00:00Z\",\"uid\":\"123\",\"username\":\"foo\",\"arguments\":[\"yay\",\"asdf\"],\"body\":\"stuff\"}\n"
	b := bytes.Buffer{}
	l2 := EmailLogger{
		Writer: &b,
	}
	lerr = l2.EncodeJSON(e)
	if lerr != nil {
		t.Fatal(lerr.Err)
	}
	actual := b.String()
	if expected != actual {
		t.Fatalf("expected %q, got %q", expected, actual)
	}
}

type ErrorReader struct{}

func (_ ErrorReader) Read(_ []byte) (int, error) {
	return 0, fmt.Errorf("oh no")
}

func ConstUsername() (string, string) {
	return "123", "foobar"
}

func ConstTime() string {
	return "2009-11-10T23:00:00Z"
}

func TestPopulateEntry(t *testing.T) {
	// test stdin read failure
	l1 := EmailLogger{
		Body: ErrorReader{},
		User: ConstUsername,
		Time: ConstTime,
	}
	e1 := LogEntry{}
	lerr := l1.Populate(&e1)
	if lerr == nil {
		t.Fatal("expected an error")
	}
	if lerr.Tag != "stdin-failed" {
		t.Fatalf("unexpected tag %q", lerr.Tag)
	}

	// test entry population
	l2 := EmailLogger{
		Body: strings.NewReader("hello"),
		User: ConstUsername,
		Time: ConstTime,
		Args: []string{"yay", "stuff"},
	}
	e2 := LogEntry{}
	lerr = l2.Populate(&e2)
	if lerr != nil {
		t.Fatal(lerr.Err)
	}
	if e2.Time != "2009-11-10T23:00:00Z" {
		t.Errorf("bad time %q", e2.Time)
	}
	if e2.UserID != "123" {
		t.Errorf("bad user ID %q", e2.UserID)
	}
	if e2.Username != "foobar" {
		t.Errorf("bad username %q", e2.Username)
	}
	if e2.Body != "hello" {
		t.Errorf("bad body %q", e2.Body)
	}
	if a := e2.Arguments; len(a) != 2 || a[0] != "yay" || a[1] != "stuff" {
		t.Errorf("bad arguments %v", e2.Arguments)
	}
}

func TestEmit(t *testing.T) {
	dir, err := ioutil.TempDir("", "emit-test")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dir)
	n := path.Join(dir, "test.log.json")

	l := EmailLogger{
		LogPath: n,
		Args:    []string{"fro", "bozz"},
		Body:    strings.NewReader("hello\nworld\n"),
		User:    ConstUsername,
		Time:    ConstTime,
	}
	lerr := l.Emit()
	if lerr != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(n)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{\"time\":\"2009-11-10T23:00:00Z\",\"uid\":\"123\",\"username\":\"foobar\",\"arguments\":[\"fro\",\"bozz\"],\"body\":\"hello\\nworld\\n\"}\n"
	if string(b) != expected {
		t.Fatalf("got %q, expected %q", b, expected)
	}
}
