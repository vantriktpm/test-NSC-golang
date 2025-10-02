package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortCode   string `json:"short_code"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
	CreatedAt   string `json:"created_at"`
}

func main() {
	baseURL := "http://localhost:8080"
	apiURL := baseURL + "/api/v1/shorten"

	// Test URLs
	testURLs := []string{
		"https://www.google.com",
		"https://www.github.com",
		"https://www.stackoverflow.com",
		"https://www.reddit.com",
		"https://www.youtube.com",
	}

	// Test duplicate URLs (should be fast)
	duplicateURLs := []string{
		"https://www.google.com",
		"https://www.google.com",
		"https://www.google.com",
	}

	fmt.Println("Testing URL Shortener Performance")
	fmt.Println("=================================")

	// Test 1: Initial URL shortening
	fmt.Println("\n1. Testing initial URL shortening...")
	start := time.Now()

	var wg sync.WaitGroup
	results := make(chan time.Duration, len(testURLs))

	for _, url := range testURLs {
		wg.Add(1)
		go func(testURL string) {
			defer wg.Done()
			reqStart := time.Now()

			reqBody := ShortenRequest{URL: testURL}
			jsonData, _ := json.Marshal(reqBody)

			resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer resp.Body.Close()

			var result ShortenResponse
			json.NewDecoder(resp.Body).Decode(&result)

			duration := time.Since(reqStart)
			results <- duration

			fmt.Printf("✓ %s -> %s (took %v)\n", testURL, result.ShortURL, duration)
		}(url)
	}

	wg.Wait()
	close(results)

	totalTime := time.Since(start)
	fmt.Printf("Total time for %d URLs: %v\n", len(testURLs), totalTime)

	// Test 2: Duplicate URL handling (should be very fast)
	fmt.Println("\n2. Testing duplicate URL handling...")
	start = time.Now()

	for _, url := range duplicateURLs {
		reqStart := time.Now()

		reqBody := ShortenRequest{URL: url}
		jsonData, _ := json.Marshal(reqBody)

		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		defer resp.Body.Close()

		var result ShortenResponse
		json.NewDecoder(resp.Body).Decode(&result)

		duration := time.Since(reqStart)
		fmt.Printf("✓ Duplicate %s -> %s (took %v)\n", url, result.ShortURL, duration)
	}

	duplicateTime := time.Since(start)
	fmt.Printf("Total time for %d duplicate URLs: %v\n", len(duplicateURLs), duplicateTime)

	// Test 3: Redirect performance
	fmt.Println("\n3. Testing redirect performance...")

	// Get a short URL first
	reqBody := ShortenRequest{URL: "https://www.example.com"}
	jsonData, _ := json.Marshal(reqBody)
	resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Printf("Error creating test URL: %v\n", err)
		return
	}
	defer resp.Body.Close()

	var result ShortenResponse
	json.NewDecoder(resp.Body).Decode(&result)

	shortCode := result.ShortCode
	redirectURL := baseURL + "/" + shortCode

	// Test redirect multiple times
	redirectTimes := 10
	start = time.Now()

	for i := 0; i < redirectTimes; i++ {
		reqStart := time.Now()

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err := client.Get(redirectURL)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			continue
		}
		resp.Body.Close()

		duration := time.Since(reqStart)
		fmt.Printf("✓ Redirect %d: %v (status: %d)\n", i+1, duration, resp.StatusCode)
	}

	redirectTime := time.Since(start)
	fmt.Printf("Total time for %d redirects: %v\n", redirectTimes, redirectTime)

	fmt.Println("\n=================================")
	fmt.Println("Performance Test Complete!")
	fmt.Printf("Average time per initial URL: %v\n", totalTime/time.Duration(len(testURLs)))
	fmt.Printf("Average time per duplicate URL: %v\n", duplicateTime/time.Duration(len(duplicateURLs)))
	fmt.Printf("Average time per redirect: %v\n", redirectTime/time.Duration(redirectTimes))
}
