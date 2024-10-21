package igc

import (
	"fmt"
	"strings"
)

// A Manufacturer is a manufacturer.
type Manufacturer struct {
	Name string
	TLC  string
	SCC  byte
}

// ApprovedManufacturers is the list of approved manufacturers.
var ApprovedManufacturers = []Manufacturer{
	{TLC: "ACT", SCC: 'I', Name: "Aircotec"},
	{TLC: "AVX", Name: "Avionix"},
	{TLC: "CAM", SCC: 'C', Name: "Cambridge Aero Instruments"},
	{TLC: "CNI", Name: "ClearNav Instruments"},
	{TLC: "DSX", SCC: 'D', Name: "Data Swan/DSX"},
	{TLC: "EWA", SCC: 'E', Name: "EW Avionics"},
	{TLC: "FIL", SCC: 'F', Name: "Filser"},
	{TLC: "FLA", SCC: 'G', Name: "Flarm (Flight Alarm)"},
	{TLC: "FLY", Name: "Flytech"},
	{TLC: "GCS", SCC: 'A', Name: "Garrecht"},
	{TLC: "IMI", SCC: 'M', Name: "IMI Gliding Equipment"},
	{TLC: "LGS", Name: "Logstream"},
	{TLC: "LXN", SCC: 'L', Name: "LX Navigation"},
	{TLC: "LXV", SCC: 'V', Name: "LXNAV d.o.o."},
	{TLC: "NAV", Name: "Naviter"},
	{TLC: "NTE", SCC: 'N', Name: "New Technologies s.r.l."},
	{TLC: "NKL", SCC: 'K', Name: "Nielsen Kellerman"},
	{TLC: "PES", SCC: 'P', Name: "Peschges"},
	{TLC: "PFE", Name: "PressFinish Electronics"},
	{TLC: "PRT", SCC: 'R', Name: "Print Technik"},
	{TLC: "RCE", Name: "RC Electronics"},
	{TLC: "SCH", SCC: 'H', Name: "Scheffel"},
	{TLC: "SDI", SCC: 'S', Name: "Streamline Data Instruments"},
	{TLC: "TRI", SCC: 'T', Name: "Triadis Engineering GmbH"},
	{TLC: "ZAN", SCC: 'Z', Name: "Zander"},
}

// NonApprovedManufacturers is an unofficial list of non-approved manufacturers.
var NonApprovedManufacturers = []Manufacturer{
	{TLC: "XAH", SCC: 'X', Name: "Ascent"},
	{TLC: "XBM", SCC: 'X', Name: "Burnair"},
	{TLC: "XCM", SCC: 'X', Name: "Naviter"},
	{TLC: "XCS", SCC: 'X', Name: "XCSoar"},
	{TLC: "XCT", SCC: 'X', Name: "XC Track"},
	{TLC: "XGD", SCC: 'X', Name: "GpsDump"},
	{TLC: "XFH", SCC: 'X', Name: "Flyskyhy"},
	{TLC: "XFL", SCC: 'X', Name: "FlyMe"},
	{TLC: "XFM", SCC: 'X', Name: "Flymaster"},
	{TLC: "XLK", SCC: 'X', Name: "LK8000"},
	{TLC: "XNA", SCC: 'X', Name: "Naviter"},
	{TLC: "XSD", SCC: 'X', Name: "Stodeus"},
	{TLC: "XSE", SCC: 'X', Name: "Syride"},
	{TLC: "XSR", SCC: 'X', Name: "Syride"},
	{TLC: "XSX", SCC: 'X', Name: "Skytraxx"},
	{TLC: "XTR", SCC: 'X', Name: "XC Tracer"},
	{TLC: "XTT", SCC: 'X', Name: "LiveTrack24"},
	{TLC: "XVB", SCC: 'X', Name: "VairBration"},
}

var (
	// ApprovedManufacturersByTLC is a map of three-letter codes to approved manufacturers.
	ApprovedManufacturersByTLC map[string]*Manufacturer

	// ManufacturersByTLC is a map of three-letter codes to approved or non-approved manufacturers.
	ManufacturersByTLC map[string]*Manufacturer
)

func init() {
	ApprovedManufacturersByTLC = make(map[string]*Manufacturer, len(ApprovedManufacturers))
	ManufacturersByTLC = make(map[string]*Manufacturer, len(ApprovedManufacturers)+len(NonApprovedManufacturers))
	for i, manufacturer := range ApprovedManufacturers {
		if _, ok := ApprovedManufacturersByTLC[manufacturer.TLC]; ok {
			panic(fmt.Sprintf("%s: duplicate manufacturer", manufacturer.TLC))
		}
		ApprovedManufacturersByTLC[manufacturer.TLC] = &ApprovedManufacturers[i]
		ManufacturersByTLC[manufacturer.TLC] = &ApprovedManufacturers[i]
	}
	for i, manufacturer := range NonApprovedManufacturers {
		if _, ok := ManufacturersByTLC[manufacturer.TLC]; ok {
			panic(fmt.Sprintf("%s: duplicate manufacturer", manufacturer.TLC))
		}
		ManufacturersByTLC[manufacturer.TLC] = &NonApprovedManufacturers[i]
	}
}

// Approved returns whether m is approved.
func (m *Manufacturer) Approved() bool {
	return !strings.HasPrefix(m.TLC, "X")
}
