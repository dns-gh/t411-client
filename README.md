# t411-client

[![Go Report Card](https://goreportcard.com/badge/github.com/dns-gh/t411-client)](https://goreportcard.com/report/github.com/dns-gh/t411-client)

[![GoDoc](https://godoc.org/github.com/dns-gh/t411-client/t411client?status.png)]
(https://godoc.org/github.com/dns-gh/t411-client/t411client)

t411-client is a Go web client for the t411 website API: https://api.t411.li/

## Motivation

Used in https://github.com/dns-gh/torrents-bot

Feel free to join my efforts!

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
START: categories_test.go:23: MySuite.TestCategoriesTree
one category with no id found
PASS: categories_test.go:23: MySuite.TestCategoriesTree 0.336s

START: torrents_test.go:147: MySuite.TestDownloadTorrentByID
PASS: torrents_test.go:147: MySuite.TestDownloadTorrentByID     1.102s

START: torrents_test.go:165: MySuite.TestDownloadTorrentByTerms
PASS: torrents_test.go:165: MySuite.TestDownloadTorrentByTerms  0.925s

START: torrents_test.go:14: MySuite.TestMakeURL
PASS: torrents_test.go:14: MySuite.TestMakeURL  0.000s

START: t411client_test.go:27: MySuite.TestNewT411
PASS: t411client_test.go:27: MySuite.TestNewT411        0.803s

START: torrents_test.go:94: MySuite.TestSearchAllTorrents
PASS: torrents_test.go:94: MySuite.TestSearchAllTorrents        0.887s

START: torrents_test.go:63: MySuite.TestSearchTorrentsByTerms
PASS: torrents_test.go:63: MySuite.TestSearchTorrentsByTerms    2.809s

START: torrents_test.go:105: MySuite.TestSearchTorrentsByTermsComplete
PASS: torrents_test.go:105: MySuite.TestSearchTorrentsByTermsComplete   0.727s

START: torrents_test.go:130: MySuite.TestSearchTorrentsSortingBySeeders
PASS: torrents_test.go:130: MySuite.TestSearchTorrentsSortingBySeeders  0.469s

START: terms_test.go:24: MySuite.TestTermsTree
PASS: terms_test.go:24: MySuite.TestTermsTree   0.278s

START: torrents_test.go:179: MySuite.TestTorrentsDetails
PASS: torrents_test.go:179: MySuite.TestTorrentsDetails 1.007s

START: users_test.go:7: MySuite.TestUsersProfile
PASS: users_test.go:7: MySuite.TestUsersProfile 1.280s

OK: 12 passed
--- PASS: Test (10.65s)
PASS
ok      github.com/dns-gh/t411-client/t411client        10.897s
```

## LICENSE

See included LICENSE file.