package cmd

import (
	"bufio"
	"crypto/tls"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	ffuf "github.com/puzzlepeaches/ffufw/cmd/ffuf"
	"github.com/sirupsen/logrus"
)

// sharedHTTPClient is reused across all HTTP requests to enable connection pooling
var sharedHTTPClient = &http.Client{
	Transport: &http.Transport{
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     30 * time.Second,
	},
	Timeout: 10 * time.Second,
}

func convertTech(url string, fingerprints []string) ffuf.TechData {
	// logTech(url, fingerprints)
	technologies := defineStruct(url, fingerprints)
	techData := ffuf.TechData{
		Url:    technologies.url,
		Iis:    technologies.iis,
		Apache: technologies.apache,
		Nginx:  technologies.nginx,
		Php:    technologies.php,
		Java:   technologies.java,
		Python: technologies.python,
		Api:    technologies.api,
		Sap:    technologies.sap,
		Ruby:   technologies.ruby,
		Adobe:  technologies.adobe,
	}

	return techData

}

func removeMicrosoftUrls(urls []string) []string {

	var newUrls []string

	var excludedStrings = []string{
		"autodiscover",
		"lyncdiscover",
		"enterpriseenrollment",
		"enterpriseregistration",
		"_sip",
		"_sipfederationtls",
		"_tcp",
		"_tls",
		"msoid",
		"sip",
	}

	for _, url := range urls {
		excluded := false
		for _, exStr := range excludedStrings {
			if strings.Contains(url, exStr) {
				excluded = true
				break
			}
		}
		if !excluded {
			newUrls = append(newUrls, url)
		}
	}

	return newUrls

}

func readInputFile(inputFile string) ([]string, error) {
	// Open the file
	file, err := os.Open(inputFile)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the file line by line
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	// Check for errors from scanner
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func expandPath(path string) string {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			logrus.Fatalf("Could not get home directory: %v", err)
		}
		path = filepath.Join(homeDir, path[2:])
	}
	return path
}
