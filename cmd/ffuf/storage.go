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
	URL Url
	// Tech                   TechData
	Tech                   map[string]bool
	Concurrency            int
	OutputDir              string
	FFUFPath               string
	FFUFPostprocessingPath string
	configFile             string
}
