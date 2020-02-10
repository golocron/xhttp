package main

import (
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/golocron/xhttp"
	"github.com/pkg/errors"
)

const (
	baseURL = "https://httpbin.org"
)

func main() {
	// Make a request using the default client.
	if err := example1(); err != nil {
		log.Printf("http request failed: %s", err)
	}

	// Download and save a file.
	if err := example2(); err != nil {
		log.Printf("http request failed: %s", err)
	}
}

func example1() error {
	url := strings.Join([]string{baseURL, "get"}, "/")
	resp, err := xhttp.GET(url)
	if err != nil {
		return errors.Wrapf(err, "http request failed: method %s, url %s", http.MethodGet, url)
	}

	if resp.StatusCode != 200 {
		return errors.Errorf("http request failed: code %d, status %s", resp.StatusCode, resp.Status)
	}

	log.Printf("http request succeeded: method %s, body:\n%s", http.MethodGet, string(resp.Body))

	return nil
}

func example2() error {
	dir, err := os.UserHomeDir()
	if err != nil {
		return errors.Wrap(err, "failed to determine home directory")
	}

	url := strings.Join([]string{baseURL, "image/jpeg"}, "/")

	if err := xhttp.DownloadFile(url, filepath.Join(dir, "Downloads", "xhttp_example2.jpg")); err != nil {
		return errors.Wrapf(err, "http request failed: download file %s", url)
	}

	return nil
}
