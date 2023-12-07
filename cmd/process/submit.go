package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

func SubmitURL(gowitnessAddress string, result string) error {

	// Craft the json payload
	payload := map[string]string{
		"url":     result,
		"oneshot": "false",
	}

	// Marshal the payload
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	// Craft the request
	req, err := http.NewRequest("POST", gowitnessAddress+"/api/screenshot", bytes.NewBuffer(jsonPayload))
	if err != nil {
		return err
	}

	// Set the content type
	req.Header.Set("Content-Type", "application/json")

	// Create a new HTTP client
	client := &http.Client{}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	logrus.Debugf("Submitted URL to gowitness: %s [RESP: %s]", result, resp.Status)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	defer resp.Body.Close()

	return nil

}
