package igc

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strconv"
	"time"
	"unicode"
)

var errNoDate = errors.New("no date")

type atoiSyntaxError struct {
	num string
}

func (e *atoiSyntaxError) Error() string {
	return strconv.Quote(e.num) + ": syntax error"
}

type invalidAdditionError struct {
	addition RecordAddition
	message  string
}

func (e *invalidAdditionError) Error() string {
	return e.addition.TLC + ": " + e.message
}

type invalidRecordError byte

func (e invalidRecordError) Error() string {
	return "invalid " + string(e) + " record"
}

type missingAdditionError struct {
	addition RecordAddition
}

func (e *missingAdditionError) Error() string {
	return "missing " + e.addition.TLC + " addition"
}

type unknownRecordTypeError byte

func (e unknownRecordTypeError) Error() string {
	if unicode.IsPrint(rune(e)) {
		return string(e) + ": unknown record type"
	}
	return fmt.Sprintf(`"\x%02X": unknown record type`, byte(e))
}

var (
	invalidCharsRx = regexp.MustCompile(`([^\x20\x22-\x23\x25-\x29\x2b-\x5b\x5d\x5f-\x7d])`)

	aRecordRx                  = regexp.MustCompile(`\AA([A-Z]{3})(.*)\z`)
	bRecordRx                  = regexp.MustCompile(`\AB(\d{2})(\d{2})(\d{2})(\d{2})(\d{5})([NS])(\d{3})(\d{5})([EW])([AV])([0-9\-]\d{4})([0-9\-]\d{4})(.*)\z`)
	cRecordDeclarationRx       = regexp.MustCompile(`\AC(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})(\d{2})(\d{4})([0-9\-]\d)(.*)\z`)
	cRecordWaypointRx          = regexp.MustCompile(`\AC(\d{2})(\d{5})([NS])(\d{3})(\d{5})([EW])(.*)\z`)
	dRecordRx                  = regexp.MustCompile(`\AD([12])(\d{4})\z`)
	eRecordRx                  = regexp.MustCompile(`\AE(\d{2})(\d{2})(\d{2})([A-Z]{3})(.*)\z`)
	eRecordWithoutTLCRx        = regexp.MustCompile(`\AE(\d{2})(\d{2})(\d{2})(.*)\z`)
	fRecordRx                  = regexp.MustCompile(`\AF(\d{2})(\d{2})(\d{2})((?:\d{2})*)\z`)
	hRecordRx                  = regexp.MustCompile(`\AH([FOP])([0-9A-Z]{3})(.*?)(?::(.*))?\z`)
	hRecordWithInvalidSourceRx = regexp.MustCompile(`\AH([A-Z])([0-9A-Z]{3})(.*?)(?::(.*))?\z`)
	hfdteRecordRx              = regexp.MustCompile(`\AHFDTE(\d{2})(\d{2})(\d{2})\z`)
	hfdteRecordValueRx         = regexp.MustCompile(`(\d{2})(\d{2})(\d{2})(?:,(\d{2}))?\z`)
	hffxaRecordRx              = regexp.MustCompile(`\AHFFXA(\d+)\z`)
	gRecordRx                  = regexp.MustCompile(`\AG(.*)\z`)
	ijmRecordRx                = regexp.MustCompile(`\A[IJM](\d{2})((?:\d{4}[A-Z]{3})*)\z`)
	knRecordRx                 = regexp.MustCompile(`\A[KN](\d{2})(\d{2})(\d{2})(.*)\z`)
	lRecordRx                  = regexp.MustCompile(`\AL([A-Z]{3})(.*)\z`)
	lRecordWithoutTLCRx        = regexp.MustCompile(`\AL(.*)\z`)
)

type parser struct {
	allowInvalidChars      bool
	date                   time.Time
	prevTime               time.Time
	cRecords               []Record
	bRecordAdditions       []RecordAddition
	bRecordsAdditionsByTLC map[string]*RecordAddition
	ladBRecordAddition     *RecordAddition
	latMinMul              int
	latMinDiv              float64
	lodBRecordAddition     *RecordAddition
	lonMinMul              int
	lonMinDiv              float64
	tdsBRecordAddition     *RecordAddition
	fracSecondMul          int
	kRecordAdditions       []RecordAddition
	nRecordAdditions       []RecordAddition
}

type ParseOption func(*parser)

