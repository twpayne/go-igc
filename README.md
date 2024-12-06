# go-igc

[![PkgGoDev](https://pkg.go.dev/badge/github.com/twpayne/go-igc)](https://pkg.go.dev/github.com/twpayne/go-igc)

Package `igc` handles [IGC
files](https://www.fai.org/page/igc-approved-flight-recorders).

## Features

* Robust, flexible parser for real IGC files, including common deviations from the IGC
  specification.
* Support for all IGC record types.
* Support for B record additions.
* Support for K record additions.
* Support for N record additions.
* Support for sub-second resolution timestamps with the `TDS` B record addition.
* Support for high-resolution coordinates with the `LAD` and `LOD` B record
  additions.
* Support for UTC midnight rollover.
* Support for [CIVL's Open Validation
  Server](http://vali.fai-civl.org/webservice.html).

## Validation

A simple command line client for CIVL's Open Validation server is included.
Install and run it with:

```bash
$ go install github.com/twpayne/go-igc/cmd/validate-igc@latest
$ validate-igc filename.igc
filename.igc: Valid
$ echo $?
0
```

The exit code is `0` if the IGC file is valid, `1` if it is invalid, or `2` if
it could not be validated.

## License

MIT