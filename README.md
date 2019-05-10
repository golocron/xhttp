# xhttp

[![GoDoc](https://godoc.org/github.com/golocron/xhttp?status.svg)](https://godoc.org/github.com/golocron/xhttp) [![Go Report Card](https://goreportcard.com/badge/github.com/golocron/xhttp)](https://goreportcard.com/report/github.com/golocron/xhttp)

xhttp provides a custom http client based on net/http.

## Description

The package provides an http client with the following features:

- Customized timeouts
- Proper handling of response Body
- Method for file downloading
- Ability to send plain `net/http.Request`
- Convenient methods to set up headers of a request


## Install

```bash
go get github.com/golocron/xhttp
```

## Examples

Examples can be found [here](examples/examples.go)

```bash
go run examples/examples.go

2019/05/10 21:03:52 http request succeeded: method GET, body:
{
  "args": {},
  "headers": {
    "Accept-Encoding": "gzip",
    "Host": "httpbin.org",
    "User-Agent": "Go-http-client/1.1"
  },
  "origin": "127.0.0.1",
  "url": "https://httpbin.org/get"
}
```
