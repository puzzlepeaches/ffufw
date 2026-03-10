package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
)

// Result represents a single finding from ffuf with its metadata
type Result struct {
	URL         string `json:"url"`
	Status      int    `json:"status"`
	Length      int    `json:"length"`
	Words       int    `json:"words"`
	Lines       int    `json:"lines"`
	ContentType string `json:"content-type"`
	RedirectURL string `json:"redirectlocation"`
}

type Output struct {
	Commandline string   `json:"commandline"`
	Time        string   `json:"time"`
	Results     []Result `json:"results"`
	Config      struct {
		URL string `json:"url"`
	} `json:"config"`
}

func ParseOutput(outputFile string) ([]Result, error) {
	// Open the output file
	file, err := os.Open(outputFile)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %s", err)
	}
	defer file.Close()

	// Read the file
	byteValue, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("Error reading file: %s", err)
	}

	// Initialize a new Output struct
	var output Output

	// Unmarshal the JSON
	err = json.Unmarshal(byteValue, &output)
	if err != nil {
		return nil, fmt.Errorf("Error unmarshalling JSON: %s", err)
	}

	return output.Results, nil
}
