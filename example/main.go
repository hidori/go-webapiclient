package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/hidori/go-webapiclient"
)

func main() {
	// Create a new client with default HTTP client
	client := webapiclient.NewClient(http.DefaultClient.Do, "http://google.com")

	// Prepare the request
	request := &webapiclient.Request{
		Method:               http.MethodGet,
		Path:                 "/",
		ExpectedStatusCodes:  []int{http.StatusOK, http.StatusMovedPermanently, http.StatusFound},
		ExpectedContentTypes: []string{"text/html"},
	}

	// Make the request
	response, err := client.Do(context.Background(), request, nil)
	if err != nil {
		log.Fatalf("Failed to make request: %v", err)
	}
	defer func() {
		_ = response.Body.Close()
	}()

	// Read the response body
	bodyBytes, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("Failed to read response body: %v", err)
	}

	// Output the first 256 bytes of the response body
	bodyLength := len(bodyBytes)
	if bodyLength > 256 {
		bodyLength = 256
	}

	fmt.Printf("Response Status: %d\n", response.StatusCode)
	fmt.Printf("First %d bytes of response body:\n", bodyLength)
	fmt.Printf("%s\n", string(bodyBytes[:bodyLength]))
}
