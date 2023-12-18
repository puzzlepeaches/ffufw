<h1 align="center">ffufw</h1>

<h3 align="center">ffuf with that special sauce</h3>

## Install

```
go install github.com/puzzlepeaches/ffufw@latest
```

## Why? 
`ffufw` is a wrapper around ffuf to make directory bruteforcing easier and more intelligent. The tool has the following features:

* Multithreaded execution of ffuf commands for a list of URLs
* Dynamic detection of technologies using gowappalyzer
* Dynamic generation of ffuf commands based on detected technologies (includes custom wordlists and extensions)
* Parsing of ffuf output to remove junk and identify potentially interesting endpoints
* The ability to pass discovered URLs to gowitness for review 
* The ability to exclude URLs utilizing a WAF from the scan 

## Getting started

This project requires Go to be installed. Install instructions can be found [here](https://golang.org/doc/install). Alternatively, you can quickly install go using the following command and [repo](https://github.com/canha/golang-tools-install-script):

```
wget -q -O - https://git.io/vQhTU | bash
```

Install the tool using the following command:

```
go install github.com/puzzlepeaches/ffufw@latest
```

The following tools are required for the tool to run:

* [ffuf](https://github.com/ffuf/ffuf)
* [ffufPostProcessing](https://github.com/Damian89/ffufPostprocessin)

Install the requirements using the following commands:

```
go install github.com/Damian89/ffufPostprocessing@latest
go install github.com/ffuf/ffuf/v2@latest
```

Wordlists, if not already present on your system will be downloaded on the first run to the directory `~/.ffufw/wordlists/`. For a list of all wordlists downloaded, see `cmd/wordlists/storage.go`. Custom wordlists are not currently supported.

## Usage

The help menu for the tool is as follows:

```
ffuf with that special sauce

Usage:
  ffufw [flags] -i <input file> -o <output directory>
  ffufw [command]

Available Commands:
  help        Help about any command
  version     Print the version number of the generated code example

Flags:
  -t, --concurrency int             Set the concurrency level for scanning (default 3)
  -c, --config string               Specify the config file for FFUF (default "~/.ffufrc")
  -e, --exclude-waf                 Exclude WAFs from the scans.
      --ffuf string                 Specify the path to the ffuf binary (default "ffuf")
      --ffufPostprocessing string   Specify the path to the ffufPostprocessing binary (default "ffufPostprocessing")
  -g, --gowitness string            Specify the address for the gowitness API. Ensure format is http://<ip>:<port>
  -h, --help                        help for ffufw
  -i, --input string                Specify the list of URLs to scan
  -o, --output string               Specify the output directory for FFUF results
  -q, --quiet                       Enable silent mode (no additional information printed)
  -r, --replay-proxy string         Specify the address for a replay proxy. Ensure format is http://<ip>:<port>
  -v, --verbose                     Enable verbose mode (print additional information)

Use "ffufw [command] --help" for more information about a command.
```

## Examples

Very basic usage of the tool with a custom ffuf config file and verbose output:

```
ffufw -o /tmp/output/ -i /tmp/targets.txt -c /opt/.ffufrc -v
```

Basic usage with the output being shipped to gowitness:

```
ffufw -o /tmp/output/ -i /tmp/targets.txt -g http://127.0.0.1:9999
```

Usage with custom ffuf and ffufPostprocessing binaries:

```
ffufw --ffuf /usr/local/bin/ffuf --ffufPostprocessing /usr/local/bin/ffufPostprocessing -o /tmp/output/ -i /tmp/targets.txt
```

Usage with custom concurrency (number of URLs to scan at once):

```
ffufw -o /tmp/output/ -i /tmp/targets.txt -c /opt/.ffufrc -t 5
```

Basic usage with gowitness, verbose output, and WAF exclusion:

```
ffufw -o /tmp/output/ -i /tmp/urls.txt -c /opt/.ffufrc -v -e -g http://127.0.0.1:9000
```

Basic usage with 5 threads and submission to a replay proxy (Burp, Zap, etc):

```
ffufw -o /tmp/output/ -i /tmp/urls.txt -c /opt/.ffufrc -t 5 -r http://127.0.0.1:8080
```

## TODO

- Support for custom wordlists
- Refactor to support easy additions of technology check additions
- Ability to ignore certain technologies
- Ability to add custom technologies
- Ability to specify single wordlists for all URLs
- Better logging and error handling


## References & Thanks 

* [gowappalyzer](https://github.com/projectdiscovery/wappalyzergo)
* [ffuf](https://github.com/ffuf/ffuf)
* [ffufPostProcessing](https://github.com/Damian89/ffufPostprocessing)
* [gowitness](https://github.com/sensepost/gowitness)
