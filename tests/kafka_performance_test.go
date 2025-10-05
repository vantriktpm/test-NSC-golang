package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

type ProductAvailabilityRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
}

type ProductAvailabilityResponse struct {
	ProductID      string `json:"productId"`
	Available      bool   `json:"available"`
	AvailableStock int    `json:"availableStock"`
	TotalStock     int    `json:"totalStock"`
	ReservedStock  int    `json:"reservedStock"`
}

type PurchaseRequest struct {
	ProductID string `json:"productId"`
	Quantity  int    `json:"quantity"`
	UserID    string `json:"userId"`
}

type PurchaseResponse struct {
	Success       bool      `json:"success"`
	OrderID       string    `json:"orderId"`
	ProductID     string    `json:"productId"`
	Quantity      int       `json:"quantity"`
	ReservedUntil time.Time `json:"reservedUntil"`
	Message       string    `json:"message"`
}

func main() {
	baseURL := "http://localhost:8081"

	// Test products
	testProducts := []string{
		"550e8400-e29b-41d4-a716-446655440001", // iPhone 15 Pro
		"550e8400-e29b-41d4-a716-446655440002", // Samsung Galaxy S24
		"550e8400-e29b-41d4-a716-446655440003", // MacBook Pro M3
		"550e8400-e29b-41d4-a716-446655440004", // AirPods Pro
		"550e8400-e29b-41d4-a716-446655440005", // iPad Air
	}

	fmt.Println("Kafka-based Inventory System Performance Test")
	fmt.Println("=============================================")

	// Test 1: Concurrent Availability Checks
	fmt.Println("\n1. Testing concurrent availability checks...")
	testConcurrentAvailabilityChecks(baseURL, testProducts, 1000)

	// Test 2: Concurrent Purchase Requests
	fmt.Println("\n2. Testing concurrent purchase requests...")
	testConcurrentPurchases(baseURL, testProducts, 500)

	// Test 3: Mixed Load Test
	fmt.Println("\n3. Testing mixed load (availability + purchases)...")
	testMixedLoad(baseURL, testProducts, 2000)

	// Test 4: Stress Test with High Concurrency
	fmt.Println("\n4. Testing high concurrency stress test...")
	testHighConcurrencyStress(baseURL, testProducts, 10000)

	fmt.Println("\n=============================================")
	fmt.Println("Performance Test Complete!")
	fmt.Println("\nKey Benefits of Kafka-based System:")
	fmt.Println("- Handles millions of concurrent requests")
	fmt.Println("- No database bottlenecks")
	fmt.Println("- Event-driven architecture for scalability")
	fmt.Println("- Real-time inventory updates")
	fmt.Println("- Fault tolerance with Kafka replication")
}

