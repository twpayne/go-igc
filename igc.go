// Package igc parses IGC files.
//
// See https://www.fai.org/sites/default/files/igc_fr_specification_with_al8_2023-2-1_0.pdf.
package igc

import (
	"io"
	"strconv"
	"time"
)

// An Error is an error at a line.
type Error struct {
	Line int
	Err  error
}

func (e *Error) Error() string {
	return strconv.Itoa(e.Line) + ": " + e.Err.Error()
}

func (e *Error) Unwrap() error {
	return e.Err
}

// A Record is a record.
type Record any

// A Source is the data source of an H record.
type Source byte

// Sources.
const (
	SourceFlightRecorder Source = 'F'
	SourceOther          Source = 'O'
	SourcePilot          Source = 'P'
)

// A Validity is a GPS fix validity.
type Validity byte

// Validities.
const (
	Validity2D Validity = 'V'
	Validity3D Validity = 'A'
)

// A GPSQualifier is a GPS qualifier.
type GPSQualifier byte

// GPSQualifiers.
const (
	GPSQualifierGPS  GPSQualifier = '1'
	GPSQualifierDGPS GPSQualifier = '2'
)

// An BKRecordAddition is an addition to a B or K record.
type BKRecordAddition struct {
	TLC          string
	StartColumn  int
	FinishColumn int
}

// An ARecord is an A record.
type ARecord struct {
	ManufacturerID         string
	UniqueFlightRecorderID string
	AdditionalData         string
}

// A BRecord is a B record.
type BRecord struct {
	Time          time.Time
	Lat           float64
	Lon           float64
	Validity      Validity
	AltWGS84      float64
	AltBarometric float64
	Additions     map[string]int
}

// A FirstCRecord is a first C record.
type FirstCRecord struct {
	DeclarationTime    time.Time
	FlightYear         int
	FlightMonth        int
	FlightDay          int
	TaskNumber         int
	NumberOfTurnpoints int
	Text               string
}

// A CRecord is a C record.
type CRecord struct {
	Lat  float64
	Lon  float64
	Text string
}

// A DRecord is a D record.
type DRecord struct {
	GPSQualifier  GPSQualifier
	DGPSStationID int
}

// An ERecord is an E record.
type ERecord struct {
	Time time.Time
	TLC  string
	Text string
}

// An FRecord is an F record.
type FRecord struct {
	Time         time.Time
	SatelliteIDs []int
}

// An HRecord is an H record.
type HRecord struct {
	Source   Source
	TLC      string
	LongName string
	Value    string
}

// An HFDTERecord is an HFDTE record.
type HFDTERecord struct {
	HRecord
	Date         time.Time
	FlightNumber int
}

// A GRecord is a G record.
type GRecord struct {
	Text string
}

// An IRecord is an I record.
type IRecord struct {
	Additions []BKRecordAddition
}

// A JRecord is a J record.
type JRecord struct {
	Additions []BKRecordAddition
}

// A KRecord is a K record.
type KRecord struct {
	Time      time.Time
	Additions map[string]int
}

// An LRecord is an L record.
type LRecord struct {
	Input string
	Text  string
}

// An IGC is a parsed IGC file.
type IGC struct {
	Records       []Record
	BRecords      []*BRecord
	HRecordsByTLC map[string]Record
	KRecords      []*KRecord
	Errs          []error
}

// Parse parses an IGC from r.
func Parse(r io.Reader) (*IGC, error) {
	return newParser().parse(r)
}
