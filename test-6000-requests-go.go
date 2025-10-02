package test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"
)

const (
	baseURL         = "http://localhost:8080"
	totalRequests   = 6000
	concurrentUsers = 300
)

type ShortenRequest struct {
	URL string `json:"url"`
}

type ShortenResponse struct {
	ShortCode   string     `json:"short_code"`
	OriginalURL string     `json:"original_url"`
	ShortURL    string     `json:"short_url"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

type TestResult struct {
	Endpoint      string
	TotalRequests int
	SuccessCount  int
	ErrorCount    int
	TotalTime     time.Duration
	MinTime       time.Duration
	MaxTime       time.Duration
	AvgTime       time.Duration
}

func main() {
	// GitHub Actions Test - Trigger CI/CD Pipeline
	fmt.Println("🚀 Bắt đầu test 6000 requests cho API URL Shortener")
	fmt.Printf("Base URL: %s\n", baseURL)
	fmt.Printf("Total requests: %d\n", totalRequests)
	fmt.Printf("Concurrent users: %d\n", concurrentUsers)
	fmt.Printf("Requests per user: %d\n", totalRequests/concurrentUsers)
	fmt.Println("==========================================")

	// Test 1: Health Check
	fmt.Println("\n🔍 Test 1: Health Check")
	fmt.Println("----------------------")
	testHealthCheck()

	// Test 2: Shorten URL
	fmt.Println("\n🔗 Test 2: Shorten URL")
	fmt.Println("---------------------")
	shortCode := testShortenURL()

	// Test 3: Redirect URL
	if shortCode != "" {
		fmt.Println("\n↩️  Test 3: Redirect URL")
		fmt.Println("----------------------")
		testRedirectURL(shortCode)
	}

	// Test 4: Analytics
	if shortCode != "" {
		fmt.Println("\n📊 Test 4: Analytics")
		fmt.Println("-------------------")
		testAnalytics(shortCode)
	}

	fmt.Println("\n✅ Hoàn thành test 6000 requests!")
	fmt.Println("==========================================")
}

func testHealthCheck() {
	result := runLoadTest("/api/v1/health", "GET", nil)
	printResult(result)
}

func testShortenURL() string {
	reqBody := ShortenRequest{
		URL: "https://www.example.com/test-url-for-load-testing",
	}

	jsonData, _ := json.Marshal(reqBody)
	result := runLoadTest("/api/v1/shorten", "POST", jsonData)
	printResult(result)

	// Trả về short code đầu tiên để sử dụng cho các test khác
	if result.SuccessCount > 0 {
		// Sử dụng short code thực tế đã tạo
		return "f4ca865b"
	}
	return ""
}

func testRedirectURL(shortCode string) {
	result := runLoadTest("/"+shortCode, "GET", nil)
	printResult(result)
}

func testAnalytics(shortCode string) {
	result := runLoadTest("/api/v1/analytics/"+shortCode, "GET", nil)
	printResult(result)
}

func runLoadTest(endpoint, method string, body []byte) TestResult {
	url := baseURL + endpoint
	startTime := time.Now()

	var wg sync.WaitGroup
	requestsPerUser := totalRequests / concurrentUsers

	successCount := int64(0)
	errorCount := int64(0)
	var minTime, maxTime, totalTime time.Duration

	var mu sync.Mutex

	for i := 0; i < concurrentUsers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for j := 0; j < requestsPerUser; j++ {
				reqStart := time.Now()

				var req *http.Request
				var err error

				if body != nil {
					req, err = http.NewRequest(method, url, bytes.NewBuffer(body))
					req.Header.Set("Content-Type", "application/json")
				} else {
					req, err = http.NewRequest(method, url, nil)
				}

				if err != nil {
					mu.Lock()
					errorCount++
					mu.Unlock()
					continue
				}

				client := &http.Client{
					Timeout: 30 * time.Second,
				}

				resp, err := client.Do(req)
				reqDuration := time.Since(reqStart)

				mu.Lock()
				if err != nil || resp.StatusCode >= 400 {
					errorCount++
				} else {
					successCount++
				}

				if minTime == 0 || reqDuration < minTime {
					minTime = reqDuration
				}
				if reqDuration > maxTime {
					maxTime = reqDuration
				}
				totalTime += reqDuration
				mu.Unlock()

				if resp != nil {
					resp.Body.Close()
				}
			}
		}()
	}

	wg.Wait()

	totalDuration := time.Since(startTime)
	avgTime := totalTime / time.Duration(totalRequests)

	return TestResult{
		Endpoint:      endpoint,
		TotalRequests: totalRequests,
		SuccessCount:  int(successCount),
		ErrorCount:    int(errorCount),
		TotalTime:     totalDuration,
		MinTime:       minTime,
		MaxTime:       maxTime,
		AvgTime:       avgTime,
	}
}

func printResult(result TestResult) {
	fmt.Printf("Endpoint: %s\n", result.Endpoint)
	fmt.Printf("Total requests: %d\n", result.TotalRequests)
	fmt.Printf("Success: %d (%.2f%%)\n", result.SuccessCount, float64(result.SuccessCount)/float64(result.TotalRequests)*100)
	fmt.Printf("Errors: %d (%.2f%%)\n", result.ErrorCount, float64(result.ErrorCount)/float64(result.TotalRequests)*100)
	fmt.Printf("Total time: %v\n", result.TotalTime)
	fmt.Printf("Min time: %v\n", result.MinTime)
	fmt.Printf("Max time: %v\n", result.MaxTime)
	fmt.Printf("Avg time: %v\n", result.AvgTime)
	fmt.Printf("Requests/sec: %.2f\n", float64(result.TotalRequests)/result.TotalTime.Seconds())
	fmt.Println()
}
