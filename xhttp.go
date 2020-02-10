// Package xhttp provides a custom http client based on net/http.
package xhttp

import (
	"bytes"
	"crypto/tls"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// The defaults.
const (
	DefaultClientTimeout         = 30 * time.Second
	DefaultDialTimeout           = 10 * time.Second
	DefaultKeepAlive             = 30 * time.Second
	DefaultIdleConnTimeout       = 90 * time.Second
	DefaultTLSHandshakeTimeout   = 10 * time.Second
	DefaultExpectContinueTimeout = 1 * time.Second
	DefaultMaxIdleConns          = 100
)

var (
	// DefaultClient is used if no custom HTTP client is defined.
	DefaultClient *Client = NewClient()
)

// Do makes a request based on http.Request using DefaultClient.
func Do(req *http.Request) (*http.Response, error) {
	return DefaultClient.Do(req)
}

// Get makes a Get request using DefaultClient.
func Get(url string) (*http.Response, error) {
	return DefaultClient.Get(url)
}

// Head makes a Head request using DefaultClient.
func Head(url string) (*http.Response, error) {
	return DefaultClient.Head(url)
}

// Post makes a Post request using DefaultClient.
func Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return DefaultClient.Post(url, contentType, body)
}

// PostForm makes a PostForm request using DefaultClient.
func PostForm(url string, data url.Values) (*http.Response, error) {
	return DefaultClient.PostForm(url, data)
}

// GET makes a GET request using DefaultClient.
func GET(url string) (*Response, error) {
	return DefaultClient.GET(url)
}

// POST makes a POST request using DefaultClient.
func POST(url, contentType string, body []byte) (*Response, error) {
	return DefaultClient.POST(url, contentType, body)
}

// Send makes a request based on Request using DefaultClient.
func Send(req *Request) (*Response, error) {
	return DefaultClient.Send(req)
}

// DownloadFile downloads the file located at the url.
// It stores the result in a file with the filename.
func DownloadFile(url, filename string) error {
	return DefaultClient.DownloadFile(url, filename)
}

// ClientConfig holds the configuration for a Client.
type ClientConfig struct {
	Timeout               time.Duration
	DialTimeout           time.Duration
	KeepAlive             time.Duration
	IdleConnTimeout       time.Duration
	TLSHanshakeTimeout    time.Duration
	ExpectContinueTimeout time.Duration
	MaxIdleConns          int
	SkipTLSVerify         bool
	IncludeRootCA         bool
}

// Client represents a custom http client wrapper around net/http.Client.
type Client struct {
	HTTPClient http.Client
}

// NewClient returns a Client with customized default settings.
//
// More details here https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
func NewClient() *Client {
	return createClient(DefaultClientConfig())
}

// NewClientWithConfig creates a new Client with the settings specified in cfg.
func NewClientWithConfig(cfg *ClientConfig) *Client {
	return createClient(cfg)
}

// createClient creates a new Client using custom tls.Config and http.Transport.
func createClient(cfg *ClientConfig) *Client {
	// Create a custom tls config.
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipTLSVerify,
	}

	// TODO: Create a pool with root CA.
	// if customCACerts {}

	// Create a custom transport.
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   cfg.DialTimeout,
			KeepAlive: cfg.KeepAlive,
		}).DialContext,
		IdleConnTimeout:       cfg.IdleConnTimeout,
		TLSHandshakeTimeout:   cfg.TLSHanshakeTimeout,
		ExpectContinueTimeout: cfg.ExpectContinueTimeout,
		MaxIdleConns:          cfg.MaxIdleConns,
		TLSClientConfig:       tlsConfig,
	}

	// Finally, create a custom http client.
	client := &Client{
		HTTPClient: http.Client{
			Timeout:   cfg.Timeout,
			Transport: transport,
		},
	}

	return client
}

// DefaultClientConfig returns ClientConfig with default settings.
func DefaultClientConfig() *ClientConfig {
	return &ClientConfig{
		Timeout:               DefaultClientTimeout,
		DialTimeout:           DefaultDialTimeout,
		KeepAlive:             DefaultKeepAlive,
		IdleConnTimeout:       DefaultIdleConnTimeout,
		TLSHanshakeTimeout:    DefaultTLSHandshakeTimeout,
		ExpectContinueTimeout: DefaultExpectContinueTimeout,
		MaxIdleConns:          DefaultMaxIdleConns,
		SkipTLSVerify:         false,
		IncludeRootCA:         false,
	}
}

