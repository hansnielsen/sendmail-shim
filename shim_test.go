package main

import (
	"io/ioutil"
	"os"
	"path"
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