func WithAllowInvalidChars(allowInvalidChars bool) ParseOption {
	return func(p *parser) {
		p.allowInvalidChars = allowInvalidChars
	}
}

func newParser(options ...ParseOption) *parser {
	p := &parser{
		bRecordsAdditionsByTLC: make(map[string]*RecordAddition),
		latMinMul:              1,
		latMinDiv:              6e4,
		lonMinMul:              1,
		lonMinDiv:              6e4,
		fracSecondMul:          1e9,
	}
	for _, o := range options {
		o(p)
	}
	return p
}

func (p *parser) parse(r io.Reader) (*IGC, error) {
	var lines []string
	scanner := bufio.NewScanner(r)
	scanner.Split(scanLines)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return p.parseLines(lines)
}

func (p *parser) parseLines(lines []string) (*IGC, error) {
	records := make([]Record, 0, len(lines))
	var bRecords []*BRecord
	hRecordsByTLC := make(map[string]*HRecord)
	var kRecords []*KRecord
	var errs []error
	for i, lineStr := range lines {
		if len(lineStr) == 0 {
			continue
		}
		line := []byte(lineStr)

		var record Record
		var err error
		switch line[0] {
		case 'A':
			record, err = p.parseARecord(line)
		case 'B':
			record, err = p.parseBRecord(line)
		case 'C':
			record, err = p.parseCRecord(line)
		case 'D':
			record, err = p.parseDRecord(line)
		case 'E':
			record, err = p.parseERecord(line)
		case 'F':
			record, err = p.parseFRecord(line)
		case 'G':
			record, err = p.parseGRecord(line)
		case 'H':
			record, err = p.parseHRecord(line)
		case 'I':
			record, err = p.parseIRecord(line)
		case 'J':
			record, err = p.parseJRecord(line)
		case 'K':
			record, err = p.parseKRecord(line)
		case 'L':
			record, err = p.parseLRecord(line)
		case 'M':
			record, err = p.parseMRecord(line)
		case 'N':
			record, err = p.parseNRecord(line)
		default:
			err = unknownRecordTypeError(line[0])
		}
		if !p.allowInvalidChars {
			if match := invalidCharsRx.FindStringSubmatch(lineStr); match != nil {
				invalidChar := match[1][0]
				var invalidCharErr error
				if '\x20' <= invalidChar && invalidChar <= '\x7f' {
					invalidCharErr = fmt.Errorf("'%c': invalid character", invalidChar)
				} else {
					invalidCharErr = fmt.Errorf("'\\x%02x': invalid character", invalidChar)
				}
				if err == nil {
					err = invalidCharErr
				} else {
					err = errors.Join(err, invalidCharErr)
				}
			}
		}
		records = append(records, record)
		if err != nil {
			errs = append(errs, &Error{
				Line: i + 1,
				Err:  err,
			})
		}

		switch record := record.(type) {
		case *BRecord:
			if record != nil {
				bRecords = append(bRecords, record)
			}
		case *HRecord:
			if record != nil {
				hRecordsByTLC[record.TLC] = record
			}
		case *HFDTERecord:
			if record != nil {
				hRecordsByTLC[record.TLC] = &record.HRecord
				p.date = record.Date
			}
		case *IRecord:
			if record != nil {
				p.bRecordAdditions = append(p.bRecordAdditions, record.Additions...)
				for i, bRecordAddition := range record.Additions {
					p.bRecordsAdditionsByTLC[bRecordAddition.TLC] = &record.Additions[i]
				}
				if ladBRecordAddition, ok := p.bRecordsAdditionsByTLC["LAD"]; ok {
					p.ladBRecordAddition = ladBRecordAddition
					n := ladBRecordAddition.FinishColumn - ladBRecordAddition.StartColumn + 1
					p.latMinMul = intPow(10, n)
					p.latMinDiv = float64(6e4 * intPow(10, n))
				}
				if lodBRecordAddition, ok := p.bRecordsAdditionsByTLC["LOD"]; ok {
					p.lodBRecordAddition = lodBRecordAddition
					n := lodBRecordAddition.FinishColumn - lodBRecordAddition.StartColumn + 1
					p.lonMinMul = intPow(10, n)
					p.lonMinDiv = float64(6e4 * intPow(10, n))
				}
				if tdsBRecordAddition, ok := p.bRecordsAdditionsByTLC["TDS"]; ok {
					p.tdsBRecordAddition = tdsBRecordAddition
					n := tdsBRecordAddition.FinishColumn - tdsBRecordAddition.StartColumn + 1
					p.fracSecondMul = intPow(10, 9-n)
				}
			}
		case *JRecord:
			if record != nil {
				p.kRecordAdditions = record.Additions
			}
		case *KRecord:
			if record != nil {
				kRecords = append(kRecords, record)
			}
		case *MRecord:
			if record != nil {
				p.nRecordAdditions = record.Additions
			}
		}
	}

	return &IGC{
		Records:       records,
		Errs:          errs,
		HRecordsByTLC: hRecordsByTLC,
		BRecords:      bRecords,
		KRecords:      kRecords,
	}, nil
}

