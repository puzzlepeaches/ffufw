package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	wordlists "github.com/puzzlepeaches/ffufw/cmd/wordlists"
	"github.com/sirupsen/logrus"
)

func expandPath(path string) (string, error) {
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		path = filepath.Join(homeDir, path[2:])
	}
	return path, nil
}

func parseURL(url string, outputDir string) (*Url, error) {

	// Remove trailing slash
	if strings.HasSuffix(url, "/") {
		url = url[:len(url)-1]
	}

	// Craft fuzzing string URL/FUZZ
	fuzzUrl := url + "/FUZZ"

	// Format output directory from URL
	hostDir := strings.ReplaceAll(url, "://", "_")
	hostDir = strings.ReplaceAll(hostDir, "/", "_")

	// Create output directory
	outputDir = filepath.Join(outputDir, hostDir)
	if _, err := os.Stat(outputDir); os.IsNotExist(err) {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return nil, err
		}
	}

	// Create new Url instance
	urlInstance := &Url{
		url:       url,
		fuzzUrl:   fuzzUrl,
		outputDir: outputDir,
	}

	return urlInstance, nil
}

func techDataToMap(techData TechData) map[string]bool {
	return map[string]bool{
		"Iis":    techData.Iis,
		"Apache": techData.Apache,
		"Nginx":  techData.Nginx,
		"Php":    techData.Php,
		"Java":   techData.Java,
		"Python": techData.Python,
		"Api":    techData.Api,
		"Sap":    techData.Sap,
		"Ruby":   techData.Ruby,
		"Adobe":  techData.Adobe,
	}
}

func NewFFUF(url string, technologies TechData, concurrency int, outputDir string, ffufPath string, ffufPostprocessingPath string, configFile string) (*FFUF, error) {

	// Expand paths
	outputDir, err := expandPath(outputDir)
	if err != nil {
		return nil, err
	}
	ffufPath, err = expandPath(ffufPath)
	if err != nil {
		return nil, err
	}
	ffufPostprocessingPath, err = expandPath(ffufPostprocessingPath)
	if err != nil {
		return nil, err
	}
	configFile, err = expandPath(configFile)
	if err != nil {
		return nil, err
	}

	urlInstance, err := parseURL(url, outputDir)
	if err != nil {
		return nil, err
	}

	// Initialize FFUF instance
	ffufInstance := &FFUF{
		URL:                    *urlInstance,
		Tech:                   techDataToMap(technologies),
		Concurrency:            concurrency,
		FFUFPath:               ffufPath,
		FFUFPostprocessingPath: ffufPostprocessingPath,
		configFile:             configFile,
	}

	return ffufInstance, nil
}

// craftBaseArgs returns the base arguments common to all ffuf commands for this instance
func craftBaseArgs(ffufInstance *FFUF) []string {
	args := []string{"-u", ffufInstance.URL.fuzzUrl}

	if ffufInstance.configFile != "" && ffufInstance.configFile != "~/.ffufrc" {
		args = append(args, "-config", ffufInstance.configFile)
	}

	return args
}

