package cmd

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sirupsen/logrus"
)

func SubmitGowitness(gowitnessAddress string, result string) error {

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
	client := &http.Client{
		Timeout: time.Second * 30,
	}

	// Send the request
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // drain body for connection reuse

	logrus.Debugf("Submitted URL to gowitness: %s [RESP: %s]", result, resp.Status)

	if resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil

}

func SubmitReplayProxy(replayProxy string, result string) error {
	// Sending to replay proxy

	proxyURL, err := url.Parse(replayProxy)
	if err != nil {
		return fmt.Errorf("could not parse replay proxy URL %s: %w", replayProxy, err)
	}

	// Create a new HTTP client with the proxy, a timeout, and disable SSL verification
	client := &http.Client{
		Transport: &http.Transport{
			Proxy:           http.ProxyURL(proxyURL),
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
		Timeout: time.Second * 10,
	}

	// Create a new request
	req, err := http.NewRequest("GET", result, nil)
	if err != nil {
		return fmt.Errorf("could not create request for %s: %w", result, err)
	}

	// Set the User-Agent to mimic a Chrome browser
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// Send the request to the proxy
	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("Error sending request to replay proxy: %s", err)
		return err
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body) // drain body for connection reuse

	// Show the response status code
	logrus.Debugf("Submitted URL to replay proxy: %s [RESP: %s]", result, resp.Status)

	return nil
}