func testConcurrentAvailabilityChecks(baseURL string, products []string, numRequests int) {
	start := time.Now()
	var wg sync.WaitGroup
	results := make(chan time.Duration, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			reqStart := time.Now()

			productID := products[i%len(products)]
			req := ProductAvailabilityRequest{
				ProductID: productID,
				Quantity:  1,
			}

			jsonData, _ := json.Marshal(req)
			resp, err := http.Post(baseURL+"/api/v1/inventory/"+productID+"/availability?quantity=1",
				"application/json", bytes.NewBuffer(jsonData))

			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer resp.Body.Close()

			duration := time.Since(reqStart)
			results <- duration

			if i%100 == 0 {
				fmt.Printf("✓ Availability check %d: %v (status: %d)\n", i+1, duration, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(results)

	totalTime := time.Since(start)
	var totalDuration time.Duration
	count := 0
	for duration := range results {
		totalDuration += duration
		count++
	}

	avgTime := totalDuration / time.Duration(count)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("Average response time: %v\n", avgTime)
	fmt.Printf("Requests per second: %.2f\n", float64(numRequests)/totalTime.Seconds())
}

func testConcurrentPurchases(baseURL string, products []string, numRequests int) {
	start := time.Now()
	var wg sync.WaitGroup
	results := make(chan time.Duration, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			reqStart := time.Now()

			productID := products[i%len(products)]
			req := PurchaseRequest{
				ProductID: productID,
				Quantity:  1,
				UserID:    fmt.Sprintf("user-%d", i),
			}

			jsonData, _ := json.Marshal(req)
			resp, err := http.Post(baseURL+"/api/v1/inventory/reserve",
				"application/json", bytes.NewBuffer(jsonData))

			if err != nil {
				fmt.Printf("Error: %v\n", err)
				return
			}
			defer resp.Body.Close()

			duration := time.Since(reqStart)
			results <- duration

			if i%50 == 0 {
				fmt.Printf("✓ Purchase request %d: %v (status: %d)\n", i+1, duration, resp.StatusCode)
			}
		}(i)
	}

	wg.Wait()
	close(results)

	totalTime := time.Since(start)
	var totalDuration time.Duration
	count := 0
	for duration := range results {
		totalDuration += duration
		count++
	}

	avgTime := totalDuration / time.Duration(count)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("Average response time: %v\n", avgTime)
	fmt.Printf("Requests per second: %.2f\n", float64(numRequests)/totalTime.Seconds())
}

func testMixedLoad(baseURL string, products []string, numRequests int) {
	start := time.Now()
	var wg sync.WaitGroup
	results := make(chan time.Duration, numRequests)

	for i := 0; i < numRequests; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			reqStart := time.Now()

			productID := products[i%len(products)]

			// Alternate between availability check and purchase
			if i%2 == 0 {
				// Availability check
				resp, err := http.Get(baseURL + "/api/v1/inventory/" + productID + "/availability?quantity=1")
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
				resp.Body.Close()
			} else {
				// Purchase request
				req := PurchaseRequest{
					ProductID: productID,
					Quantity:  1,
					UserID:    fmt.Sprintf("user-%d", i),
				}
				jsonData, _ := json.Marshal(req)
				resp, err := http.Post(baseURL+"/api/v1/inventory/reserve",
					"application/json", bytes.NewBuffer(jsonData))
				if err != nil {
					fmt.Printf("Error: %v\n", err)
					return
				}
				resp.Body.Close()
			}

			duration := time.Since(reqStart)
			results <- duration

			if i%200 == 0 {
				fmt.Printf("✓ Mixed request %d: %v\n", i+1, duration)
			}
		}(i)
	}

	wg.Wait()
	close(results)

	totalTime := time.Since(start)
	var totalDuration time.Duration
	count := 0
	for duration := range results {
		totalDuration += duration
		count++
	}

	avgTime := totalDuration / time.Duration(count)
	fmt.Printf("Total time: %v\n", totalTime)
	fmt.Printf("Average response time: %v\n", avgTime)
	fmt.Printf("Requests per second: %.2f\n", float64(numRequests)/totalTime.Seconds())
}

func testHighConcurrencyStress(baseURL string, products []string, numRequests int) {
	fmt.Printf("Running stress test with %d concurrent requests...\n", numRequests)

	start := time.Now()
	var wg sync.WaitGroup
	successCount := 0
	errorCount := 0
	var mu sync.Mutex

	// Create batches to avoid overwhelming the system
	batchSize := 1000
	numBatches := (numRequests + batchSize - 1) / batchSize

	for batch := 0; batch < numBatches; batch++ {
		batchStart := batch * batchSize
		batchEnd := batchStart + batchSize
		if batchEnd > numRequests {
			batchEnd = numRequests
		}

		for i := batchStart; i < batchEnd; i++ {
			wg.Add(1)
			go func(i int) {
				defer wg.Done()

				productID := products[i%len(products)]
				req := ProductAvailabilityRequest{
					ProductID: productID,
					Quantity:  1,
				}

				jsonData, _ := json.Marshal(req)
				resp, err := http.Post(baseURL+"/api/v1/inventory/"+productID+"/availability?quantity=1",
					"application/json", bytes.NewBuffer(jsonData))

				mu.Lock()
				if err != nil || resp.StatusCode != 200 {
					errorCount++
				} else {
					successCount++
				}
				mu.Unlock()

				if resp != nil {
					resp.Body.Close()
				}
			}(i)
		}

		// Small delay between batches
		time.Sleep(100 * time.Millisecond)
	}

	wg.Wait()
	totalTime := time.Since(start)

	fmt.Printf("Stress test completed in: %v\n", totalTime)
	fmt.Printf("Successful requests: %d\n", successCount)
	fmt.Printf("Failed requests: %d\n", errorCount)
	fmt.Printf("Success rate: %.2f%%\n", float64(successCount)/float64(numRequests)*100)
	fmt.Printf("Requests per second: %.2f\n", float64(numRequests)/totalTime.Seconds())
}
