package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
)

// fetchKraData fetches KRA data based on the provided URL environment variable
func (a *Api)fetchKraData() (*KraTaxData, error) {
	
	// Retrieve the KRA data URL from environment variable
	url := os.Getenv("MAYBETS_KRA_DATA_URL")
	if url == "" {
		return nil, fmt.Errorf("MAYBETS_KRA_DATA_URL environment variable is not set")
	}

	// Make HTTP GET request to fetch KRA data
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch KRA data: %v", err)
	}
	defer resp.Body.Close()

	// Check HTTP response status
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var kraData KraTaxData
	if err := json.Unmarshal(body, &kraData); err != nil {
		return nil, fmt.Errorf("failed to decode KRA data: %v", err)
	}

	return &kraData, nil
}
