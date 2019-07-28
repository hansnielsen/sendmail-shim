sendmail-shim
=============

This is a small utility that logs outgoing email directly to a file. No SMTP, mail servers, or networking is involved.

Useful when you have a centralized log collector and want to keep records of outbound email in a scalable and searchable fashion.

See https://www.stackallocated.com/blog/2019/stop-using-smtp for more words about this.

Building and installing
-----------------------
This program only uses the Go standard library and should work with older versions of Go. It has been tested on Go 1.11 and later.
```
$ go build shim.go
$ sudo -s
# chown root:root shim
# chmod u+s shim
# mv shim /usr/sbin/sendmail
```

If you want to change the log filename, edit the line in `shim.go`.

Running
-------
```
$ /usr/sbin/sendmail foo@example.com
Subject: Test

Hello World
^D
$ tail -1 /var/log/sendmail-shim.log.json
{"time":"2019-07-28T09:25:47Z","uid":"501","username":"hans","arguments":["foo@example.com"],"body":"Subject: Test\n\nHello World\n"}
$
```

Testing
-------
```
$ go test
```
