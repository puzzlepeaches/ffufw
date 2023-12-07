package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
)

func downloadFile(filepath string, url string) error {
	// Create file
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Download file
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// Write body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

func createDirectory(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		// Create directory
		err := os.Mkdir(path, 0755)
		if err != nil {
			logrus.Fatalf("Could not create directory at %s", path)
		}
	}
}

func getWordlists(wordlists []Wordlist, dir string) {
	// Create directory
	createDirectory(dir)

	// Download wordlists
	for _, wordlist := range wordlists {
		// Check if wordlist exists
		wordlistPath := filepath.Join(dir, wordlist.Name+".txt")
		if _, err := os.Stat(wordlistPath); os.IsNotExist(err) {
			// Download wordlist
			logrus.Infof("Downloading %s wordlist", wordlist.Name)
			err := downloadFile(wordlistPath, wordlist.URL)
			if err != nil {
				logrus.Fatalf("Could not download %s wordlist: %s", wordlist.Name, err)
			}
		}
		if wordlist.Name == "leaky-paths" {
			// open file and remove leading slash
			removeLeadingSlash(wordlistPath)
		}
	}
}

func removeLeadingSlash(wordlistPath string) {
	file, err := os.OpenFile(wordlistPath, os.O_RDWR, 0644)
	if err != nil {
		logrus.Fatalf("Could not open file: %s", err)
	}
	defer file.Close()

	// Read the file line by line and remove leading slash
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "/") {
			line = strings.TrimPrefix(line, "/")
		}
		lines = append(lines, line)
	}

	// Check for errors from scanner
	if err := scanner.Err(); err != nil {
		logrus.Fatalf("Could not read file: %s", err)
	}

	// Write the updated lines back to the file
	file.Seek(0, 0)
	file.Truncate(0)
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()
}

func WordlistPath() {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("Could not find home directory: %s", err)
	}
	// Check if wordlists directory exists
	wordlistsDir := filepath.Join(home, ".ffufw", "wordlists")
	createDirectory(wordlistsDir)
}

func getMiscWordlists() {
	home, _ := os.UserHomeDir()
	miscWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "misc")
	getWordlists(MiscWordlists, miscWordlistsDir)
}

func getIisWordlists() {
	home, _ := os.UserHomeDir()
	iisWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "iis")
	getWordlists(IisWordlists, iisWordlistsDir)
}

func getPhpWordlists() {
	home, _ := os.UserHomeDir()
	phpWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "php")
	getWordlists(PhpWordlists, phpWordlistsDir)
}

func getJavaWordlists() {
	home, _ := os.UserHomeDir()
	javaWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "java")
	getWordlists(JavaWordlists, javaWordlistsDir)
}

func getApiWordlists() {
	home, _ := os.UserHomeDir()
	apiWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "api")
	getWordlists(ApiWordlists, apiWordlistsDir)
}

func getPythonWorlists() {
	home, _ := os.UserHomeDir()
	pythonWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "python")
	getWordlists(PythonWordlists, pythonWordlistsDir)
}

func getRubyWorlists() {
	home, _ := os.UserHomeDir()
	rubyWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "ruby")
	getWordlists(RubyWordlists, rubyWordlistsDir)
}

func getSapWordlists() {
	home, _ := os.UserHomeDir()
	sapWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "sap")
	getWordlists(SapWordlists, sapWordlistsDir)
}

func getNginxWordlists() {
	home, _ := os.UserHomeDir()
	nginxWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "nginx")
	getWordlists(NginxWordlists, nginxWordlistsDir)
}

func getAdobeWordlists() {
	home, _ := os.UserHomeDir()
	adobeWordlistsDir := filepath.Join(home, ".ffufw", "wordlists", "adobe")
	getWordlists(AdobeWordlists, adobeWordlistsDir)
}

func GetWordlistsAll() {
	getMiscWordlists()
	getIisWordlists()
	getPhpWordlists()
	getJavaWordlists()
	getApiWordlists()
	getPythonWorlists()
	getRubyWorlists()
	getSapWordlists()
	getNginxWordlists()
	getAdobeWordlists()
}