func (p *parser) parseARecord(line []byte) (*ARecord, error) {
	m := aRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('A')
	}
	var aRecord ARecord
	aRecord.ManufacturerID = string(m[1])
	if _, ok := ApprovedManufacturersByTLC[string(m[1])]; ok {
		uniqueFlightRecorderID, additionalData, _ := bytes.Cut(m[2], []byte("-"))
		aRecord.UniqueFlightRecorderID = string(uniqueFlightRecorderID)
		aRecord.AdditionalData = string(additionalData)
	} else {
		aRecord.UniqueFlightRecorderID = string(m[2])
	}
	return &aRecord, nil
}

func (p *parser) parseBRecord(line []byte) (*BRecord, error) {
	var errs []error
	m := bRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('B')
	}
	var bRecord BRecord
	nanosecond := 0
	if p.tdsBRecordAddition != nil {
		var fractionalSecond int
		var ok bool
		fractionalSecond, errs, ok = p.tdsBRecordAddition.intValue(line, errs)
		if ok {
			nanosecond = p.fracSecondMul * fractionalSecond
		}
	}
	bRecord.Time, errs = p.parseTime(m[1], m[2], m[3], nanosecond, errs)
	latDeg, _ := atoi(m[4])
	latMin, _ := atoi(m[5])
	if p.ladBRecordAddition != nil {
		var lad int
		var ok bool
		lad, errs, ok = p.ladBRecordAddition.intValue(line, errs)
		if ok {
			latMin = p.latMinMul*latMin + lad
		}
	}
	bRecord.Lat = float64(latDeg) + float64(latMin)/p.latMinDiv
	if m[6][0] == 'S' {
		bRecord.Lat = -bRecord.Lat
	}
	lonDeg, _ := atoi(m[7])
	lonMin, _ := atoi(m[8])
	if p.lodBRecordAddition != nil {
		var lod int
		var ok bool
		lod, errs, ok = p.lodBRecordAddition.intValue(line, errs)
		if ok {
			lonMin = p.lonMinMul*lonMin + lod
		}
	}
	bRecord.Lon = float64(lonDeg) + float64(lonMin)/p.lonMinDiv
	if m[9][0] == 'W' {
		bRecord.Lon = -bRecord.Lon
	}
	bRecord.Validity = Validity(m[10][0])
	altBarometric, _ := atoi(m[11])
	bRecord.AltBarometric = float64(altBarometric)
	altWGS84, _ := atoi(m[12])
	bRecord.AltWGS84 = float64(altWGS84)
	if len(p.bRecordAdditions) > 0 {
		additions := make(map[string]int, len(p.bRecordAdditions))
		for _, addition := range p.bRecordAdditions {
			var value int
			var ok bool
			value, errs, ok = addition.intValue(line, errs)
			if ok {
				additions[addition.TLC] = value
			}
		}
		bRecord.Additions = additions
	}
	return &bRecord, errors.Join(errs...)
}

func (p *parser) parseCRecord(line []byte) (Record, error) {
	if len(p.cRecords) == 0 {
		if cRecordDeclaration, err := p.parseCRecordDeclaration(line); err == nil {
			return cRecordDeclaration, nil
		}
	}
	m := cRecordWaypointRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('C')
	}
	var cRecordWaypoint CRecordWaypoint
	latDeg, _ := atoi(m[1])
	latMin, _ := atoi(m[2])
	cRecordWaypoint.Lat = float64(latDeg) + float64(latMin)/6e4
	if m[3][0] == 'S' {
		cRecordWaypoint.Lat = -cRecordWaypoint.Lat
	}
	lonDeg, _ := atoi(m[4])
	lonMin, _ := atoi(m[5])
	cRecordWaypoint.Lon = float64(lonDeg) + float64(lonMin)/6e4
	if m[6][0] == 'W' {
		cRecordWaypoint.Lon = -cRecordWaypoint.Lon
	}
	cRecordWaypoint.Text = string(m[7])
	return &cRecordWaypoint, nil
}

