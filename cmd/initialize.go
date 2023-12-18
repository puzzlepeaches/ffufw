package cmd

import (
	"bufio"
	"crypto/tls"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"time"

	valid "github.com/asaskevich/govalidator"
	log "github.com/puzzlepeaches/ffufw/log"
	"github.com/sirupsen/logrus"
)

func setLogging(quiet bool, verbose bool) {
	var level logrus.Level
	if quiet {
		level = logrus.ErrorLevel
	} else if verbose {
		level = logrus.DebugLevel
		logrus.Debug("Verbose logging enabled!")
	} else {
		level = logrus.InfoLevel
	}

	logrus.SetLevel(level)     // set level for global logger in logrus
	log.SetDefaultLevel(level) // set level for defaultLogger in log package
}

func checkBinary(path string, name string) string {
	if path == name {
		path, err := exec.LookPath(name)
		if err != nil {
			logrus.Fatalf("%s is not available in PATH", name)
		}
		return path
	} else {
		path = expandPath(path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			logrus.Fatalf("Could not find %s binary at %s", name, path)
		}
		return path
	}

}

func checkGowitness(address string) {
	if address != "" {
		if !valid.IsRequestURL(address) {
			logrus.Fatalf("Invalid URL for gowitness at %s", address)
		}
	}

	// Make sure endpoint is reachable
	if address != "" {

		req, err := http.NewRequest("GET", address, nil)
		if err != nil {
			logrus.Fatalf("Could not create request to %s", address)
		}

		// Create a new HTTP client with a timeout and disable SSL verification
		tr := &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client := &http.Client{Transport: tr}

		// Send the request
		resp, err := client.Do(req)
		if err != nil {
			logrus.Fatalf("Could not reach gowitness at %s", address)
		}
		defer resp.Body.Close()

		// Check the status code
		if resp.StatusCode != http.StatusOK {
			logrus.Fatalf("Unexpected status code from gowitness at %s: %d", address, resp.StatusCode)
		}
	}
}

func checkOutput(outputDir string) {
	outputDir = expandPath(outputDir)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		logrus.Debugf("Creating output directory at %s", outputDir)

		// Create directory
		err := os.Mkdir(outputDir, 0755)
		if err != nil {
			logrus.Fatalf("Could not create output directory at %s", err)
			logrus.Fatalf("Could not create output directory at %s", outputDir)
		}
	}
}

func createWordlistDir() {
	path := expandPath("~/.ffufw/wordlists")

	// Create parent directory if it doesn't exist
	parentDir := filepath.Dir(path)
	if _, err := os.Stat(parentDir); os.IsNotExist(err) {
		err := os.Mkdir(parentDir, 0755)
		if err != nil {
			logrus.Fatalf("Could not create parent directory at %s", parentDir)
		}
	}

	// Create wordlist directory if it doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		logrus.Debugf("Creating wordlist directory at %s", path)

		// Create directory
		err := os.Mkdir(path, 0755)
		if err != nil {
			logrus.Fatalf("Could not create wordlist directory at %s", path)
		}
	}
}

func checkInput(inputFile string) {
	// Expand the path and check if input file exists
	inputFile = expandPath(inputFile)
	_, err := os.Stat(inputFile)
	if os.IsNotExist(err) {
		logrus.Fatalf("Could not find input file at %s", inputFile)
	} else if err != nil {
		logrus.Fatalf("Could not open input file at %s", inputFile)
	}

	// Open the input file
	file, err := os.Open(inputFile)
	if err != nil {
		logrus.Fatalf("Could not open input file at %s", inputFile)
	}
	defer file.Close()

	// Create a new scanner and check if URLs are valid line by line with govalidator
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if !valid.IsRequestURL(scanner.Text()) {
			logrus.Fatalf("Invalid URL in input file: %s", scanner.Text())
		}
	}

	// Check for errors in the scanner
	if err := scanner.Err(); err != nil {
		logrus.Fatalf("Could not read input file at %s", inputFile)
	}
}

func checkFfufConfig(configFile string) {
	configFile = expandPath(configFile)
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logrus.Errorf("Could not find config file at %s", configFile)
		logrus.Debug("Continuing without config file")
	}
}

func checkReplayProxy(replayProxy string) {
	if replayProxy != "" {
		if !valid.IsRequestURL(replayProxy) {
			logrus.Fatalf("Invalid URL for replay proxy at %s", replayProxy)
		}
	}

	// Check if an HTTP proxy is accessible at the replayProxy address
	if replayProxy != "" {

		proxyURL, err := url.Parse(replayProxy)
		if err != nil {
			logrus.Fatalf("Could not parse replay proxy URL: %s", replayProxy)
		}

		// Create a new HTTP client with the proxy, a timeout, and disable SSL verification
		client := &http.Client{
			Transport: &http.Transport{
				Proxy:           http.ProxyURL(proxyURL),
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
			},
			Timeout: time.Second * 10,
		}

		// Send a GET request to the proxy
		resp, err := client.Get("https://httpbin.org/status/200")
		if err != nil {
			logrus.Debug(err)
			logrus.Fatalf("Could not reach proxy at %s", replayProxy)
		}
		defer resp.Body.Close()

		// Check the status code
		if resp.StatusCode != http.StatusOK {
			logrus.Fatalf("Unexpected status code from proxy at %s: %d", replayProxy, resp.StatusCode)
		}
	}
}
