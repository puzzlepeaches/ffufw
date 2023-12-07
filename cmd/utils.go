package cmd

import (
	"bufio"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	valid "github.com/asaskevich/govalidator"
	ffuf "github.com/puzzlepeaches/ffufw/cmd/ffuf"
	"github.com/sirupsen/logrus"
)

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

func removMicrosoftUrls(urls []string) []string {

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

func checkURL(url string) (bool, error) {
	// Check if URL is valid
	if !valid.IsURL(url) {
		return false, errors.New("Invalid URL")
	}

	// Create a new HTTP client with a timeout and disable SSL verification
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 10,
	}

	// Create a new request and add Chrome user agent
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// Check if URL is reachable
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	return true, nil
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
			logrus.Fatalf("Could not get home directory")
		}
		path = filepath.Join(homeDir, path[2:])
	}
	return path
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}
