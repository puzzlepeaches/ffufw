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
	stdout       bool
	quiet        bool
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
	rootCmd.Flags().BoolP("stdout", "s", false, "Print output to stdout")
	rootCmd.Flags().BoolP("quiet", "q", false, "Do not print additional information (silent mode)")

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

	// If stdout is set, set quiet to true
	if cmd.Flag("stdout").Changed {
		stdout = true
		quiet = true

		// Disable logging
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.ErrorLevel)
	}
	if cmd.Flag("quiet").Changed {
		quiet = true
		// Disable logging
		logrus.SetOutput(os.Stdout)
		logrus.SetLevel(logrus.ErrorLevel)
	}

	// Ensure config file exists
	if configFile == "~/.ffufrc" {
		configFile = os.ExpandEnv("$HOME/.ffufrc")
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			logrus.Errorf("Config file %s does not exist\n", configFile)

			logrus.Info("Continuing without config file")
			configFile = ""
		}

		logrus.Info("Using default config file ~/.ffufrc")
	} else {
		if _, err := os.Stat(configFile); os.IsNotExist(err) {
			configFile = os.ExpandEnv(configFile)
			if _, err := os.Stat(configFile); os.IsNotExist(err) {
				logrus.Errorf("Config file %s does not exist\n", configFile)
				logrus.Info("Continuing without config file")
				configFile = ""
			}
			logrus.Infof("Using config file %s", configFile)
		}
		logrus.Infof("Using config file %s", configFile)
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
		logrus.Info("Trying to install ffuf...")
		cmd := exec.Command("go", "install", "github.com/ffuf/ffuf/v2@latest")
		if err := cmd.Run(); err != nil {
			logrus.Errorf("Error installing ffuf: %v", err)
			logrus.Error("Install ffuf using: go install github.com/ffuf/ffuf/v2@latest")
			return
		}
	}

	// Ensure ffufPostprocessing is installed
	_, err = exec.LookPath("ffufPostprocessing")
	if err != nil {
		logrus.Error("ffufPostprocessing is not installed.")

		// Try to install ffufPostprocessing
		logrus.Info("Trying to install ffufPostprocessing...")
		cmd := exec.Command("go", "install", "github.com/Damian89/ffufPostprocessing@latest")
		if err := cmd.Run(); err != nil {
			logrus.Errorf("Error installing ffufPostprocessing: %v", err)
			logrus.Error("Install ffufPostprocessing using:  go install github.com/Damian89/ffufPostprocessing@latest")
			return
		}
	}

	sem := make(chan bool, concurrency)
	var wg sync.WaitGroup

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		url := scanner.Text()
		sem <- true
		wg.Add(1)
		go runFFUF(url, outputDir, wordlistPath, configFile, stdout, quiet, &wg, sem)
	}

	wg.Wait()
	close(sem)

	if err := scanner.Err(); err != nil {
		logrus.Errorf("Error reading URLs file: %v\n", err)
	}
}
