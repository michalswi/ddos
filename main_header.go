package main

import (
	"bytes"
	"fmt"
	"net/http"
	"sync"
	"time"
)

func main() {
	targetUrl := "http://127.0.0.1:8080"

	// number of concurrent connections
	numConnections := 3
	// number of requests per connection
	requestsPerConnection := 12

	bearerToken := "token_value_here"

	// JSON payload
	payload := []byte(`{"name": "dupa", "withAssets": false}`)

	var wg sync.WaitGroup

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	errChan := make(chan error, numConnections*requestsPerConnection)

	for i := 0; i < numConnections; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < requestsPerConnection; j++ {
				req, err := http.NewRequest("POST", targetUrl, bytes.NewBuffer(payload))
				if err != nil {
					errChan <- fmt.Errorf("connection %d, request %d failed to create: %v", id, j, err)
					continue
				}

				token := fmt.Sprintf("Bearer %s", bearerToken)
				req.Header.Set("Authorization", token)
				req.Header.Set("Content-Type", "application/json")

				resp, err := client.Do(req)
				if err != nil {
					errChan <- fmt.Errorf("connection %d, request %d failed: %v", id, j, err)
					continue
				}
				defer resp.Body.Close()

				fmt.Printf("Connection %d, Request %d: Status %s\n", id, j, resp.Status)
			}
		}(i)
	}
	wg.Wait()
	close(errChan)
	for err := range errChan {
		fmt.Println(err)
	}
	fmt.Println("All requests completed")
}
