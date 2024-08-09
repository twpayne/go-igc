package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/twpayne/go-igc"
)

type Range[T any] struct {
	Min T
	Max T
}

type BSummary struct {
	Duration      friendlyDuration
	Time          Range[time.Time]
	TimeDeltas    map[int]int
	Lat           Range[float64]
	Lon           Range[float64]
	AltWGS84      Range[float64]
	AltBarometric Range[float64]
	Additions     map[string]*Range[int] `json:",omitempty"`
}

type KSummary struct {
	TimeDeltas map[int]int
	Additions  map[string]*Range[int] `json:",omitempty"`
}

type Summary struct {
	Filename      string
	Size          int64
	BRecordFreq   float64
	KRecordFreq   float64
	Records       int
	RecordCounts  map[string]int
	HRecordsByTLC map[string]string
	B             *BSummary `json:",omitempty"`
	K             *KSummary `json:",omitempty"`
}

type friendlyDuration time.Duration

func (d friendlyDuration) MarshalJSON() ([]byte, error) {
	seconds := int((time.Duration(d) + time.Second/2) / time.Second)
	return []byte(fmt.Sprintf(`"%d:%02d:%02d"`, seconds/3600, seconds/60%60, seconds%60)), nil
}

func summarizeFile(filename string) (*Summary, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return nil, err
	}

	igc, err := igc.Parse(file)
	if err != nil {
		return nil, err
	}

	duration := igc.BRecords[len(igc.BRecords)-1].Time.Sub(igc.BRecords[0].Time)

	recordCounts := make(map[string]int)
	for _, record := range igc.Records {
		recordCounts[string(record.Type())]++
	}

	hRecordsByTLC := make(map[string]string, len(igc.HRecordsByTLC))
	for tlc, hRecord := range igc.HRecordsByTLC {
		hRecordsByTLC[tlc] = hRecord.Value
	}

	var bSummary *BSummary
	if len(igc.BRecords) > 0 {
		bRecordTimeDeltas := make(map[int]int)
		latRange := Range[float64]{Min: math.Inf(1), Max: math.Inf(-1)}
		lonRange := Range[float64]{Min: math.Inf(1), Max: math.Inf(-1)}
		altWGS84Range := Range[float64]{Min: math.Inf(1), Max: math.Inf(-1)}
		altBarometricRange := Range[float64]{Min: math.Inf(1), Max: math.Inf(-1)}
		bAdditionRanges := make(map[string]*Range[int], len(igc.BRecords[0].Additions))
		for i, bRecord := range igc.BRecords {
			if i != 0 {
				bRecordTimeDeltas[int(bRecord.Time.Sub(igc.BRecords[i-1].Time)/time.Second)]++
			}
			latRange.Min = min(latRange.Min, bRecord.Lat)
			latRange.Max = max(latRange.Max, bRecord.Lat)
			lonRange.Min = min(lonRange.Min, bRecord.Lon)
			lonRange.Max = max(lonRange.Max, bRecord.Lon)
			altWGS84Range.Min = min(altWGS84Range.Min, bRecord.AltWGS84)
			altWGS84Range.Max = max(altWGS84Range.Max, bRecord.AltWGS84)
			altBarometricRange.Min = min(altBarometricRange.Min, bRecord.AltBarometric)
			altBarometricRange.Max = max(altBarometricRange.Max, bRecord.AltBarometric)
			for additionKey, additionValue := range bRecord.Additions {
				if additionRange, ok := bAdditionRanges[additionKey]; ok {
					additionRange.Min = min(additionRange.Min, additionValue)
					additionRange.Max = max(additionRange.Max, additionValue)
				} else {
					bAdditionRanges[additionKey] = &Range[int]{Min: additionValue, Max: additionValue}
				}
			}
		}
		bSummary = &BSummary{
			Duration: friendlyDuration(duration),
			Time: Range[time.Time]{
				Min: igc.BRecords[0].Time,
				Max: igc.BRecords[len(igc.BRecords)-1].Time,
			},
			TimeDeltas:    bRecordTimeDeltas,
			Lon:           lonRange,
			Lat:           latRange,
			AltWGS84:      altWGS84Range,
			AltBarometric: altBarometricRange,
			Additions:     bAdditionRanges,
		}
	}

	var kSummary *KSummary
	if len(igc.KRecords) > 0 {
		kRecordTimeDeltas := make(map[int]int)
		kAdditionRanges := make(map[string]*Range[int], len(igc.BRecords[0].Additions))
		for i, kRecord := range igc.KRecords {
			if i != 0 {
				kRecordTimeDeltas[int(kRecord.Time.Sub(igc.KRecords[i-1].Time)/time.Second)]++
			}
			for additionKey, additionValue := range kRecord.Additions {
				if additionRange, ok := kAdditionRanges[additionKey]; ok {
					additionRange.Min = min(additionRange.Min, additionValue)
					additionRange.Max = max(additionRange.Max, additionValue)
				} else {
					kAdditionRanges[additionKey] = &Range[int]{Min: additionValue, Max: additionValue}
				}
			}
		}
		kSummary = &KSummary{
			TimeDeltas: kRecordTimeDeltas,
			Additions:  kAdditionRanges,
		}
	}

	return &Summary{
		Filename:      filename,
		Size:          fileInfo.Size(),
		BRecordFreq:   float64(len(igc.BRecords)-1) * float64(time.Second) / float64(duration),
		KRecordFreq:   float64(len(igc.KRecords)-1) * float64(time.Second) / float64(duration),
		Records:       len(igc.Records),
		RecordCounts:  recordCounts,
		HRecordsByTLC: hRecordsByTLC,
		B:             bSummary,
		K:             kSummary,
	}, nil
}

func run() error {
	flag.Parse()

	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	for _, arg := range flag.Args() {
		summary, err := summarizeFile(arg)
		if err != nil {
			return err
		}
		if err := encoder.Encode(summary); err != nil {
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
