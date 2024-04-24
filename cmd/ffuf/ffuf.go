package cmd

import (
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
		// Create directory
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

func TechCommands(ffufInstance *FFUF, command string, url string, customWordlist string) ([]string, error) {

	// Define the command
	techCommands := []string{}

	// Define base path for wordlists
	wordlistPath, err := expandPath("~/.ffufw/wordlists")
	if err != nil {
		return nil, err
	}

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

	// Define a map for non-tech wordlists
	nonTechCommands := []string{}

	if customWordlist != "" {
		// Construct command for custom wordlist
		customCommand := command + " -w " + customWordlist + " -of json" + " -od " + ffufInstance.URL.outputDir + " -o " + ffufInstance.URL.outputDir + "/results.custom.json"
		techCommands = append(techCommands, customCommand)
	} else {

		for tech, enabled := range ffufInstance.Tech {
			if enabled {
				// Define the tech folder and wordlist path
				folderName := strings.ToLower(tech)
				techWordlistPath := filepath.Join(wordlistPath, folderName)

				// Construct commands for each wordlist
				for _, wordlist := range techWordlists[tech] {
					wordlistFile := filepath.Join(techWordlistPath, wordlist.Name) + ".txt"
					outputFile := ffufInstance.URL.outputDir + "/results." + wordlist.Name + ".json"

					techCommand := command + " -w " + wordlistFile + " -of json -od " + ffufInstance.URL.outputDir + " -o " + outputFile
					techCommands = append(techCommands, techCommand)
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

				// Construct command for the raft-large-words wordlist
				wordlistFile := filepath.Join(wordlistPath, "misc", "raft-large-words.txt")
				outputFile := ffufInstance.URL.outputDir + "/results.raft-large-words.json"

				techCommand := command + " -w " + wordlistFile + "-e " + extension + " -of json -od " + ffufInstance.URL.outputDir + " -o " + outputFile
				techCommands = append(techCommands, techCommand)
			}
		}

		// Construct non-tech wordlists only once
		for _, wordlist := range wordlists.MiscWordlists {

			// Skip raft-large-words
			// TODO I don't remember why I did this
			if wordlist.Name == "raft-large-words" {
				continue
			}

			// Construct wordlist path
			techWordlistPath := filepath.Join(wordlistPath, "misc", wordlist.Name) + ".txt"

			// Construct command for each wordlist from original
			techCommand := command + " -w " + techWordlistPath + " -of json" + " -od " + ffufInstance.URL.outputDir + " -o " + ffufInstance.URL.outputDir + "/results." + wordlist.Name + ".json"

			// Append command to nonTechCommands
			nonTechCommands = append(nonTechCommands, techCommand)

		}

		// Append non-tech commands to techCommands
		techCommands = append(techCommands, nonTechCommands...)
	}

	return techCommands, nil

}

func CraftCommand(ffufInstance *FFUF) string {

	// Define the command
	// command := ffufInstance.FFUFPath + " -u " + ffufInstance.URL.fuzzUrl + " -mc all "
	command := ffufInstance.FFUFPath + " -u " + ffufInstance.URL.fuzzUrl

	if ffufInstance.configFile != "" {
		command += " -config " + ffufInstance.configFile
	}

	return command

}

func RunFfuf(ffufInstance *FFUF, techCommand string) error {

	// Split the command string into command and arguments
	parts := strings.Fields(techCommand)
	cmd := exec.Command(parts[0], parts[1:]...)

	// Execute the command and capture the output
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

func RunPostProcessing(ffufInstance *FFUF, techCommand string) (string, error) {

	// Select output file from last item in techCommand
	outputFile := strings.Split(techCommand, " ")[len(strings.Split(techCommand, " "))-1]

	// Get wordlist name from output file
	wordlistName := strings.Split(outputFile, "/")[len(strings.Split(outputFile, "/"))-1]
	wordlistName = strings.TrimSuffix(wordlistName, filepath.Ext(wordlistName))

	// Define the command
	command := ffufInstance.FFUFPostprocessingPath +
		" -overwrite-result-file" +
		" -delete-all-bodies" +
		" -bodies-folder " + ffufInstance.URL.outputDir +
		" -result-file " + outputFile

	logrus.Debugf("Running postprocessing command: %s", command)

	// Split the command string into command and arguments
	parts := strings.Fields(command)
	cmd := exec.Command(parts[0], parts[1:]...)
	err := cmd.Run()
	if err != nil {
		return outputFile, err
	}

	return outputFile, err

}
