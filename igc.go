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
type Record interface {
	Type() byte
	Valid() bool
}

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

func (r ARecord) Type() byte  { return 'A' }
func (r ARecord) Valid() bool { return true }

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

func (r BRecord) Type() byte  { return 'B' }
func (r BRecord) Valid() bool { return true }

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

func (r FirstCRecord) Type() byte  { return 'C' }
func (r FirstCRecord) Valid() bool { return true }

// A CRecord is a C record.
type CRecord struct {
	Lat  float64
	Lon  float64
	Text string
}

func (r CRecord) Type() byte  { return 'C' }
func (r CRecord) Valid() bool { return true }

// A DRecord is a D record.
type DRecord struct {
	GPSQualifier  GPSQualifier
	DGPSStationID int
}

func (r DRecord) Type() byte  { return 'D' }
func (r DRecord) Valid() bool { return true }

// An ERecord is an E record.
type ERecord struct {
	Time time.Time
	TLC  string
	Text string
}

func (r ERecord) Type() byte  { return 'E' }
func (r ERecord) Valid() bool { return true }

// An ERecordWithoutTLC is an E record without a three-letter code.
type ERecordWithoutTLC struct {
	Time time.Time
	Text string
}

func (r ERecordWithoutTLC) Type() byte  { return 'E' }
func (r ERecordWithoutTLC) Valid() bool { return false }

// An FRecord is an F record.
type FRecord struct {
	Time         time.Time
	SatelliteIDs []int
}

func (r FRecord) Type() byte  { return 'F' }
func (r FRecord) Valid() bool { return true }

// An HRecord is an H record.
type HRecord struct {
	Source   Source
	TLC      string
	LongName string
	Value    string
}

func (r HRecord) Type() byte  { return 'H' }
func (r HRecord) Valid() bool { return true }

// An HRecordWithInvalidSource is an H record.
type HRecordWithInvalidSource struct {
	Source   string
	TLC      string
	LongName string
	Value    string
}

func (r HRecordWithInvalidSource) Type() byte  { return 'H' }
func (r HRecordWithInvalidSource) Valid() bool { return false }

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

func (r GRecord) Type() byte { return 'G' }

// Valid returns whether the G record does not have any syntax errors. Since G
// records have freeform text, they never contain syntax errors, so this
// function always returns true.
//
// To determine whether a G record is valid cryptographically, use
// [github.com/twpayne/go-igc/civlovs] instead.
func (r GRecord) Valid() bool { return true }

// An IRecord is an I record.
type IRecord struct {
	Additions []BKRecordAddition
}

func (r IRecord) Type() byte  { return 'I' }
func (r IRecord) Valid() bool { return true }

// A JRecord is a J record.
type JRecord struct {
	Additions []BKRecordAddition
}

func (r JRecord) Type() byte  { return 'J' }
func (r JRecord) Valid() bool { return true }

// A KRecord is a K record.
type KRecord struct {
	Time      time.Time
	Additions map[string]int
}

func (r KRecord) Type() byte  { return 'K' }
func (r KRecord) Valid() bool { return true }

// An LRecord is an L record.
type LRecord struct {
	Input string
	Text  string
}

func (r LRecord) Type() byte  { return 'L' }
func (r LRecord) Valid() bool { return true }

// An LRecordWithoutTLC is an L record without a three-letter code.
type LRecordWithoutTLC struct {
	Text string
}

func (r LRecordWithoutTLC) Type() byte  { return 'L' }
func (r LRecordWithoutTLC) Valid() bool { return false }

// An IGC is a parsed IGC file.
type IGC struct {
	Records       []Record
	BRecords      []*BRecord
	HRecordsByTLC map[string]*HRecord
	KRecords      []*KRecord
	Errs          []error
}

// Parse parses an IGC from r.
func Parse(r io.Reader) (*IGC, error) {
	return newParser().parse(r)
}

// Parse parses an IGC from lines.
func ParseLines(lines []string) (*IGC, error) {
	return newParser().parseLines(lines)
}