// Do performs a request based on http.Request.
func (c *Client) Do(req *http.Request) (*http.Response, error) {
	return c.HTTPClient.Do(req)
}

// Get makes a Get request.
func (c *Client) Get(url string) (*http.Response, error) {
	return c.HTTPClient.Get(url)
}

// Head makes a Head request.
func (c *Client) Head(url string) (*http.Response, error) {
	return c.HTTPClient.Head(url)
}

// Post makes a Post request.
func (c *Client) Post(url, contentType string, body io.Reader) (*http.Response, error) {
	return c.HTTPClient.Post(url, contentType, body)
}

// PostForm makes a PostForm request.
func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	return c.HTTPClient.PostForm(url, data)
}

// CloseIdleConnections calls to the underlying client's CloseIdleConnections.
func (c *Client) CloseIdleConnections() {
	c.HTTPClient.CloseIdleConnections()
}

// GET makes a Get request.
func (c *Client) GET(url string) (*Response, error) {
	req := NewRequest(http.MethodGet, url, nil)

	return c.Send(req)
}

// POST makes a Post request.
func (c *Client) POST(url, contentType string, body []byte) (*Response, error) {
	req := NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", contentType)

	return c.Send(req)
}

// Send makes a request.
func (c *Client) Send(request *Request) (*Response, error) {
	// Build the HTTP request object.
	req, err := c.buildRequest(request)
	if err != nil {
		return nil, err
	}

	// Build the HTTP client and make the request.
	resp, err := c.Do(req)
	if resp != nil {
		defer resp.Body.Close()
	}

	if err != nil {
		return nil, err
	}

	return c.buildResponse(resp)
}

// DownloadFile downloads the file located at an url and stores it in the given path.
func (c *Client) DownloadFile(url, filename string) error {
	resp, err := c.GET(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return fmt.Errorf("download failed: %s", resp.Status)
	}

	return ioutil.WriteFile(filename, resp.Body, 0644)
}

// buildRequest creates a http.Request from Request.
func (c *Client) buildRequest(req *Request) (*http.Request, error) {
	url := req.BaseURL
	if len(req.Param) != 0 {
		url = strings.Join([]string{req.BaseURL, "?", req.Param.Encode()}, "")
	}

	r, err := http.NewRequest(req.Method, url, bytes.NewBuffer(req.Body))
	if err != nil {
		return nil, err
	}

	if len(req.Header) != 0 {
		r.Header = req.Header
	}

	return r, nil
}

// buildResponse builds Response from http.Response.
//
// It takes care of closing the body.
func (c *Client) buildResponse(resp *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	response := &Response{
		StatusCode: resp.StatusCode,
		Status:     resp.Status,
		Body:       body,
		Headers:    resp.Header,
	}

	return response, nil
}

// Request defines a request.
type Request struct {
	Method  string
	BaseURL string
	Body    []byte
	Header  http.Header
	Param   url.Values
}

// NewRequest returns a Request ready for use.
func NewRequest(m string, u string, b []byte) *Request {
	return &Request{
		Method:  m,
		BaseURL: u,
		Body:    b,
		Header:  make(http.Header),
		Param:   make(url.Values),
	}
}

// SetContentTypeJSON sets the Content-Type header to "application/json".
func (r *Request) SetContentTypeJSON() {
	if r.Header == nil {
		r.Header = make(http.Header)
	}

	r.Header.Set("Content-Type", "application/json")
}

// SetContentType sets the Content-Type header to a given value.
func (r *Request) SetContentType(value string) {
	if r.Header == nil {
		r.Header = make(http.Header)
	}

	r.Header.Set("Content-Type", value)
}

// SetAuthorization sets the Authorization header to a given value.
func (r *Request) SetAuthorization(value string) {
	if r.Header == nil {
		r.Header = make(http.Header)
	}

	r.Header.Set("Authorization", value)
}

// Response holds data from a response.
type Response struct {
	StatusCode int
	Status     string
	Body       []byte
	Headers    http.Header
}
