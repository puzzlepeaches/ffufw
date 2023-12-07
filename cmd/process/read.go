package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

type Output struct {
	Commandline string `json:"commandline"`
	Time        string `json:"time"`
	Results     []struct {
		URL string `json:"url"`
	} `json:"results"`
	Config struct {
		URL string `json:"url"`
	} `json:"config"`
}

func ParseOutput(outputFile string) ([]string, error) {
	// Open the output file
	file, err := os.Open(outputFile)
	if err != nil {
		return nil, fmt.Errorf("Error opening file: %s", err)
	}
	defer file.Close()

	// Read the file
	byteValue, err := ioutil.ReadAll(file)
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

	// Store the results in a list
	urls := make([]string, len(output.Results))
	for i, result := range output.Results {
		urls[i] = result.URL
	}
	// Return the list of URLs
	return urls, nil
}