func TechCommands(ffufInstance *FFUF, url string, customWordlist string) ([]FfufCommand, error) {

	commands := []FfufCommand{}

	// Define base path for wordlists
	wordlistPath, err := expandPath("~/.ffufw/wordlists")
	if err != nil {
		return nil, err
	}

	baseArgs := craftBaseArgs(ffufInstance)

	// Define a map of tech to wordlists
	techWordlists := map[string][]wordlists.Wordlist{
		"Iis":    wordlists.IisWordlists,
		"Php":    wordlists.PhpWordlists,
		"Java":   wordlists.JavaWordlists,
		"Api":    wordlists.ApiWordlists,
		"Python": wordlists.PythonWordlists,
		"Ruby":   wordlists.RubyWordlists,
		"Sap":    wordlists.SapWordlists,
		"Nginx":  wordlists.NginxWordlists,
		"Adobe":  wordlists.AdobeWordlists,
	}

	if customWordlist != "" {
		// Construct command for custom wordlist
		outputFile := filepath.Join(ffufInstance.URL.outputDir, "results.custom.json")
		args := append([]string{}, baseArgs...)
		args = append(args, "-w", customWordlist, "-of", "json", "-od", ffufInstance.URL.outputDir, "-o", outputFile)
		commands = append(commands, FfufCommand{
			BinaryPath:   ffufInstance.FFUFPath,
			Args:         args,
			OutputFile:   outputFile,
			WordlistName: "custom",
			OutputDir:    ffufInstance.URL.outputDir,
		})
	} else {

		for tech, enabled := range ffufInstance.Tech {
			if enabled {
				// Check if this tech has specialized wordlists
				wlists, hasWordlists := techWordlists[tech]
				if !hasWordlists || len(wlists) == 0 {
					logrus.Infof("No specialized wordlists for %s, using generic wordlists", tech)
					continue
				}

				// Define the tech folder and wordlist path
				folderName := strings.ToLower(tech)
				techWordlistPath := filepath.Join(wordlistPath, folderName)

				// Construct commands for each wordlist
				for _, wordlist := range wlists {
					wordlistFile := filepath.Join(techWordlistPath, wordlist.Name+".txt")
					outputFile := filepath.Join(ffufInstance.URL.outputDir, "results."+wordlist.Name+".json")

					args := append([]string{}, baseArgs...)
					args = append(args, "-w", wordlistFile, "-of", "json", "-od", ffufInstance.URL.outputDir, "-o", outputFile)
					commands = append(commands, FfufCommand{
						BinaryPath:   ffufInstance.FFUFPath,
						Args:         args,
						OutputFile:   outputFile,
						WordlistName: wordlist.Name,
						OutputDir:    ffufInstance.URL.outputDir,
					})
				}

				// Define the extension based on the tech
				var extension string
				switch tech {
				case "Iis":
					extension = ".aspx,.asp"
				case "Php":
					extension = ".php"
				case "Java":
					extension = ".jsp"
				case "Python":
					extension = ".py,.pyc"
				case "Ruby":
					extension = ".rb"
				case "Api":
					extension = ".json,.yaml"
				default:
					continue
				}

				// Construct command for the raft-large-words wordlist with tech-specific extensions
				// Use tech name in output file to prevent overwrites when multiple techs detected
				wordlistFile := filepath.Join(wordlistPath, "misc", "raft-large-words.txt")
				outputFile := filepath.Join(ffufInstance.URL.outputDir, fmt.Sprintf("results.raft-large-words-%s.json", folderName))

				args := append([]string{}, baseArgs...)
				args = append(args, "-w", wordlistFile, "-e", extension, "-of", "json", "-od", ffufInstance.URL.outputDir, "-o", outputFile)
				commands = append(commands, FfufCommand{
					BinaryPath:   ffufInstance.FFUFPath,
					Args:         args,
					OutputFile:   outputFile,
					WordlistName: fmt.Sprintf("raft-large-words (%s extensions)", folderName),
					OutputDir:    ffufInstance.URL.outputDir,
				})
			}
		}

		// Construct non-tech wordlists only once
		for _, wordlist := range wordlists.MiscWordlists {

			// Skip raft-large-words (handled above with tech-specific extensions)
			if wordlist.Name == "raft-large-words" {
				continue
			}

			// Construct wordlist path
			wordlistFile := filepath.Join(wordlistPath, "misc", wordlist.Name+".txt")
			outputFile := filepath.Join(ffufInstance.URL.outputDir, "results."+wordlist.Name+".json")

			args := append([]string{}, baseArgs...)
			args = append(args, "-w", wordlistFile, "-of", "json", "-od", ffufInstance.URL.outputDir, "-o", outputFile)
			commands = append(commands, FfufCommand{
				BinaryPath:   ffufInstance.FFUFPath,
				Args:         args,
				OutputFile:   outputFile,
				WordlistName: wordlist.Name,
				OutputDir:    ffufInstance.URL.outputDir,
			})
		}
	}

	return commands, nil

}

func RunFfuf(cmd FfufCommand, verbose bool) error {

	logrus.Debugf("Running ffuf: %s %s", cmd.BinaryPath, strings.Join(cmd.Args, " "))

	// Execute the command
	c := exec.Command(cmd.BinaryPath, cmd.Args...)

	// Show ffuf output in verbose mode
	if verbose {
		c.Stdout = os.Stdout
		c.Stderr = os.Stderr
	}

	err := c.Run()
	if err != nil {
		return err
	}

	return nil
}

func RunPostProcessing(ffufInstance *FFUF, cmd FfufCommand) (string, error) {

	// Define the command
	ppArgs := []string{
		"-overwrite-result-file",
		"-delete-all-bodies",
		"-bodies-folder", cmd.OutputDir,
		"-result-file", cmd.OutputFile,
	}

	logrus.Debugf("Running postprocessing: %s %s", ffufInstance.FFUFPostprocessingPath, strings.Join(ppArgs, " "))

	// Execute the command
	c := exec.Command(ffufInstance.FFUFPostprocessingPath, ppArgs...)
	err := c.Run()
	if err != nil {
		return cmd.OutputFile, err
	}

	return cmd.OutputFile, nil
}
