package cmd

import (
	"bufio"
	"os"
	"os/exec"
	"sync"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	concurrency  int
	outputDir    string
	wordlistPath string
	configFile   string
)

var rootCmd = &cobra.Command{
	Use:   "ffufw [urls_file]",
	Short: "A FFUF wrapper with concurrency support",
	Long: `ffufw is a command-line tool that wraps the FFUF (Fuzz Faster U Fool) 
utility, allowing you to run concurrent scans against multiple targets and post-process the results.`,
	Args: cobra.ExactArgs(1),
	Run:  runCommand,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logrus.Error(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "t", 3, "Concurrency level for scanning")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Output directory for FFUF results")
	rootCmd.Flags().StringVarP(&wordlistPath, "wordlist", "w", "", "Path to the wordlist")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "~/.ffufrc", "Config file for FFUF")

	rootCmd.MarkFlagRequired("output")
	rootCmd.MarkFlagRequired("wordlist")
}

func runCommand(cmd *cobra.Command, args []string) {
	urlsFile := args[0]

	file, err := os.Open(urlsFile)

	if err != nil {
		logrus.Errorf("Error opening URLs file: %v\n", err)
		return
	}
	defer file.Close()

	// Ensure config file exists
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		logrus.Errorf("Config file %s does not exist\n", configFile)
		logrus.Info("Continuing without config file")
		configFile = ""
	}

	// Ensure wordlist file exists
	if _, err := os.Stat(wordlistPath); os.IsNotExist(err) {
		logrus.Errorf("Wordlist file %s does not exist\n", wordlistPath)
		return
	}

	// Ensure ffuf is installed
	_, err = exec.LookPath("ffuf")
	if err != nil {
		logrus.Error("FFUF is not installed.")
		logrus.Error("Install ffuf using: go install github.com/ffuf/ffuf/v2@latest")
		return
	}

	// Ensure ffufPostprocessing is installed
	_, err = exec.LookPath("ffufPostprocessing")
	if err != nil {
		logrus.Error("ffufPostprocessing is not installed.")
		logrus.Error("Install ffufPostprocessing using:  go install github.com/Damian89/ffufPostprocessing@latest")
		return
	}

	sem := make(chan bool, concurrency)
	var wg sync.WaitGroup

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		sem <- true
		wg.Add(1)
		go runFFUF(url, outputDir, wordlistPath, configFile, &wg, sem)
	}

	wg.Wait()
	close(sem)

	if err := scanner.Err(); err != nil {
		logrus.Errorf("Error reading URLs file: %v\n", err)
	}
}
