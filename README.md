# t411-client

[![Go Report Card](https://goreportcard.com/badge/github.com/dns-gh/t411-client)](https://goreportcard.com/report/github.com/dns-gh/t411-client)

[![GoDoc](https://godoc.org/github.com/dns-gh/t411-client/t411client?status.png)]
(https://godoc.org/github.com/dns-gh/t411-client/t411client)

t411-client is a Go web client for the t411 website API: https://api.t411.li/

## Motivation

TODO

## Installation

- It requires Go language of course. You can set it up by downloading it here: https://golang.org/dl/
- Install it here C:/Go.
- Set your GOPATH, GOROOT and PATH environment variables:

```
export GOROOT=C:/Go
export GOPATH=WORKING_DIR
export PATH=C:/Go/bin:${PATH}
```

- Download and install the package:

```
@working_dir $ go get github.com/dns-gh/t411-client/...
@working_dir $ go install github.com/dns-gh/t411-client/t411client
```

## Example

TODO

## Tests

For example:
```
@working_dir $ export T411_USERNAME=your_username && export T411_PASSWORD=your_password && go test ...t411client -gocheck.vv -test.v -gocheck.f Test
=== RUN   Test
START: t411client_test.go:27: MySuite.TestNewT411
PASS: t411client_test.go:27: MySuite.TestNewT411        0.538s

START: torrents_test.go:7: MySuite.TestReqURL
PASS: torrents_test.go:7: MySuite.TestReqURL    0.001s

START: users_test.go:5: MySuite.TestUsersProfile
PASS: users_test.go:5: MySuite.TestUsersProfile 0.171s

OK: 3 passed
--- PASS: Test (0.71s)
PASS
ok      github.com/dns-gh/t411-client/t411client        0.877s
```

## LICENSE

See included LICENSE file.