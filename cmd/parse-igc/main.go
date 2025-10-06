// parse-igc parses the IGC files passed to it on the command line and prints
// any errors.
package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/twpayne/go-igc"
)

func parseFile(filename string, options []igc.ParseOption) (*igc.IGC, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return igc.Parse(file, options...)
}

func run() error {
	allowInvalidChars := flag.Bool("allow-invalid-chars", true, "allow invalid characters")
	flag.Parse()
	options := []igc.ParseOption{
		igc.WithAllowInvalidChars(*allowInvalidChars),
	}
	allOK := true
	for _, arg := range flag.Args() {
		switch igcResult, err := parseFile(arg, options); {
		case err != nil:
			return err
		case len(igcResult.Errs) == 0:
			fmt.Println(arg + ": ok")
		default:
			allOK = false
			for _, igcErr := range igcResult.Errs {
				var igcError *igc.Error
				if errors.As(igcErr, &igcError) {
					fmt.Printf("%s:%d: %v\n", arg, igcError.Line, igcError.Err)
				} else {
					fmt.Printf("%s: %v\n", arg, igcErr)
				}
			}
		}
	}
	if !allOK {
		os.Exit(1)
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
