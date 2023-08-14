# ffufw

ffuf with that special sauce


## Description
ffuf by default isn't great at handling multiple URLs and detecting junk URL paths. This tool aims to fix that by providing a wrapper around ffuf that will handle the following:

* Read a file of URLs and run ffuf against each one with concurrency
* Detect junk paths and remove them from the output using [ffufPostProcessing](https://github.com/Damian89/ffufPostprocessing).

## Getting started

This project requires Go to be installed. Install instructions can be found [here](https://golang.org/doc/install). 

Running it then should be as simple as:

```console
$ make
$ ./bin/ffufw
```

Install is also possible using go install:

```
go install github.com/puzzlepeaches/ffufw@latest
```

## Usage

```
Usage:
  ffufw [urls_file] [flags]
  ffufw [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  version     Print the version number of generated code example

Flags:
  -t, --concurrency int   Concurrency level for scanning (default 3)
  -c, --config string     Config file for FFUF (default "~/.ffufrc")
  -h, --help              help for ffufw
  -o, --output string     Output directory for FFUF results
  -w, --wordlist string   Path to the wordlist

Use "ffufw [command] --help" for more information about a command.
```

The goal of this tool is to run ffuf and properly process output for review. Instead of hoping -ac did its job, we instead using ffufPostProcessing to do the work for us. 

### Testing

``make test``