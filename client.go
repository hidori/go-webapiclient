package webapiclient

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/pkg/errors"
)

type Client interface {
	Do(request *Request, edit EditRequestFunc) (*Response, error)
}

type Request struct {
	Method               string
	Path                 string
	Headers              map[string][]string
	Body                 []byte
	ExpectedStatusCodes  []int
	ExpectedContentTypes []string
}

type Response struct {
	StatusCode int
	Headers    map[string][]string
	Body       []byte
}

type EditRequestFunc func(httpRequest *http.Request) error

type DoFunc func(httpRequest *http.Request) (*http.Response, error)

type ClientImpl struct {
	do      DoFunc
	baseURL string
}

func NewClient(do DoFunc, baseURL string) *ClientImpl {
	return &ClientImpl{
		do:      do,
		baseURL: baseURL,
	}
}

func (c *ClientImpl) Do(request *Request, edit EditRequestFunc) (*Response, error) {
	httpRequest, err := c.buildHTTPRequest(request)
	if err != nil {
		return nil, err
	}

	if edit != nil {
		if err := edit(httpRequest); err != nil {
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

	if err := c.validateResponse(httpResponse, request); err != nil {
		return nil, err
	}

	return c.readResponse(httpResponse)
}

func (c *ClientImpl) buildHTTPRequest(request *Request) (*http.Request, error) {
	var requestBody io.Reader
	if request.Method != http.MethodGet && request.Body != nil {
		requestBody = bytes.NewReader(request.Body)
	}

	baseURL, err := url.Parse(c.baseURL)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	requestURL, err := baseURL.Parse(request.Path)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	httpRequest, err := http.NewRequestWithContext(context.Background(), request.Method, requestURL.String(), requestBody)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for key, values := range request.Headers {
		for _, value := range values {
			httpRequest.Header.Add(key, value)
		}
	}

	return httpRequest, nil
}

func (c *ClientImpl) validateResponse(httpResponse *http.Response, request *Request) error {
	if len(request.ExpectedStatusCodes) > 0 && !slices.Contains(request.ExpectedStatusCodes, httpResponse.StatusCode) {
		return errors.Errorf("unexpected status code: %d", httpResponse.StatusCode)
	}

	contentType := httpResponse.Header.Get("Content-Type")
	if len(request.ExpectedContentTypes) > 0 && !slices.ContainsFunc(request.ExpectedContentTypes, func(prefix string) bool {
		return strings.HasPrefix(strings.ToLower(contentType), strings.ToLower(prefix))
	}) {
		return errors.Errorf("unexpected content type: %s", contentType)
	}

	return nil
}

func (c *ClientImpl) readResponse(httpResponse *http.Response) (*Response, error) {
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
