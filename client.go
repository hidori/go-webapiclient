// Package webapiclient provides a simple HTTP client for making API requests.
package webapiclient

import (
	"context"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

// Compile-time check to ensure client implements Client interface.
var _ Client = (*client)(nil)

// Client is an interface for making API requests.
type Client interface {
	// Do executes an HTTP request with optional request editing and returns the response.
	Do(ctx context.Context, request *Request, edit EditRequestFunc) (*Response, error)
}

// Request represents an HTTP request to be made by the client.
type Request struct {
	Method               string
	Path                 string
	Headers              map[string][]string
	Body                 io.Reader
	ExpectedStatusCodes  []int
	ExpectedContentTypes []string
}

// Response represents an HTTP response returned by the client.
type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

// EditRequestFunc is a function type for editing HTTP requests before they are sent.
type EditRequestFunc func(httpRequest *http.Request) error

// DoFunc is a function type for executing HTTP requests.
type DoFunc func(httpRequest *http.Request) (*http.Response, error)

// client is the default implementation of the Client interface.
type client struct {
	do      DoFunc
	baseURL string
}

// NewClient creates a new client instance with the specified DoFunc and base URL.
func NewClient(do DoFunc, baseURL string) Client {
	return &client{
		do:      do,
		baseURL: baseURL,
	}
}

// Do executes an HTTP request with optional request editing and returns the response.
func (c *client) Do(ctx context.Context, request *Request, edit EditRequestFunc) (*Response, error) {
	httpRequest, err := c.buildHTTPRequest(ctx, request)
	if err != nil {
		return nil, err
	}

	if edit != nil {
		err := edit(httpRequest)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	httpResponse, err := c.do(httpRequest)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	defer func() {
		_ = httpResponse.Body.Close()
	}()

	err = c.validateResponse(httpResponse, request)
	if err != nil {
		return nil, err
	}

	return c.readResponse(httpResponse)
}

func (c *client) buildHTTPRequest(ctx context.Context, request *Request) (*http.Request, error) {
	var requestBody io.Reader
	if request.Method != http.MethodGet && request.Body != nil {
		requestBody = request.Body
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	requestURL, err := baseURL.Parse(request.Path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	httpRequest, err := http.NewRequestWithContext(ctx, request.Method, requestURL.String(), requestBody)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for key, values := range request.Headers {
		normalizedKey := http.CanonicalHeaderKey(key)
		for _, value := range values {
			httpRequest.Header.Add(normalizedKey, value)
		}
	}

	return httpRequest, nil
}

func (c *client) validateResponse(httpResponse *http.Response, request *Request) error {
	if len(request.ExpectedStatusCodes) > 0 && !slices.Contains(request.ExpectedStatusCodes, httpResponse.StatusCode) {
		return errors.Errorf("unexpected status code: %d", httpResponse.StatusCode)
	}

	contentType := httpResponse.Header.Get("Content-Type")
	if len(request.ExpectedContentTypes) > 0 && !slices.ContainsFunc(
		request.ExpectedContentTypes,
		func(prefix string) bool {
			return strings.HasPrefix(strings.ToLower(contentType), strings.ToLower(prefix))
		},
	) {
		return errors.Errorf("unexpected content type: %s", contentType)
	}

	return nil
}

func (c *client) readResponse(httpResponse *http.Response) (*Response, error) {
	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &Response{
		StatusCode: httpResponse.StatusCode,
		Headers:    httpResponse.Header.Clone(),
		Body:       responseBody,
	}, nil
}
