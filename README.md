# go-igc

[![PkgGoDev](https://pkg.go.dev/badge/github.com/twpayne/go-igc)](https://pkg.go.dev/github.com/twpayne/go-igc)

Package `igc` parses [IGC
files](https://www.fai.org/sites/default/files/igc_fr_specification_with_al8_2023-2-1_0.pdf)
robustly.

## Features

* Flexible parser for real IGC files, including common deviations from the IGC
  specification.
* Support for all IGC record types.
* Support for B record additions.
* Support for K record additions.
* Support for sub-second resolution timestamps with the `TDS` B record addition.
* Support for high-resolution coordinates with the `LAD` and `LOD` B record
  additions.
* Support for UTC midnight rollover.

## License

MIT