func (p *parser) parseCRecordDeclaration(line []byte) (*CRecordDeclaration, error) {
	m := cRecordDeclarationRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('C')
	}
	var cRecordDeclaration CRecordDeclaration
	declarationDay, _ := atoi(m[1])
	declarationMonth, _ := atoi(m[2])
	declarationTwoDigitYear, _ := atoi(m[3])
	declarationHour, _ := atoi(m[4])
	declarationMinute, _ := atoi(m[5])
	declarationSecond, _ := atoi(m[6])
	cRecordDeclaration.DeclarationTime = time.Date(
		makeYear(declarationTwoDigitYear), time.Month(declarationMonth), declarationDay,
		declarationHour, declarationMinute, declarationSecond, 0,
		time.UTC,
	)
	cRecordDeclaration.FlightDay, _ = atoi(m[7])
	cRecordDeclaration.FlightMonth, _ = atoi(m[8])
	cRecordDeclaration.FlightYear, _ = atoi(m[9])
	cRecordDeclaration.TaskNumber, _ = atoi(m[10])
	cRecordDeclaration.NumberOfTurnpoints, _ = atoi(m[11])
	cRecordDeclaration.Text = string(m[12])
	return &cRecordDeclaration, nil
}

func (p *parser) parseDRecord(line []byte) (*DRecord, error) {
	m := dRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('D')
	}
	var dRecord DRecord
	dRecord.GPSQualifier = GPSQualifier(m[1][0])
	dRecord.DGPSStationID, _ = atoi(m[2])
	return &dRecord, nil
}

func (p *parser) parseERecord(line []byte) (Record, error) {
	var errs []error
	m := eRecordRx.FindSubmatch(line)
	if m == nil {
		if m := eRecordWithoutTLCRx.FindSubmatch(line); m != nil {
			var invalidERecord ERecordWithoutTLC
			invalidERecord.Time, errs = p.parseTime(m[1], m[2], m[3], 0, errs)
			invalidERecord.Text = string(m[4])
			return &invalidERecord, errors.Join(errs...)
		}
		return nil, invalidRecordError('E')
	}
	var eRecord ERecord
	eRecord.Time, errs = p.parseTime(m[1], m[2], m[3], 0, errs)
	eRecord.TLC = string(m[4])
	eRecord.Text = string(m[5])
	return &eRecord, errors.Join(errs...)
}

func (p *parser) parseFRecord(line []byte) (*FRecord, error) {
	m := fRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('F')
	}
	var fRecord FRecord
	var errs []error
	fRecord.Time, errs = p.parseTime(m[1], m[2], m[3], 0, errs)
	n := len(m[4]) / 2
	satelliteIDs := make([]int, 0, n)
	for i := range n {
		satelliteID, _ := atoi(m[4][2*i : 2*i+2])
		satelliteIDs = append(satelliteIDs, satelliteID)
	}
	fRecord.SatelliteIDs = satelliteIDs
	return &fRecord, errors.Join(errs...)
}

func (p *parser) parseGRecord(line []byte) (*GRecord, error) {
	m := gRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('G')
	}
	var gRecord GRecord
	gRecord.Text = string(m[1])
	return &gRecord, nil
}

