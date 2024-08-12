package main

import (
	"context"
	"fmt"
	"os"

	"github.com/twpayne/go-igc/civlovs"
)

func validate(ctx context.Context, s *civlovs.Client, filename string) (civlovs.Status, *civlovs.Response, error) {
	f, err := os.Open(filename)
	if err != nil {
		return civlovs.StatusUnknown, nil, err
	}
	defer f.Close()
	return s.ValidateIGC(ctx, filename, f)
}

func main() {
	s := civlovs.NewClient()
	worstStatus := civlovs.StatusValid
	ctx := context.Background()
	for _, filename := range os.Args[1:] {
		status, _, err := validate(ctx, s, filename)
		switch status {
		case civlovs.StatusValid:
			fmt.Printf("%s: %s\n", filename, status)
		case civlovs.StatusInvalid:
			fmt.Printf("%s: %s: %s\n", filename, status, err)
			if worstStatus < 1 {
				worstStatus = 1
			}
		case civlovs.StatusUnknown:
			fmt.Printf("%s: %s: %s\n", filename, status, err)
			if worstStatus < 2 {
				worstStatus = 2
			}
		}
	}
	os.Exit(int(worstStatus))
}
