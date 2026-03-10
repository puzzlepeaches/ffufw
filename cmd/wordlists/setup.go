package cmd

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

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

func createDirectory(path string) error {
	if err := os.MkdirAll(path, 0755); err != nil {
		return fmt.Errorf("could not create directory at %s: %w", path, err)
	}
	return nil
}

func getWordlists(wordlists []Wordlist, dir string) error {
	// Create directory
	if err := createDirectory(dir); err != nil {
		return err
	}

	// Download wordlists
	for _, wordlist := range wordlists {
		// Check if wordlist exists
		wordlistPath := filepath.Join(dir, wordlist.Name+".txt")
		if _, err := os.Stat(wordlistPath); os.IsNotExist(err) {
			// Download wordlist
			logrus.Infof("Downloading %s wordlist", wordlist.Name)
			if err := downloadFile(wordlistPath, wordlist.URL); err != nil {
				return fmt.Errorf("could not download %s wordlist: %w", wordlist.Name, err)
			}
		}
		if wordlist.Name == "leaky-paths" {
			// open file and remove leading slash
			if err := removeLeadingSlash(wordlistPath); err != nil {
				return fmt.Errorf("could not process %s: %w", wordlistPath, err)
			}
		}
	}
	return nil
}

func removeLeadingSlash(wordlistPath string) error {
	file, err := os.OpenFile(wordlistPath, os.O_RDWR, 0644)
	if err != nil {
		return fmt.Errorf("could not open file: %w", err)
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
		return fmt.Errorf("could not read file: %w", err)
	}

	// Write the updated lines back to the file
	file.Seek(0, 0)
	file.Truncate(0)
	writer := bufio.NewWriter(file)
	for _, line := range lines {
		fmt.Fprintln(writer, line)
	}
	writer.Flush()
	return nil
}

func WordlistPath() {
	// Get home directory
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("Could not find home directory: %s", err)
	}
	// Check if wordlists directory exists
	wordlistsDir := filepath.Join(home, ".ffufw", "wordlists")
	if err := createDirectory(wordlistsDir); err != nil {
		logrus.Fatalf("Could not create wordlists directory: %s", err)
	}
}

type wordlistCategory struct {
	name      string
	wordlists []Wordlist
}

func GetWordlistsAll() {
	home, err := os.UserHomeDir()
	if err != nil {
		logrus.Fatalf("Could not find home directory: %s", err)
	}
	base := filepath.Join(home, ".ffufw", "wordlists")

	categories := []wordlistCategory{
		{"misc", MiscWordlists},
		{"iis", IisWordlists},
		{"php", PhpWordlists},
		{"java", JavaWordlists},
		{"api", ApiWordlists},
		{"python", PythonWordlists},
		{"ruby", RubyWordlists},
		{"sap", SapWordlists},
		{"nginx", NginxWordlists},
		{"adobe", AdobeWordlists},
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 4) // limit concurrent downloads
	errChan := make(chan error, len(categories))

	for _, cat := range categories {
		wg.Add(1)
		go func(c wordlistCategory) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			if err := getWordlists(c.wordlists, filepath.Join(base, c.name)); err != nil {
				errChan <- err
			}
		}(cat)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		logrus.Fatalf("Error downloading wordlists: %s", err)
	}
}