func (p *parser) parseHRecord(line []byte) (Record, error) {
	if m := hfdteRecordRx.FindSubmatch(line); m != nil {
		var hfdteRecord HFDTERecord
		hfdteRecord.HRecord.Source = 'F'
		hfdteRecord.HRecord.TLC = "DTE"
		hfdteRecord.HRecord.Value = string(bytes.Join(m[1:], nil))
		day, _ := atoi(m[1])
		month, _ := atoi(m[2])
		twoDigitYear, _ := atoi(m[3])
		year := makeYear(twoDigitYear)
		hfdteRecord.Date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		return &hfdteRecord, nil
	}
	if m := hffxaRecordRx.FindSubmatch(line); m != nil {
		var hRecord HRecord
		hRecord.Source = 'F'
		hRecord.TLC = "FXA"
		hRecord.Value = string(m[1])
		return &hRecord, nil
	}
	m := hRecordRx.FindSubmatch(line)
	if m == nil {
		if m := hRecordWithInvalidSourceRx.FindSubmatch(line); m != nil {
			var hRecordWithInvalidSource HRecordWithInvalidSource
			hRecordWithInvalidSource.Source = string(m[1][0])
			hRecordWithInvalidSource.TLC = string(m[2])
			hRecordWithInvalidSource.LongName = string(m[3])
			hRecordWithInvalidSource.Value = string(m[4])
			return &hRecordWithInvalidSource, nil
		}
		return nil, invalidRecordError('H')
	}
	var hRecord HRecord
	hRecord.Source = Source(m[1][0])
	hRecord.TLC = string(m[2])
	hRecord.LongName = string(m[3])
	hRecord.Value = string(m[4])
	if hRecord.TLC == "DTE" {
		m := hfdteRecordValueRx.FindSubmatch(m[4])
		if m == nil {
			return &hRecord, invalidRecordError('H')
		}
		var hfdteRecord HFDTERecord
		hfdteRecord.HRecord = hRecord
		day, _ := atoi(m[1])
		month, _ := atoi(m[2])
		twoDigitYear, _ := atoi(m[3])
		year := makeYear(twoDigitYear)
		hfdteRecord.Date = time.Date(year, time.Month(month), day, 0, 0, 0, 0, time.UTC)
		if len(m[4]) != 0 {
			hfdteRecord.FlightNumber, _ = atoi(m[4])
		}
		return &hfdteRecord, nil
	}
	return &hRecord, nil
}

func (p *parser) parseIRecord(line []byte) (*IRecord, error) {
	additions, err := p.parseRecordAdditions(line, 36)
	if err != nil {
		return nil, err
	}
	return &IRecord{
		Additions: additions,
	}, nil
}

func (p *parser) parseJRecord(line []byte) (*JRecord, error) {
	additions, err := p.parseRecordAdditions(line, 8)
	if err != nil {
		return nil, err
	}
	return &JRecord{
		Additions: additions,
	}, nil
}

func (p *parser) parseRecordAdditions(line []byte, startColumn int) ([]RecordAddition, error) {
	m := ijmRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError(line[0])
	}
	n, _ := atoi(m[1])
	if len(m[2]) != 7*n {
		return nil, invalidRecordError(line[0])
	}
	var errs []error
	additions := make([]RecordAddition, 0, n)
	for i := range n {
		var addition RecordAddition
		addition.StartColumn, _ = atoi(m[2][7*i : 7*i+2])
		addition.FinishColumn, _ = atoi(m[2][7*i+2 : 7*i+4])
		addition.TLC = string(m[2][7*i+4 : 7*i+7])
		var message string
		switch {
		case addition.StartColumn != startColumn:
			message = "invalid start column"
		case addition.FinishColumn < addition.StartColumn:
			message = "invalid finish column"
		}
		if message == "" {
			additions = append(additions, addition)
			startColumn = addition.FinishColumn + 1
		} else {
			err := &invalidAdditionError{
				addition: addition,
				message:  message,
			}
			errs = append(errs, err)
		}
	}
	return additions, errors.Join(errs...)
}

func (p *parser) parseKRecord(line []byte) (*KRecord, error) {
	m := knRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('K')
	}
	var kRecord KRecord
	var errs []error
	kRecord.Time, errs = p.parseTime(m[1], m[2], m[3], 0, errs)
	if len(p.kRecordAdditions) > 0 {
		kRecord.Additions = make(map[string]int, len(p.kRecordAdditions))
		for _, addition := range p.kRecordAdditions {
			var value int
			var ok bool
			value, errs, ok = addition.intValue(line, errs)
			if ok {
				kRecord.Additions[addition.TLC] = value
			}
		}
	}
	return &kRecord, errors.Join(errs...)
}

func (p *parser) parseLRecord(line []byte) (Record, error) {
	m := lRecordRx.FindSubmatch(line)
	if m == nil {
		if m := lRecordWithoutTLCRx.FindSubmatch(line); m != nil {
			var lRecord LRecordWithoutTLC
			lRecord.Text = string(m[1])
			return &lRecord, nil
		}
		return nil, invalidRecordError('L')
	}
	var lRecord LRecord
	lRecord.Input = string(m[1])
	lRecord.Text = string(m[2])
	return &lRecord, nil
}

