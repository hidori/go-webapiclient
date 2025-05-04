package webapiclient

import (
	"bytes"
	"io"
	"net/http"
	"path"
	"strings"

	"github.com/pkg/errors"
)

type Client interface {
	Do(request *Request, editor HTTPRequestEditorFunc) (*Response, error)
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

type HTTPRequestEditorFunc func(httpRequest *http.Request) error

type DoFunc func(httpRequest *http.Request) (*http.Response, error)

type ClientImpl struct {
	do      DoFunc
	baseURL string
}

func NewClient(do DoFunc, baseURL string) Client {
	return &ClientImpl{
		do:      do,
		baseURL: baseURL,
	}
}

func (c *ClientImpl) Do(request *Request, editor HTTPRequestEditorFunc) (*Response, error) {
	var requestBody io.Reader
	if request.Method != http.MethodGet && request.Body != nil {
		requestBody = bytes.NewReader(request.Body)
	}

	httpRequest, err := http.NewRequest(request.Method, path.Join(c.baseURL, request.Path), requestBody)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	for key, values := range request.Headers {
		for _, value := range values {
			httpRequest.Header.Add(key, value)
		}
	}

	if editor != nil {
		err := editor(httpRequest)
		if err != nil {
			return nil, errors.WithStack(err)
		}
	}

	httpResponse, err := c.do(httpRequest)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	defer httpResponse.Body.Close()

	if len(request.ExpectedStatusCodes) > 0 && !containsInt(request.ExpectedStatusCodes, httpResponse.StatusCode) {
		return nil, errors.Errorf("unexpected status code: %d", httpResponse.StatusCode)
	}

	contentType := httpResponse.Header.Get("Content-Type")
	if len(request.ExpectedContentTypes) > 0 && !containsStringPrefix(request.ExpectedContentTypes, contentType) {
		return nil, errors.Errorf("unexpected content type: %s", contentType)
	}

	responseBody, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	response := &Response{
		StatusCode: httpResponse.StatusCode,
		Headers:    httpResponse.Header.Clone(),
		Body:       responseBody,
	}

	return response, nil
}

// Helper function to check if a list of integers contains a specific integer
func containsInt(list []int, item int) bool {
	for _, v := range list {
		if v == item {
			return true
		}
	}
	return false
}

// Helper function to check if any strings in a list are prefixes of another string
func containsStringPrefix(list []string, s string) bool {
	for _, prefix := range list {
		if strings.HasPrefix(s, prefix) {
			return true
		}
	}
	return false
}
