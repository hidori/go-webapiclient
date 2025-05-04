# Go Web API Client

A simple and flexible HTTP client library for making API requests in Go applications.

## Features

- Simple and intuitive API design
- Context support for request cancellation and timeouts
- Flexible request body handling with `io.Reader` interface
- Built-in response validation (status codes and content types)
- Header normalization using `http.CanonicalHeaderKey`
- Comprehensive error handling with stack traces
- Testable design with dependency injection

## Installation

```bash
go get github.com/hidori/go-webapiclient
```

## Requirements

- Go 1.24 or later

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"
    "net/http"

    "github.com/hidori/go-webapiclient"
)

func main() {
    // Create a new client with default HTTP client
    client := webapiclient.NewClient(http.DefaultClient.Do, "https://api.example.com")

    // Prepare the request
    request := &webapiclient.Request{
        Method:               http.MethodGet,
        Path:                 "/users/1",
        ExpectedStatusCodes:  []int{http.StatusOK},
        ExpectedContentTypes: []string{"application/json"},
    }

    // Make the request
    response, err := client.Do(context.Background(), request, nil)
    if err != nil {
        log.Fatalf("Failed to make request: %v", err)
    }

    fmt.Printf("Status: %d\n", response.StatusCode)
    fmt.Printf("Body: %s\n", string(response.Body))
}
```

## Usage

### Creating a Client

```go
import "net/http"

// Using default HTTP client
client := webapiclient.NewClient(http.DefaultClient.Do, "https://api.example.com")

// Using custom HTTP client
customClient := &http.Client{
    Timeout: 30 * time.Second,
}
client := webapiclient.NewClient(customClient.Do, "https://api.example.com")
```

### Making Requests

#### GET Request

```go
request := &webapiclient.Request{
    Method:               http.MethodGet,
    Path:                 "/users",
    ExpectedStatusCodes:  []int{http.StatusOK},
    ExpectedContentTypes: []string{"application/json"},
}

response, err := client.Do(context.Background(), request, nil)
```

#### POST Request with JSON Body

```go
import "strings"

jsonBody := `{"name": "John Doe", "email": "john@example.com"}`

request := &webapiclient.Request{
    Method:               http.MethodPost,
    Path:                 "/users",
    Headers: map[string][]string{
        "Content-Type": {"application/json"},
    },
    Body:                 strings.NewReader(jsonBody),
    ExpectedStatusCodes:  []int{http.StatusCreated},
    ExpectedContentTypes: []string{"application/json"},
}

response, err := client.Do(context.Background(), request, nil)
```

#### Request with Custom Headers

```go
request := &webapiclient.Request{
    Method: http.MethodGet,
    Path:   "/protected",
    Headers: map[string][]string{
        "Authorization": {"Bearer your-token-here"},
        "User-Agent":    {"MyApp/1.0"},
    },
    ExpectedStatusCodes:  []int{http.StatusOK},
    ExpectedContentTypes: []string{"application/json"},
}
```

#### Request Editing

You can modify the HTTP request before it's sent using the `EditRequestFunc`:

```go
editFunc := func(req *http.Request) error {
    // Add query parameters
    q := req.URL.Query()
    q.Add("page", "1")
    q.Add("limit", "10")
    req.URL.RawQuery = q.Encode()
    return nil
}

response, err := client.Do(context.Background(), request, editFunc)
```

### Error Handling

The library provides detailed error information with stack traces:

```go
response, err := client.Do(context.Background(), request, nil)
if err != nil {
    // Error includes stack trace information
    fmt.Printf("Error: %+v\n", err)
    return
}
```

### Response Structure

```go
type Response struct {
    StatusCode int                 // HTTP status code
    Headers    map[string][]string // Response headers
    Body       []byte              // Response body
}
```

## API Reference

### Types

#### `Client` Interface

```go
type Client interface {
    Do(ctx context.Context, request *Request, edit EditRequestFunc) (*Response, error)
}
```

#### `Request` Structure

```go
type Request struct {
    Method               string              // HTTP method (GET, POST, etc.)
    Path                 string              // Request path
    Headers              map[string][]string // Request headers
    Body                 io.Reader           // Request body
    ExpectedStatusCodes  []int               // Expected HTTP status codes
    ExpectedContentTypes []string            // Expected content types
}
```

#### `EditRequestFunc` Type

```go
type EditRequestFunc func(httpRequest *http.Request) error
```

### Functions

#### `NewClient`

```go
func NewClient(do DoFunc, baseURL string) *ClientImpl
```

Creates a new client instance with the specified HTTP function and base URL.

## Development

### Prerequisites

- Go 1.24 or later
- Docker (for linting)

### Running Tests

```bash
make test
```

### Running Linter

```bash
make lint
```

### Running Example

```bash
make run
```

### Building and Testing

```bash
# Run tests and linting
make test lint

# Format code
make format
```

## Examples

See the [`example/main.go`](example/main.go) file for a complete working example.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for your changes
5. Run `make test lint` to ensure everything passes
6. Submit a pull request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