func (p *parser) parseMRecord(line []byte) (*MRecord, error) {
	additions, err := p.parseRecordAdditions(line, 8)
	if err != nil {
		return nil, err
	}
	return &MRecord{
		Additions: additions,
	}, nil
}

func (p *parser) parseNRecord(line []byte) (*NRecord, error) {
	m := knRecordRx.FindSubmatch(line)
	if m == nil {
		return nil, invalidRecordError('N')
	}
	var nRecord NRecord
	var errs []error
	nRecord.Time, errs = p.parseTime(m[1], m[2], m[3], 0, errs)
	if len(p.nRecordAdditions) > 0 {
		nRecord.Additions = make(map[string]int, len(p.nRecordAdditions))
		for _, addition := range p.nRecordAdditions {
			var value int
			var ok bool
			value, errs, ok = addition.intValue(line, errs)
			if ok {
				nRecord.Additions[addition.TLC] = value
			}
		}
	}
	return &nRecord, errors.Join(errs...)
}

func (p *parser) parseTime(hourData, minuteData, secondData []byte, nanosecond int, errs []error) (time.Time, []error) {
	if p.date.IsZero() {
		return time.Time{}, append(errs, errNoDate)
	}
	hour, _ := atoi(hourData)
	minute, _ := atoi(minuteData)
	second, _ := atoi(secondData)
	durationSinceMidnight := time.Duration(hour)*time.Hour +
		time.Duration(minute)*time.Minute +
		time.Duration(second)*time.Second +
		time.Duration(nanosecond)*time.Nanosecond
	for {
		t := p.date.Add(durationSinceMidnight)
		if !t.Before(p.prevTime) {
			p.prevTime = t
			return t, errs
		}
		p.date = p.date.AddDate(0, 0, 1)
	}
}

func (a *RecordAddition) bytesValue(line []byte, errs []error) ([]byte, []error, bool) {
	if len(line) < a.FinishColumn {
		return nil, append(errs, &missingAdditionError{
			addition: *a,
		}), false
	}
	return line[a.StartColumn-1 : a.FinishColumn], errs, true
}

func (a *RecordAddition) intValue(line []byte, errs []error) (int, []error, bool) {
	data, errs, ok := a.bytesValue(line, errs)
	if !ok {
		return 0, errs, false
	}
	result, err := atoi(data)
	if err != nil {
		return 0, append(errs, err), false
	}
	return result, errs, true
}

func atoi(data []byte) (int, error) {
	digits := data
	sign := 1
	if len(data) > 0 && data[0] == '-' {
		sign = -sign
		digits = digits[1:]
	}
	if len(digits) == 0 {
		return 0, &atoiSyntaxError{
			num: string(data),
		}
	}
	result := 0
	for _, b := range digits {
		if b < '0' || '9' < b {
			return 0, &atoiSyntaxError{
				num: string(data),
			}
		}
		result = 10*result + int(b) - '0'
	}
	return result * sign, nil
}

// trimCRs drops terminal \rs from data.
func trimCRs(data []byte) []byte {
	for i := len(data) - 1; i >= 0; i-- {
		if data[i] != '\r' {
			return data[:i+1]
		}
	}
	return nil
}

// scanLines is a bufio.SplitFunc that splits lines according the regexp \r*\n.
func scanLines(data []byte, atEOF bool) (int, []byte, error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.IndexByte(data, '\n'); i >= 0 {
		return i + 1, trimCRs(data[:i]), nil
	}
	if atEOF {
		return len(data), trimCRs(data), nil
	}
	return 0, nil, nil
}

// intPow returns x raised to the power of y.
func intPow(x, y int) int {
	result := 1
	for ; y != 0; y >>= 1 {
		if y&1 == 1 {
			result *= x
		}
		x *= x
	}
	return result
}

// makeYear converts a four-digit year to a two-digit year.
func makeYear(twoDigitYear int) int {
	// The initial IGC standard was developed in 1993. See
	// https://www.fai.org/sites/default/files/igc-approval_table_history_-_2021-8-22.pdf.
	if twoDigitYear >= 93 {
		return 1900 + twoDigitYear
	}
	return 2000 + twoDigitYear
}
