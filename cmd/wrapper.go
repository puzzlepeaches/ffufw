package cmd

import (
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// func runFFUF(urlString string, outputDir string, wordlistPath string, socksProxy string, wg *sync.WaitGroup, sem chan bool) {
func runFFUF(urlString string, outputDir string, wordlistPath string, configFile string, wg *sync.WaitGroup, sem chan bool) {

	defer wg.Done()
	defer func() { <-sem }()

	parsedURL, err := url.Parse(urlString)
	if err != nil {
		logrus.Errorf("Error parsing URL %s: %v", urlString, err)
		return
	}

	// Strip / from end of URL
	if strings.TrimSuffix(urlString, "/") != urlString {
		urlString = strings.TrimSuffix(urlString, "/")
	}

	fuzzURL := urlString + "/FUZZ"

	hostDir := strings.ReplaceAll(parsedURL.Host, ":", "_") // Replace colon in case of port
	hostOutputDir := filepath.Join(outputDir, hostDir)

	err = os.MkdirAll(hostOutputDir, 0755)
	if err != nil {
		logrus.Errorf("Error creating directory %s: %v", hostOutputDir, err)
		return
	}

	resultFilePath := filepath.Join(hostOutputDir, "results.json")
	ffufBaseCmd := []string{"-u", fuzzURL, "-w", wordlistPath, "-o", resultFilePath, "-od", hostOutputDir, "-of", "json"}

	if configFile != "" {
		ffufBaseCmd = append(ffufBaseCmd, "-c", configFile)
	}

	cmd := exec.Command("ffuf", ffufBaseCmd...)

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Error running FFUF on %s: %v", urlString, err)
		return
	}

	logrus.Infof("FFUF completed for URL %s", urlString)

	for i := 0; i < 10; i++ { // Retry up to 10 times
		if _, err := os.Stat(resultFilePath); !os.IsNotExist(err) {
			break
		}
		time.Sleep(time.Second) // Sleep for 1 second between checks

		logrus.Infof("Waiting for FFUF results file to be created for URL %s", urlString)
	}

	// Running ffufPostprocessing tool
	postProcCmd := exec.Command("ffufPostprocessing", "-result-file", resultFilePath, "-bodies-folder", hostOutputDir, "-delete-bodies", "-overwrite-result-file")

	if err := postProcCmd.Run(); err != nil {
		logrus.Errorf("Error running ffufPostprocessing on %s: %v", hostOutputDir, err)
		return
	}

	logrus.Infof("Postprocessing completed for URL %s", urlString)
}
