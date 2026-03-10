package cmd

import (
	"fmt"
	"os"
	"sync"
	"sync/atomic"
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

		urls = removeMicrosoftUrls(urls)
		totalURLs := len(urls)

		urlChan := make(chan string, concurrency)
		errChan := make(chan error, totalURLs*10)

		// Progress counters
		var urlsProcessed int64
		var urlsSkipped int64
		var urlsFailed int64
		var totalFindings int64

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
							logrus.Infof("WAF detected for URL: %s [%s] - skipping", url, waf)
							atomic.AddInt64(&urlsSkipped, 1)
							continue
						} else {
							logrus.Debugf("No WAF detected for URL: %s", url)
						}
					}

					current := atomic.AddInt64(&urlsProcessed, 1)
					logrus.Infof("[%d/%d] Starting: %s", current, totalURLs, url)

					fingerprints, err := detectTech(url)
					if err != nil {
						logrus.Warnf("[%d/%d] Failed tech detection: %s [%s]", current, totalURLs, url, err)
						errChan <- &urlError{url: url, err: err}
						atomic.AddInt64(&urlsFailed, 1)
						continue
					}

					techData := convertTech(url, fingerprints)
					ffufInstance, err := ffuf.NewFFUF(url, techData, concurrency, outputDir, ffufPath, ffufPostprocessingPath, configFile)
					if err != nil {
						logrus.Warnf("[%d/%d] Failed to initialize: %s [%s]", current, totalURLs, url, err)
						errChan <- err
						atomic.AddInt64(&urlsFailed, 1)
						continue
					}

					if customWordlist != "" {
						customWordlistPath = expandPath(customWordlist)
					} else {
						customWordlistPath = ""
					}

					techCommands, err := ffuf.TechCommands(ffufInstance, url, customWordlistPath)
					if err != nil {
						logrus.Warnf("[%d/%d] Failed to generate commands: %s [%s]", current, totalURLs, url, err)
						errChan <- &urlError{url: url, err: err}
						atomic.AddInt64(&urlsFailed, 1)
						continue
					}

					// Collect all results across wordlists for deduplication
					allResults := make(map[string]struct{})
					urlStartTime := time.Now()

					for i, techCommand := range techCommands {
						logrus.Infof("[%d/%d] Scanning %s with %s (%d/%d wordlists)",
							current, totalURLs, url, techCommand.WordlistName, i+1, len(techCommands))

						if err := ffuf.RunFfuf(techCommand, verbose); err != nil {
							logrus.Warnf("Failed: %s with %s [%s]", url, techCommand.WordlistName, err)
							errChan <- &urlError{url: url, err: err}
							continue
						}
						outputFile, err := ffuf.RunPostProcessing(ffufInstance, techCommand)
						if err != nil {
							logrus.Warnf("Post-processing failed: %s with %s [%s]", url, techCommand.WordlistName, err)
							errChan <- &urlError{url: url, err: err}
							continue
						}

						// Collect results for deduplication
						results, err := process.ParseOutput(outputFile)
						if err != nil {
							errChan <- &urlError{url: url, err: err}
							continue
						}
						for _, result := range results {
							allResults[result.URL] = struct{}{}
						}
					}

					// Submit deduplicated results once per URL
					uniqueCount := len(allResults)
					if uniqueCount > 0 {
						atomic.AddInt64(&totalFindings, int64(uniqueCount))

						if gowitnessAddress != "" {
							for resultURL := range allResults {
								if err := process.SubmitGowitness(gowitnessAddress, resultURL); err != nil {
									errChan <- &urlError{url: url, err: err}
								}
							}
							logrus.Infof("Submitted %d unique URLs to gowitness for %s", uniqueCount, url)
						}

						if replayProxy != "" {
							for resultURL := range allResults {
								if err := process.SubmitReplayProxy(replayProxy, resultURL); err != nil {
									errChan <- &urlError{url: url, err: err}
								}
							}
							logrus.Infof("Submitted %d unique URLs to replay proxy for %s", uniqueCount, url)
						}
					}

					logrus.Infof("[%d/%d] Finished: %s [%d findings, %s]",
						current, totalURLs, url, uniqueCount, time.Since(urlStartTime).Round(time.Second))
				}
			}()
		}

		go func() {
			wg.Wait()
			close(errChan)
		}()

		var errorCount int
		for range errChan {
			errorCount++
		}

		// Print scan summary
		processed := atomic.LoadInt64(&urlsProcessed)
		skipped := atomic.LoadInt64(&urlsSkipped)
		failed := atomic.LoadInt64(&urlsFailed)
		findings := atomic.LoadInt64(&totalFindings)
		logrus.Infof("Scan complete: %d URLs processed, %d skipped, %d failed, %d unique findings, %d total errors",
			processed, skipped, failed, findings, errorCount)
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
	rootCmd.Flags().StringVarP(&configFile, "config", "c", "", "Specify the config file for FFUF")
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
