package igc_test

import (
	"testing"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-igc"
)

func TestManufacturers(t *testing.T) {
	aircotec := igc.ManufacturersByTLC["ACT"]
	assert.Equal(t, &igc.Manufacturer{
		TLC:  "ACT",
		SCC:  'I',
		Name: "Aircotec",
	}, aircotec)
	assert.True(t, aircotec.Approved())

	ascent := igc.ManufacturersByTLC["XAH"]
	assert.Equal(t, &igc.Manufacturer{
		TLC:  "XAH",
		SCC:  'X',
		Name: "Ascent",
	}, ascent)
	assert.False(t, ascent.Approved())
}
