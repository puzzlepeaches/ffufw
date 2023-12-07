package cmd

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	checks "github.com/puzzlepeaches/ffufw/cmd/checks"
	ffuf "github.com/puzzlepeaches/ffufw/cmd/ffuf"
	process "github.com/puzzlepeaches/ffufw/cmd/process"
	wordlists "github.com/puzzlepeaches/ffufw/cmd/wordlists"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	concurrency            int
	configFile             string
	ffufPath               string
	ffufPostprocessingPath string
	inputFile              string
	outputDir              string
	gowitnessAddress       string
	quiet                  bool
	verbose                bool
	excludeWaf             bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "ffufw [flags] -i <input file> -o <output directory>",
	Short: "ffuf with that special sauce",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {

		// If no flags are specified, print help
		if len(os.Args) == 1 {
			cmd.Help()
			os.Exit(0)
		}

		// Initialize functions
		setLogging(quiet, verbose)
		ffufPath = checkBinary(ffufPath, "ffuf")
		ffufPostprocessingPath = checkBinary(ffufPostprocessingPath, "ffufPostprocessing")

		// Check if output directory and input file are specified
		if outputDir == "" {
			logrus.Fatalf("Output directory must be specified!")
		}
		if inputFile == "" {
			logrus.Fatalf("Input file must be specified!")
		}

		// Check output directory, input file, ffuf config and gowitness
		checkOutput(outputDir)
		checkInput(inputFile)
		checkFfufConfig(configFile)
		checkGowitness(gowitnessAddress)

		// Create wordlist directory and get all wordlists
		createWordlistDir()
		wordlists.WordlistPath()
		wordlists.GetWordlistsAll()

	},
	Run: func(cmd *cobra.Command, args []string) {

		// Read the input file
		urls, err := readInputFile(inputFile)
		if err != nil {
			logrus.Fatalf("Could not read input file at %s", inputFile)
		}

		urls = removMicrosoftUrls(urls)

		if excludeWaf {
			for _, url := range urls {
				waf, err := checks.CheckWaf(url)
				if err != nil {
					logrus.Debugf("Error checking WAF: %s", err)
				}
				if waf != "" {
					logrus.Infof("WAF detected for URL: %s", url)
					// Remove the URL from the slice
					urls = remove(urls, url)
				} else {
					logrus.Debugf("No WAF detected for URL: %s", url)
				}
			}
		}

		urlChan := make(chan string, concurrency)
		// errChan := make(chan error, concurrency)
		errChan := make(chan error, len(urls))
		urlsBeingScanned := make(map[string]bool)

		var mutex = sync.Mutex{}

		for i := 0; i < concurrency; i++ {
			go func() {
				for url := range urlChan {
					// Check if the URL is already being scanned
					mutex.Lock()
					if urlsBeingScanned[url] {
						mutex.Unlock()
						continue
					}
					urlsBeingScanned[url] = true
					mutex.Unlock()

					fingerprints, err := detectTech(url)
					if err != nil {
						errChan <- err
						continue
					}

					techData := convertTech(url, fingerprints)
					ffufInstance, err := ffuf.NewFFUF(url, techData, concurrency, outputDir, ffufPath, ffufPostprocessingPath, configFile)
					if err != nil {
						errChan <- err
						continue
					}
					command := ffuf.CraftCommand(ffufInstance)
					techCommands, err := ffuf.TechCommands(ffufInstance, command, url)
					if err != nil {
						errChan <- err
						continue
					}
					for _, techCommand := range techCommands {
						startTime := time.Now()
						logrus.Infof("Started scanning: %s", url)
						logrus.Debugf("Running command: %s", techCommand)
						if err := ffuf.RunFfuf(ffufInstance, techCommand); err != nil {
							errChan <- err
							continue
						}
						outputFile, err := ffuf.RunPostProcessing(ffufInstance, techCommand)
						if err != nil {
							errChan <- err
							continue
						}
						if gowitnessAddress != "" {
							results, err := process.ParseOutput(outputFile)
							if err != nil {
								errChan <- err
								continue
							}
							for _, result := range results {
								if err := process.SubmitURL(gowitnessAddress, result); err != nil {
									errChan <- err
									continue
								}
							}

						}
						endTime := time.Now()
						logrus.Infof("Finished scanning: %s [Duration: %s]", url, endTime.Sub(startTime))
					}

					// After processing the URL, remove it from the map
					mutex.Lock()
					delete(urlsBeingScanned, url)
					mutex.Unlock()
				}
			}()
		}
		for _, url := range urls {
			select {
			case err := <-errChan:
				// Handling timeout errors
				if strings.Contains(err.Error(), "Client.Timeout exceeded while awaiting headers") {
					logrus.Debugf("Timeout error: %s", err)
					// TODO: Remove URL from map
					continue
				} else {
					logrus.Errorf("Error running FFUF: %s", err)
					continue
				}
			default:
				urlChan <- url
			}
		}

		close(urlChan)

		// Handle errors
		// for i := 0; i < len(urls); i++ {
		// 	err := <-errChan
		// 	if err != nil {
		// 		logrus.Errorf("Error running FFUF: %s", err)
		// 		continue
		// 	}
		// }
		// Use a select statement to attempt to read from the errChan channel
		for i := 0; i < len(urls); i++ {
			select {
			case err := <-errChan:
				logrus.Errorf("Error running FFUF: %s", err)
				continue
				// default:
				// break
			}
		}
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {

	// Define command line arguments
	rootCmd.Flags().IntVarP(&concurrency, "concurrency", "t", 3, "Set the concurrency level for scanning")
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "~/.ffufrc", "Specify the config file for FFUF")
	rootCmd.Flags().StringVarP(&gowitnessAddress, "gowitness", "g", "", "Specify the address for the gowitness API. Ensure format is http://<ip>:<port>")
	rootCmd.Flags().BoolVarP(&excludeWaf, "exclude-waf", "e", false, "Exclude WAFs from the scans.")

	// Define paths for binary files
	rootCmd.Flags().StringVarP(&ffufPath, "ffuf", "", "ffuf", "Specify the path to the ffuf binary")
	rootCmd.Flags().StringVarP(&ffufPostprocessingPath, "ffufPostprocessing", "", "ffufPostprocessing", "Specify the path to the ffufPostprocessing binary")

	// Define required flags
	rootCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Specify the list of URLs to scan")
	rootCmd.Flags().StringVarP(&outputDir, "output", "o", "", "Specify the output directory for FFUF results")
	rootCmd.MarkFlagRequired("output")
	rootCmd.MarkFlagRequired("input")

	// Define display options
	rootCmd.Flags().BoolVarP(&quiet, "quiet", "q", false, "Enable silent mode (no additional information printed)")
	rootCmd.Flags().BoolVarP(&verbose, "verbose", "v", false, "Enable verbose mode (print additional information)")

}
