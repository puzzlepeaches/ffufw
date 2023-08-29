package cmd

import (
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

func runFFUF(urlString string, outputDir string, wordlistPath string, configFile string, stdout bool, quiet bool, ffufPath string, ffufPostprocessingPath string, wg *sync.WaitGroup, sem chan bool) {

	defer wg.Done()
	defer func() { <-sem }()

	parsedURL, urlString := parseURL(urlString)
	if parsedURL == nil {
		return
	}

	hostOutputDir := createOutputDirectory(parsedURL, outputDir)
	if hostOutputDir == "" {
		return
	}

	resultFilePath := runFFUFCommand(urlString, wordlistPath, hostOutputDir, configFile, ffufPath)
	if resultFilePath == "" {
		return
	}

	if !waitForResultsFile(resultFilePath, urlString) {
		return
	}

	if !runPostProcessing(resultFilePath, hostOutputDir, urlString, ffufPostprocessingPath) {
		return
	}

	if stdout {
		printResults(resultFilePath)
	}
}

func parseURL(urlString string) (*url.URL, string) {
	// Parse URL
	parsedURL, err := url.Parse(urlString)
	if err != nil {
		logrus.Errorf("Error parsing URL %s: %v", urlString, err)
		return nil, ""
	}

	// Strip / from end of URL
	urlString = strings.TrimSuffix(urlString, "/")

	return parsedURL, urlString
}

func createOutputDirectory(parsedURL *url.URL, outputDir string) string {
	// Create output directory
	hostDir := strings.ReplaceAll(parsedURL.Host, ":", "_") // Replace colon in case of port
	hostOutputDir := filepath.Join(outputDir, hostDir)

	// Error if output directory cannot be created
	if err := os.MkdirAll(hostOutputDir, 0755); err != nil {
		logrus.Errorf("Error creating directory %s: %v", hostOutputDir, err)
		return ""
	}

	return hostOutputDir
}

func runFFUFCommand(urlString string, wordlistPath string, hostOutputDir string, configFile string, ffufPath string) string {
	// Create fuzz URL
	fuzzURL := urlString + "/FUZZ"

	// Run FFUF
	resultFilePath := filepath.Join(hostOutputDir, "results.json")
	ffufBaseCmd := []string{"-u", fuzzURL, "-w", wordlistPath, "-o", resultFilePath, "-od", hostOutputDir, "-of", "json"}

	if configFile != "" {
		ffufBaseCmd = append(ffufBaseCmd, "-c", configFile)
	}

	// Check if ffufPath is not default
	var cmd *exec.Cmd

	if ffufPath != "ffuf" {
		cmd = exec.Command(ffufPath, ffufBaseCmd...)
	} else {
		cmd = exec.Command("ffuf", ffufBaseCmd...)
	}

	if err := cmd.Run(); err != nil {
		logrus.Errorf("Error running FFUF on %s: %v", urlString, err)
		return ""
	}

	logrus.Infof("FFUF completed for URL %s", urlString)

	return resultFilePath
}

func waitForResultsFile(resultFilePath string, urlString string) bool {
	for i := 0; i < 10; i++ { // Retry up to 10 times
		if _, err := os.Stat(resultFilePath); err == nil {
			return true
		}
		time.Sleep(time.Second) // Sleep for 1 second between checks

		logrus.Infof("Waiting for FFUF results file to be created for URL %s", urlString)
	}

	return false
}

func runPostProcessing(resultFilePath string, hostOutputDir string, urlString string, ffufPostprocessingPath string) bool {

	// Running ffufPostprocessing tool
	var postProcCmd *exec.Cmd
	if ffufPostprocessingPath != "ffufPostprocessing" {
		postProcCmd = exec.Command(ffufPostprocessingPath, "-delete-all-bodies", "-result-file", resultFilePath, "-bodies-folder", hostOutputDir, "-overwrite-result-file")
	} else {
		postProcCmd = exec.Command("ffufPostprocessing", "-delete-all-bodies", "-result-file", resultFilePath, "-bodies-folder", hostOutputDir, "-overwrite-result-file")
	}

	// postProcCmd := exec.Command("ffufPostprocessing", "-delete-all-bodies", "-result-file", resultFilePath, "-bodies-folder", hostOutputDir, "-overwrite-result-file")

	if err := postProcCmd.Run(); err != nil {
		logrus.Errorf("Error running ffufPostprocessing on %s: %v", hostOutputDir, err)
		return false
	}

	logrus.Infof("Postprocessing completed for URL %s", urlString)

	return true
}

func printResults(resultFilePath string) {
	// Collect output file contents
	outputFile, err := os.Open(resultFilePath)
	if err != nil {
		logrus.Errorf("Error opening results file %s: %v", resultFilePath, err)
		return
	}
	defer outputFile.Close()

	// Print output file contents with newline
	if _, err := io.Copy(os.Stdout, outputFile); err != nil {
		logrus.Errorf("Error printing results file %s: %v", resultFilePath, err)
		return
	}
	os.Stdout.WriteString("\n")
}
