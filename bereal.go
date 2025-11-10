package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"BeNotified.local/linebc"
)

// RegionMoment represents a BeReal moment for a specific region
type RegionMoment struct {
	ID  string `json:"id"`
	TS  int64  `json:"ts"`  // Unix timestamp as number
	UTC string `json:"utc"` // UTC time string
}

// BeRealResponse represents the response structure from BeReal API
type BeRealResponse struct {
	Regions map[string]RegionMoment `json:"regions"`
	Now     struct {
		TS  int64  `json:"ts"`  // Unix timestamp
		UTC string `json:"utc"` // UTC time string
	} `json:"now"`
}

// fetchBeRealLatest fetches the latest BeReal moments
func fetchBeRealLatest(apiKey string) (*BeRealResponse, error) {
	url := fmt.Sprintf("https://bereal.devin.rest/v1/moments/latest?api_key=%s", apiKey)

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Send GET request
	resp, err := client.Get(url)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check HTTP status code
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("HTTP error %d: %s", resp.StatusCode, string(body))
	}

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Parse JSON
	var response BeRealResponse
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %w", err)
	}

	return &response, nil
}

func fetchAsiaEastLatestID(apiKey string) (string, error) {
	response, err := fetchBeRealLatest(apiKey)
	if err != nil {
		return "", err
	}

	regionMoment, exists := response.Regions["asia-east"]
	if !exists {
		return "", fmt.Errorf("region 'asia-east' not found in response")
	}

	if regionMoment.ID == "" {
		return "", fmt.Errorf("no moment ID found for region 'asia-east'")
	}

	return regionMoment.ID, nil
}

func berealMain() {
	if err := linebc.InitFromEnv(); err != nil {
		log.Fatal(err)
		return
	}
	log.Println("LINE client initialized.")

	apiKey := os.Getenv("BEREAL_API_KEY")
	if apiKey == "" {
		log.Fatal("BEREAL_API_KEY environment variable not set")
		return
	}

	sentID, err := fetchAsiaEastLatestID(apiKey)
	if err != nil {
		log.Fatalf("Failed to fetch Asia East latest ID: %v", err)
		return
	}
	log.Printf("Asia East Latest Moment ID: %s", sentID)

	for {
		time.Sleep(5 * time.Second)
		currentID, err := fetchAsiaEastLatestID(apiKey)
		if err != nil {
			log.Printf("Error fetching latest ID: %v", err)
			continue
		}

		if currentID != sentID {
			log.Printf("New moment detected! Old ID: %s, New ID: %s", sentID, currentID)
			sentID = currentID

			currentDate := time.Now().Format("2006-01-02")
			message := fmt.Sprintf("Time to BeReal! (%s)", currentDate)
			if err := linebc.BroadcastText(message); err != nil {
				log.Fatal(err)
			}
		}
	}
}
