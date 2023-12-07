package cmd

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	wappalyzer "github.com/projectdiscovery/wappalyzergo"
	"github.com/sirupsen/logrus"
)

type tech struct {
	url    string
	iis    bool
	apache bool
	nginx  bool
	php    bool
	java   bool
	python bool
	api    bool
	sap    bool
	ruby   bool
	adobe  bool
}

func defineStruct(url string, fingerprints []string) tech {
	logrus.Debugf("Detecting technology for: %s", url)

	// Initialize a new tech struct
	tech := tech{
		url: url,
	}

	// Define a map for technology keywords and their corresponding struct fields
	techMap := map[string]*bool{
		"IIS":       &tech.iis,
		"ASP":       &tech.iis,
		"Microsoft": &tech.iis,
		"Apache":    &tech.apache,
		"Nginx":     &tech.nginx,
		"PHP":       &tech.php,
		"Java":      &tech.java,
		"Spring":    &tech.java,
		"Ruby":      &tech.ruby,
		"Rails":     &tech.ruby,
		"SAP":       &tech.sap,
		"Python":    &tech.python,
		"Django":    &tech.python,
		"Flask":     &tech.python,
		"gunicorn":  &tech.python,
		"Adobe":     &tech.adobe,
		"AEM":       &tech.adobe,
		"API":       &tech.api,
		"REST":      &tech.api,
		"JSON":      &tech.api,
	}

	for _, fingerprint := range fingerprints {
		// Attempt to map the technologies to the struct
		for keyword, techPtr := range techMap {
			if strings.Contains(fingerprint, keyword) {
				*techPtr = true
				if keyword == "Spring" {
					tech.api = true
				}
				// search for api in the url
				if strings.Contains(url, "api") {
					tech.api = true
				}
			}
		}
	}

	return tech
}

func detectTech(url string) ([]string, error) {

	// Check if URL is valid and reachable
	isValid, err := checkURL(url)
	if err != nil || !isValid {
		logrus.Debugf("Error checking URL: %s", err)
		return nil, err
	}

	// Get the response from the URL
	resp, err := getResponse(url)
	if err != nil {
		logrus.Debugf("Error issuing request: %s", err)
		return nil, err
	}
	defer resp.Body.Close()

	// Read the response body
	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	// Analyze the target URL
	fingerprints, err := analyzeURL(resp, data)
	if err != nil {
		return nil, err
	}

	// Convert the map to a slice
	fingerprintsString := convertMapToSlice(fingerprints)

	definedStruct := defineStruct(url, fingerprintsString)

	logrus.Debugf("Detected technologies raw: %+v", definedStruct)

	return fingerprintsString, nil

}

func getResponse(url string) (*http.Response, error) {
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
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3")

	// Get the response from the URL
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func analyzeURL(resp *http.Response, data []byte) (map[string]struct{}, error) {
	// Create a new wappalyzer instance
	wappalyzerClient, err := wappalyzer.New()
	if err != nil {
		logrus.Errorf("Could not create new wappalyzer instance: %s", err)
		return nil, err
	}

	// Analyze the target URL
	fingerprints := wappalyzerClient.Fingerprint(resp.Header, data)

	return fingerprints, nil
}

func convertMapToSlice(fingerprints map[string]struct{}) []string {
	// Convert the map to a slice
	keys := make([]string, 0, len(fingerprints))
	for k := range fingerprints {
		keys = append(keys, k)
	}

	return keys
}
