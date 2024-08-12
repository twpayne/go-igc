package civlovs_test

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-igc/civlovs"
)

func ExampleClient_ValidateIGC() {
	filename := "testdata/2006-06-10-XXX-3XI-01.IGC"
	igcFile, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer igcFile.Close()
	ctx := context.Background()
	if status, _, err := civlovs.NewClient().ValidateIGC(ctx, filename, igcFile); status == civlovs.StatusValid {
		fmt.Println("OK")
	} else {
		fmt.Println(err)
	}
	// Output: OK
}

func TestClient_ValidateIGC(t *testing.T) {
	ctx := context.Background()
	filename := "testdata/2006-06-10-XXX-3XI-01.IGC"
	data, err := os.ReadFile(filename)
	assert.NoError(t, err)
	for i, tc := range []struct {
		data           []byte
		expectedStatus civlovs.Status
	}{
		{
			data:           data,
			expectedStatus: civlovs.StatusValid,
		},
		{
			data:           regexp.MustCompile(`(?m)^G.*\r?\n`).ReplaceAll(data, nil),
			expectedStatus: civlovs.StatusInvalid,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			status, _, err := civlovs.NewClient().ValidateIGC(ctx, filename, bytes.NewReader(tc.data))
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedStatus, status)
		})
	}
}
