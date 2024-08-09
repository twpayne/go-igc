package main

import (
	"flag"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/twpayne/go-igc"
)

var igcExtensionRx = regexp.MustCompile(`(?i)\.igc\z`)

func run() error {
	flag.Parse()
	for _, arg := range flag.Args() {
		if err := fs.WalkDir(os.DirFS(arg), ".", func(path string, dirEntry fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !igcExtensionRx.MatchString(path) {
				return nil
			}
			if !dirEntry.Type().IsRegular() {
				return nil
			}
			file, err := os.Open(filepath.Join(arg, path))
			if err != nil {
				return err
			}
			defer file.Close()
			igcFile, err := igc.Parse(file)
			if err != nil {
				return err
			}
			if len(igcFile.Errs) == 0 {
				return nil
			}
			fmt.Println(filepath.Join(arg, path) + ":")
			for _, err := range igcFile.Errs {
				if !strings.HasSuffix(err.Error(), "invalid F record") {
					fmt.Println("- " + err.Error())
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
