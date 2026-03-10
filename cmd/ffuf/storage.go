package cmd

type TechData struct {
	Url    string
	Iis    bool
	Apache bool
	Nginx  bool
	Php    bool
	Java   bool
	Python bool
	Api    bool
	Sap    bool
	Ruby   bool
	Adobe  bool
}

type Url struct {
	url       string
	fuzzUrl   string
	outputDir string
}

type FFUF struct {
	URL                    Url
	Tech                   map[string]bool
	Concurrency            int
	OutputDir              string
	FFUFPath               string
	FFUFPostprocessingPath string
	configFile             string
}

// FfufCommand represents a single ffuf scan command with structured arguments
type FfufCommand struct {
	BinaryPath    string   // path to ffuf binary
	Args          []string // command-line arguments
	OutputFile    string   // path to the JSON output file
	WordlistName  string   // human-readable name of the wordlist being used
	OutputDir     string   // output directory for bodies
}
