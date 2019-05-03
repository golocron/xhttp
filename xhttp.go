// Package xhttp provides custom http client based on net/http.
package xhttp

import (
	"bytes"
	"crypto/tls"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// Library level defaults.
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

// SendRaw makes a request based on http.Request using DefaultClient.
func SendRaw(req *http.Request) (*http.Response, error) {
	return DefaultClient.SendRaw(req)
}

// Send makes a request based on Request using DefaultClient.
func Send(req *Request) (*Response, error) {
	return DefaultClient.Send(req)
}

// Get makes a Get request using DefaultClient.
func Get(url string) (*Response, error) {
	return DefaultClient.Get(url)
}

// Post makes a Post request using DefaultClient.
func Post(url, contentType string, body []byte) (*Response, error) {
	return DefaultClient.Post(url, contentType, body)
}

// DownloadFile downloads the file located at the url.
// It stores the result in a file with the filename.
func DownloadFile(url, filename string) error {
	return DefaultClient.DownloadFile(url, filename)
}

// Request defines a request.
type Request struct {
	Method  string
	BaseURL string
	Body    []byte
	Header  http.Header
	Param   url.Values
}

// NewRequest returns a Request ready for usage.
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

// Response holds data from response.
type Response struct {
	StatusCode int
	Status     string
	Body       []byte
	Headers    http.Header
}

// ClientConfig holds configuration for Client.
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

// Client represents custom http client wrapper around net/http.Client.
type Client struct {
	HTTPClient http.Client
}

// NewClient returns Client with customized default settings.
//
// More details here https://blog.cloudflare.com/the-complete-guide-to-golang-net-http-timeouts/
func NewClient() *Client {
	return createClient(DefaultClientConfig())
}

// NewClientWithConfig creates new Client with settings specified in cfg.
func NewClientWithConfig(cfg *ClientConfig) *Client {
	return createClient(cfg)
}

// createClient creates new Client using custom tls.Config and http.Transport.
func createClient(cfg *ClientConfig) *Client {
	// Create a custom tls config.
	tlsConfig := &tls.Config{
		InsecureSkipVerify: cfg.SkipTLSVerify,
	}

	// Create a pool with root CA.
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

// buildRequest creates the http.Request from Request.
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

// SendRaw performs a request based on http.Request.
func (c *Client) SendRaw(req *http.Request) (*http.Response, error) {
	return c.HTTPClient.Do(req)
}

// buildResponse builds Response from http.Response.
//
// It takes care of closing body as well.
func (c *Client) buildResponse(res *http.Response) (*Response, error) {
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	response := &Response{
		StatusCode: res.StatusCode,
		Status:     res.Status,
		Body:       body,
		Headers:    res.Header,
	}

	return response, nil
}

// Send makes request.
func (c *Client) Send(request *Request) (*Response, error) {
	// Build the HTTP request object.
	req, err := c.buildRequest(request)
	if err != nil {
		return nil, err
	}

	// Build the HTTP client and make the request.
	res, err := c.SendRaw(req)
	if err != nil {
		return nil, err
	}

	return c.buildResponse(res)
}

// Get makes GET request.
func (c *Client) Get(url string) (*Response, error) {
	req := NewRequest(http.MethodGet, url, nil)

	return c.Send(req)
}

// Post makes Post request.
func (c *Client) Post(url, contentType string, body []byte) (*Response, error) {
	req := NewRequest(http.MethodPost, url, body)
	req.Header.Set("Content-Type", contentType)

	return c.Send(req)
}

// DownloadFile performs simple file downloading and storing in a given path.
func (c *Client) DownloadFile(url, filename string) error {
	resp, err := c.Get(url)
	if err != nil {
		return err
	}

	if resp.StatusCode != 200 {
		return errors.New(resp.Status)
	}

	return ioutil.WriteFile(filename, resp.Body, 0644)
}
