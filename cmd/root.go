package cmd

import (
	"fmt"
	"os"
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
	replayProxy            string
	customWordlist         string
	customWordlistPath     string
)

type urlError struct {
	url string
	err error
}

func (e *urlError) Error() string {
	return fmt.Sprintf("Error running FFUF: %s [URL: %s]", e.err, e.url)
}

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
		checkReplayProxy(replayProxy)
		if customWordlist != "" {
			checkCustomWordlist(customWordlist)
		}

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

		urlChan := make(chan string, concurrency)
		errChan := make(chan error, len(urls))

		go func() {
			for _, url := range urls {
				urlChan <- url
			}
			close(urlChan)
		}()

		var wg sync.WaitGroup
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			go func() {
				defer wg.Done()
				for url := range urlChan {

					if excludeWaf {
						waf, err := checks.CheckWaf(url)
						if err != nil {
							logrus.Debugf("Error checking WAF: %s", err)
						}
						if waf != "" {
							logrus.Infof("WAF detected for URL: %s", url)
							// Remove the URL from the slice
							urls = remove(urls, url)
							continue
						} else {
							logrus.Debugf("No WAF detected for URL: %s", url)
						}
					}

					fingerprints, err := detectTech(url)
					if err != nil {
						errChan <- &urlError{url: url, err: err}
						continue
					}

					techData := convertTech(url, fingerprints)
					ffufInstance, err := ffuf.NewFFUF(url, techData, concurrency, outputDir, ffufPath, ffufPostprocessingPath, configFile)
					if err != nil {
						errChan <- err
						continue
					}
					command := ffuf.CraftCommand(ffufInstance)

					if customWordlist != "" {
						customWordlistPath = expandPath(customWordlist)
					} else {
						customWordlistPath = ""
					}

					techCommands, err := ffuf.TechCommands(ffufInstance, command, url, customWordlistPath)
					if err != nil {
						errChan <- &urlError{url: url, err: err}
						continue
					}
					for _, techCommand := range techCommands {
						startTime := time.Now()
						logrus.Infof("Started scanning: %s", url)
						logrus.Debugf("Running command: %s", techCommand)
						if err := ffuf.RunFfuf(ffufInstance, techCommand); err != nil {
							errChan <- &urlError{url: url, err: err}
							continue
						}
						outputFile, err := ffuf.RunPostProcessing(ffufInstance, techCommand)
						if err != nil {
							errChan <- &urlError{url: url, err: err}
							continue
						}

						// Wait for ffufpostprocessing to finish before submitting to gowitness or replayproxy
						ffufPostProcessingDone := make(chan bool)
						go func() {
							for !<-ffufPostProcessingDone {
							}
							if gowitnessAddress != "" {
								results, err := process.ParseOutput(outputFile)
								if err != nil {
									errChan <- &urlError{url: url, err: err}
								}
								for _, result := range results {
									if err := process.SubmitGowitness(gowitnessAddress, result); err != nil {
										errChan <- &urlError{url: url, err: err}
										continue
									}
								}
								logrus.Infof("Submitted %d URLs to gowitness", len(results))
							}

							if replayProxy != "" {
								results, err := process.ParseOutput(outputFile)
								if err != nil {
									errChan <- &urlError{url: url, err: err}
								}
								for _, result := range results {
									if err := process.SubmitReplayProxy(replayProxy, result); err != nil {
										errChan <- &urlError{url: url, err: err}
										continue
									}
								}
								logrus.Infof("Submitted %d URLs to replay proxy", len(results))
							}
						}()
						ffufPostProcessingDone <- true

						endTime := time.Now()
						logrus.Infof("Finished scanning: %s [Duration: %s]", url, endTime.Sub(startTime))
					}

				}
			}()
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		for err := range errChan {
			urlErr, ok := err.(*urlError)
			if ok {
				logrus.Errorf("Error running FFUF: %s [URL: %s]", urlErr.err, urlErr.url)
			} else {
				logrus.Errorf("Error running FFUF: %s", err)
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
	rootCmd.Flags().StringVarP(&replayProxy, "replay-proxy", "r", "", "Specify the address for a replay proxy. Ensure format is http://<ip>:<port>")
	rootCmd.Flags().StringVarP(&customWordlist, "custom-wordlist", "w", "", "Specify a custom wordlist to use for scanning. This disable technology detection and pre-defined wordlists for all URLs.")

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
