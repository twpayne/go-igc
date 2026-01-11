module github.com/twpayne/go-igc

go 1.24.0

tool (
	github.com/twpayne/go-igc/cmd/parse-all
	github.com/twpayne/go-igc/cmd/parse-igc
	github.com/twpayne/go-igc/cmd/summarize-igc
	github.com/twpayne/go-igc/cmd/validate-igc
)

require (
	github.com/alecthomas/assert/v2 v2.10.0
	golang.org/x/text v0.33.0
)

require (
	github.com/alecthomas/repr v0.4.0 // indirect
	github.com/hexops/gotextdiff v1.0.3 // indirect
)
