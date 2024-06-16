package igc_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/alecthomas/assert/v2"

	"github.com/twpayne/go-igc"
)

func TestParseLine(t *testing.T) {
	for _, tc := range []struct {
		name           string
		line           string
		expectedRecord igc.Record
		expectedErr    string
	}{
		{
			name: "a_record_approved_manufacturer",
			line: "AFLY05094-extra",
			expectedRecord: &igc.ARecord{
				ManufacturerID:         "FLY",
				UniqueFlightRecorderID: "05094",
				AdditionalData:         "extra",
			},
		},
		{
			name: "a_record_approved_manufacturer_with_additional_data",
			line: "AFLY05094-extra",
			expectedRecord: &igc.ARecord{
				ManufacturerID:         "FLY",
				UniqueFlightRecorderID: "05094",
				AdditionalData:         "extra",
			},
		},
		{
			name: "a_record_unapproved_manufacturer",
			line: "AXYZa-b-c",
			expectedRecord: &igc.ARecord{
				ManufacturerID:         "XYZ",
				UniqueFlightRecorderID: "a-b-c",
			},
		},
		{
			name:        "a_record_invalid",
			line:        "A",
			expectedErr: "1: invalid A record",
		},
		{
			name:        "b_record_invalid",
			line:        "B",
			expectedErr: "1: invalid B record",
		},
		{
			name:        "b_record_no_date",
			line:        "B1005364607690N00610358EA0000001265",
			expectedErr: "1: no date",
		},
		{
			name: "first_c_record",
			line: "C110524093545000000000502",
			expectedRecord: &igc.FirstCRecord{
				DeclarationTime:    time.Date(2024, time.May, 11, 9, 35, 45, 0, time.UTC),
				TaskNumber:         5,
				NumberOfTurnpoints: 2,
			},
		},
		{
			name: "first_c_record_xctrack",
			line: "C0110231200370000000000-1 Competition task",
			expectedRecord: &igc.FirstCRecord{
				DeclarationTime:    time.Date(2023, time.October, 1, 12, 0, 37, 0, time.UTC),
				NumberOfTurnpoints: -1,
				Text:               " Competition task",
			},
		},
		{
			name: "c_record",
			line: "C4415173N00604205ET-296-Trainon",
			expectedRecord: &igc.CRecord{
				Lat:  44 + 15173/6e4,
				Lon:  6 + 4205/6e4,
				Text: "T-296-Trainon",
			},
		},
		{
			name: "c_record_sw",
			line: "C4415173S00604205WT-296-Trainon",
			expectedRecord: &igc.CRecord{
				Lat:  -(44 + 15173/6e4),
				Lon:  -(6 + 4205/6e4),
				Text: "T-296-Trainon",
			},
		},
		{
			name:        "c_record_invalid",
			line:        "C",
			expectedErr: "1: invalid C record",
		},
		{
			name: "d_record",
			line: "D21234",
			expectedRecord: &igc.DRecord{
				GPSQualifier:  igc.GPSQualifierDGPS,
				DGPSStationID: 1234,
			},
		},
		{
			name:        "d_record_invalid",
			line:        "D",
			expectedErr: "1: invalid D record",
		},
		{
			name:        "e_record_invalid",
			line:        "E",
			expectedErr: "1: invalid E record",
		},
		{
			name:        "e_record_no_date",
			line:        "E100153PEV",
			expectedErr: "1: no date",
		},
		{
			name:        "f_record_invalid",
			line:        "F",
			expectedErr: "1: invalid F record",
		},
		{
			name:        "f_record_no_date",
			line:        "F1503140228322203081917043121",
			expectedErr: "1: no date",
		},
		{
			name: "g_record",
			line: "G7EEED180BADCA6F828AE4B69B0891F91",
			expectedRecord: &igc.GRecord{
				Text: "7EEED180BADCA6F828AE4B69B0891F91",
			},
		},
		{
			name: "hfdte_record",
			line: "HFDTEDATE:040624,01",
			expectedRecord: &igc.HFDTERecord{
				HRecord: igc.HRecord{
					Source:   igc.SourceFlightRecorder,
					TLC:      "DTE",
					LongName: "DATE",
					Value:    "040624,01",
				},
				Date:         time.Date(2024, time.June, 4, 0, 0, 0, 0, time.UTC),
				FlightNumber: 1,
			},
		},
		{
			name: "hfdte_record_no_flight_number",
			line: "HFDTEDATE:100923",
			expectedRecord: &igc.HFDTERecord{
				HRecord: igc.HRecord{
					Source:   igc.SourceFlightRecorder,
					TLC:      "DTE",
					LongName: "DATE",
					Value:    "100923",
				},
				Date: time.Date(2023, time.September, 10, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "hfdte_record_old",
			line: "HFDTE220495",
			expectedRecord: &igc.HFDTERecord{
				HRecord: igc.HRecord{
					Source: igc.SourceFlightRecorder,
					TLC:    "DTE",
					Value:  "220495",
				},
				Date: time.Date(1995, time.April, 22, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "hffxa_record_old",
			line: "HFFXA100",
			expectedRecord: &igc.HRecord{
				Source: igc.SourceFlightRecorder,
				TLC:    "FXA",
				Value:  "100",
			},
		},
		{
			name: "hfplt_record",
			line: "HFPLTPILOTINCHARGE:Tom Payne",
			expectedRecord: &igc.HRecord{
				Source:   igc.SourceFlightRecorder,
				TLC:      "PLT",
				LongName: "PILOTINCHARGE",
				Value:    "Tom Payne",
			},
		},
		{
			name: "hffrs_record",
			line: "HFFRSSECURITYOK",
			expectedRecord: &igc.HRecord{
				Source:   igc.SourceFlightRecorder,
				TLC:      "FRS",
				LongName: "SECURITYOK",
			},
		},
		{
			name:        "h_record_invalid",
			line:        "H",
			expectedErr: "1: invalid H record",
		},
		{
			name: "h_record_invalid_source",
			line: "HSCCLCOMPETITION CLASS:FAI-3 (PG)",
			expectedRecord: &igc.HRecordWithInvalidSource{
				Source:   "S",
				TLC:      "CCL",
				LongName: "COMPETITION CLASS",
				Value:    "FAI-3 (PG)",
			},
		},
		{
			name:        "hfdte_record_invalid",
			line:        "HFDTE",
			expectedErr: "1: invalid H record",
		},
		{
			name:           "i_record_empty",
			line:           "I00",
			expectedRecord: &igc.IRecord{},
		},
		{
			name:           "i_record_empty",
			line:           "I00",
			expectedRecord: &igc.IRecord{},
		},
		{
			name: "i_record_simple",
			line: "I013638TAS",
			expectedRecord: &igc.IRecord{
				Additions: []igc.BKRecordAddition{
					{StartColumn: 36, FinishColumn: 38, TLC: "TAS"},
				},
			},
		},
		{
			name: "i_record_two_additions",
			line: "I023638FXA3940SIU",
			expectedRecord: &igc.IRecord{
				Additions: []igc.BKRecordAddition{
					{StartColumn: 36, FinishColumn: 38, TLC: "FXA"},
					{StartColumn: 39, FinishColumn: 40, TLC: "SIU"},
				},
			},
		},
		{
			name:        "i_record_invalid",
			line:        "I",
			expectedErr: "1: invalid I record",
		},
		{
			name:        "i_record_incomplete",
			line:        "I01",
			expectedErr: "1: invalid I record",
		},
		{
			name:        "i_record_invalid_start",
			line:        "I013535AAA",
			expectedErr: "1: AAA: invalid start column",
		},
		{
			name:        "i_record_invalid_finish",
			line:        "I023636AAA3736BBB",
			expectedErr: "1: BBB: invalid finish column",
		},
		{
			name: "j_record_two_additions",
			line: "J020810WDI1113WSP",
			expectedRecord: &igc.JRecord{
				Additions: []igc.BKRecordAddition{
					{StartColumn: 8, FinishColumn: 10, TLC: "WDI"},
					{StartColumn: 11, FinishColumn: 13, TLC: "WSP"},
				},
			},
		},
		{
			name:        "j_record_invalid",
			line:        "J",
			expectedErr: "1: invalid J record",
		},
		{
			name:        "k_record_incomplete",
			line:        "K01",
			expectedErr: "1: invalid K record",
		},
		{
			name:        "k_record_no_date",
			line:        "K150452276600",
			expectedErr: "1: no date",
		},
		{
			name: "l_record",
			line: "LXNA::VEHICLE:1",
			expectedRecord: &igc.LRecord{
				Input: "XNA",
				Text:  "::VEHICLE:1",
			},
		},
		{
			name: "l_record_short",
			line: "LCU::HPGTYGLIDERTYPE:SZD 55",
			expectedRecord: &igc.LRecordWithoutTLC{
				Text: "CU::HPGTYGLIDERTYPE:SZD 55",
			},
		},
		{
			name:           "l_record_empty",
			line:           "L",
			expectedRecord: &igc.LRecordWithoutTLC{},
		},
		{
			name:        "unknown_record",
			line:        "X",
			expectedErr: "1: X: unknown record type",
		},
		{
			name:        "unknown_record_null",
			line:        "\x00",
			expectedErr: `1: "\x00": unknown record type`,
		},
		{
			name: "empty line",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := igc.ParseLines([]string{tc.line})
			assert.NoError(t, err)
			if tc.expectedErr != "" {
				assertEqualErrors(t, []string{tc.expectedErr}, actual.Errs)
			}
			if tc.expectedRecord != nil {
				assert.Equal(t, []igc.Record{tc.expectedRecord}, actual.Records)
			}
		})
	}
}

func TestParseLines(t *testing.T) {
	for _, tc := range []struct {
		name            string
		lines           []string
		expectedRecords []igc.Record
		expectedErrs    []string
	}{
		{
			name: "b_record",
			lines: []string{
				"HFDTE020508",
				"B1005364607690N00610358EA0000001265",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source: igc.SourceFlightRecorder,
						TLC:    "DTE",
						Value:  "020508",
					},
					Date: time.Date(2008, time.May, 2, 0, 0, 0, 0, time.UTC),
				},
				&igc.BRecord{
					Time:     time.Date(2008, time.May, 2, 10, 5, 36, 0, time.UTC),
					Lat:      46 + 7690/6e4,
					Lon:      6 + 10358/6e4,
					Validity: igc.Validity3D,
					AltWGS84: 1265,
				},
			},
		},
		{
			name: "b_record_negative_altitudes",
			lines: []string{
				"HFDTEDATE:211022",
				"B1652002737662N08031679WA-0037-0017",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "211022",
					},
					Date: time.Date(2022, time.October, 21, 0, 0, 0, 0, time.UTC),
				},
				&igc.BRecord{
					Time:          time.Date(2022, time.October, 21, 16, 52, 0, 0, time.UTC),
					Lat:           27 + 37662/6e4,
					Lon:           -(80 + 31679/6e4),
					Validity:      igc.Validity3D,
					AltBarometric: -37,
					AltWGS84:      -17,
				},
			},
		},
		{
			name: "b_record_with_additions",
			lines: []string{
				"HFDTEDATE:040624,01",
				"I023638FXA3940SIU",
				"B1501444708879N00832146EA009290094100612",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "040624,01",
					},
					Date:         time.Date(2024, time.June, 4, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 38, TLC: "FXA"},
						{StartColumn: 39, FinishColumn: 40, TLC: "SIU"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 4, 15, 1, 44, 0, time.UTC),
					Lat:           47 + 8879/6e4,
					Lon:           8 + 32146/6e4,
					Validity:      igc.Validity3D,
					AltBarometric: 929,
					AltWGS84:      941,
					Additions: map[string]int{
						"FXA": 6,
						"SIU": 12,
					},
				},
			},
		},
		{
			name: "b_record_missing_addition",
			lines: []string{
				"HFDTEDATE:040624,01",
				"I023638FXA3940SIU",
				"B1501444708879N00832146EA0092900941006",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "040624,01",
					},
					Date:         time.Date(2024, time.June, 4, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 38, TLC: "FXA"},
						{StartColumn: 39, FinishColumn: 40, TLC: "SIU"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 4, 15, 1, 44, 0, time.UTC),
					Lat:           47 + 8879/6e4,
					Lon:           8 + 32146/6e4,
					Validity:      igc.Validity3D,
					AltBarometric: 929,
					AltWGS84:      941,
					Additions: map[string]int{
						"FXA": 6,
					},
				},
			},
			expectedErrs: []string{
				"3: missing SIU addition",
			},
		},
		{
			name: "b_record_invalid_additions",
			lines: []string{
				"HFDTEDATE:040624,01",
				"I013638FXA",
				"B1501444708879N00832146EA0092900941-1X",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "040624,01",
					},
					Date:         time.Date(2024, time.June, 4, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 38, TLC: "FXA"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 4, 15, 1, 44, 0, time.UTC),
					Lat:           47 + 8879/6e4,
					Lon:           8 + 32146/6e4,
					Validity:      igc.Validity3D,
					AltBarometric: 929,
					AltWGS84:      941,
					Additions:     map[string]int{},
				},
			},
			expectedErrs: []string{
				`3: "-1X": syntax error`,
			},
		},
		{
			name: "b_record_tds_lad_lod",
			lines: []string{
				"HFDTEDATE:040624,01",
				"I033636TDS3737LAD3838LOD",
				"B1501444708879N00832146EA0092900941456",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "040624,01",
					},
					Date:         time.Date(2024, time.June, 4, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 36, TLC: "TDS"},
						{StartColumn: 37, FinishColumn: 37, TLC: "LAD"},
						{StartColumn: 38, FinishColumn: 38, TLC: "LOD"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 4, 15, 1, 44, 4e8, time.UTC),
					Lat:           47 + 88795/6e5,
					Lon:           8 + 321466/6e5,
					Validity:      igc.Validity3D,
					AltBarometric: 929,
					AltWGS84:      941,
					Additions: map[string]int{
						"LAD": 5,
						"LOD": 6,
						"TDS": 4,
					},
				},
			},
		},
		{
			name: "e_record",
			lines: []string{
				"HFDTE110524",
				"E100153PEV",
				"E164225BFION AH",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source: igc.SourceFlightRecorder,
						TLC:    "DTE",
						Value:  "110524",
					},
					Date: time.Date(2024, time.May, 11, 0, 0, 0, 0, time.UTC),
				},
				&igc.ERecord{
					Time: time.Date(2024, time.May, 11, 10, 1, 53, 0, time.UTC),
					TLC:  "PEV",
				},
				&igc.ERecord{
					Time: time.Date(2024, time.May, 11, 16, 42, 25, 0, time.UTC),
					TLC:  "BFI",
					Text: "ON AH",
				},
			},
		},
		{
			name: "f_record",
			lines: []string{
				"HFDTEDATE:040624,01",
				"F1503140228322203081917043121",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "040624,01",
					},
					Date:         time.Date(2024, time.June, 4, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.FRecord{
					Time:         time.Date(2024, time.June, 4, 15, 3, 14, 0, time.UTC),
					SatelliteIDs: []int{2, 28, 32, 22, 3, 8, 19, 17, 4, 31, 21},
				},
			},
		},
		{
			name: "e_record_flyskyhy",
			lines: []string{
				"HFDTE090224",
				"E030428Waypoint Waypoint reached",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source: igc.SourceFlightRecorder,
						TLC:    "DTE",
						Value:  "090224",
					},
					Date: time.Date(2024, time.February, 9, 0, 0, 0, 0, time.UTC),
				},
				&igc.ERecordWithoutTLC{
					Time: time.Date(2024, time.February, 9, 3, 4, 28, 0, time.UTC),
					Text: "Waypoint Waypoint reached", //nolint:dupword
				},
			},
		},
		{
			name: "k_record",
			lines: []string{
				"HFDTEDATE:040624,01",
				"J020810WDI1113WSP",
				"K150452276600",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "040624,01",
					},
					Date:         time.Date(2024, time.June, 4, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.JRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 8, FinishColumn: 10, TLC: "WDI"},
						{StartColumn: 11, FinishColumn: 13, TLC: "WSP"},
					},
				},
				&igc.KRecord{
					Time: time.Date(2024, time.June, 4, 15, 4, 52, 0, time.UTC),
					Additions: map[string]int{
						"WDI": 276,
						"WSP": 600,
					},
				},
			},
		},
		{
			name: "xctracer",
			lines: []string{
				"AXTR20C38FF2C110",
				"HFDTE151115",
				"B1316284654230N00839079EA0147801630",
			},
			expectedRecords: []igc.Record{
				&igc.ARecord{
					ManufacturerID:         "XTR",
					UniqueFlightRecorderID: "20C38FF2C110",
				},
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source: igc.SourceFlightRecorder,
						TLC:    "DTE",
						Value:  "151115",
					},
					Date: time.Date(2015, time.November, 15, 0, 0, 0, 0, time.UTC),
				},
				&igc.BRecord{
					Time:          time.Date(2015, time.November, 15, 13, 16, 28, 0, time.UTC),
					Lat:           46 + 54230/6e4,
					Lon:           8 + 39079/6e4,
					Validity:      igc.Validity3D,
					AltBarometric: 1478,
					AltWGS84:      1630,
				},
			},
		},
		{
			name: "cpilot",
			lines: []string{
				"ACPP274CPILOT - s/n:11002274",
				"HFDTE020613",
				"I033638FXA3940SIU4141TDS",
				"B1053525151892N00203986WA0017900275000108",
			},
			expectedRecords: []igc.Record{
				&igc.ARecord{
					ManufacturerID:         "CPP",
					UniqueFlightRecorderID: "274CPILOT - s/n:11002274",
				},
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source: igc.SourceFlightRecorder,
						TLC:    "DTE",
						Value:  "020613",
					},
					Date: time.Date(2013, time.June, 2, 0, 0, 0, 0, time.UTC),
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 38, TLC: "FXA"},
						{StartColumn: 39, FinishColumn: 40, TLC: "SIU"},
						{StartColumn: 41, FinishColumn: 41, TLC: "TDS"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2013, time.June, 2, 10, 53, 52, 8e8, time.UTC),
					Lat:           51 + 51892/6e4,
					Lon:           -(2 + 3986/6e4),
					Validity:      igc.Validity3D,
					AltBarometric: 179,
					AltWGS84:      275,
					Additions: map[string]int{
						"FXA": 0,
						"SIU": 10,
						"TDS": 8,
					},
				},
			},
		},
		{
			name: "compcheck",
			lines: []string{
				"AXCC64BCompCheck-3.2",
				"HFDTE100810",
				"I033637LAD3839LOD4040TDS",
				"B1146174031985N00726775WA010040114912340",
			},
			expectedRecords: []igc.Record{
				&igc.ARecord{
					ManufacturerID:         "XCC",
					UniqueFlightRecorderID: "64BCompCheck-3.2",
				},
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source: igc.SourceFlightRecorder,
						TLC:    "DTE",
						Value:  "100810",
					},
					Date: time.Date(2010, time.August, 10, 0, 0, 0, 0, time.UTC),
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 37, TLC: "LAD"},
						{StartColumn: 38, FinishColumn: 39, TLC: "LOD"},
						{StartColumn: 40, FinishColumn: 40, TLC: "TDS"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2010, time.August, 10, 11, 46, 17, 0, time.UTC),
					Lat:           40 + 3198512/6e6,
					Lon:           -(7 + 2677534/6e6),
					Validity:      igc.Validity3D,
					AltBarometric: 1004,
					AltWGS84:      1149,
					Additions: map[string]int{
						"LAD": 12,
						"LOD": 34,
						"TDS": 0,
					},
				},
			},
		},
		{
			name: "flymaster",
			lines: []string{
				"AXGD Flymaster LiveSD  SN03142  SW1.07b",
				"HFDTEDATE:220418,01",
				"B1316284654230N00839079EA0147801630",
			},
			expectedRecords: []igc.Record{
				&igc.ARecord{
					ManufacturerID:         "XGD",
					UniqueFlightRecorderID: " Flymaster LiveSD  SN03142  SW1.07b",
				},
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "220418,01",
					},
					Date:         time.Date(2018, time.April, 22, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.BRecord{
					Time:          time.Date(2018, time.April, 22, 13, 16, 28, 0, time.UTC),
					Lat:           46 + 54230/6e4,
					Lon:           8 + 39079/6e4,
					Validity:      igc.Validity3D,
					AltBarometric: 1478,
					AltWGS84:      1630,
				},
			},
		},
		{
			name: "utc_midnight_rollover",
			lines: []string{
				"HFDTEDATE:050624,01",
				"I023636LAD3737LOD",
				"B2359593716084N11815009WA044060467236",
				"B0000013716074N11815007WA044080467530",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "050624,01",
					},
					Date:         time.Date(2024, time.June, 5, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 36, TLC: "LAD"},
						{StartColumn: 37, FinishColumn: 37, TLC: "LOD"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 5, 23, 59, 59, 0, time.UTC),
					Lat:           37 + 160843/6e5,
					Lon:           -(118 + 150096/6e5),
					Validity:      igc.Validity3D,
					AltBarometric: 4406,
					AltWGS84:      4672,
					Additions: map[string]int{
						"LAD": 3,
						"LOD": 6,
					},
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 6, 0, 0, 1, 0, time.UTC),
					Lat:           37 + 160743/6e5,
					Lon:           -(118 + 150070/6e5),
					Validity:      igc.Validity3D,
					AltBarometric: 4408,
					AltWGS84:      4675,
					Additions: map[string]int{
						"LAD": 3,
						"LOD": 0,
					},
				},
			},
		},
		{
			name: "utc_midnight_rollover_with_new_hfdte_record",
			lines: []string{
				"HFDTEDATE:050624,01",
				"I023636LAD3737LOD",
				"B2359593716084N11815009WA044060467236",
				"HFDTEDATE:060624,01",
				"B0000013716074N11815007WA044080467530",
			},
			expectedRecords: []igc.Record{
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "050624,01",
					},
					Date:         time.Date(2024, time.June, 5, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.IRecord{
					Additions: []igc.BKRecordAddition{
						{StartColumn: 36, FinishColumn: 36, TLC: "LAD"},
						{StartColumn: 37, FinishColumn: 37, TLC: "LOD"},
					},
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 5, 23, 59, 59, 0, time.UTC),
					Lat:           37 + 160843/6e5,
					Lon:           -(118 + 150096/6e5),
					Validity:      igc.Validity3D,
					AltBarometric: 4406,
					AltWGS84:      4672,
					Additions: map[string]int{
						"LAD": 3,
						"LOD": 6,
					},
				},
				&igc.HFDTERecord{
					HRecord: igc.HRecord{
						Source:   igc.SourceFlightRecorder,
						TLC:      "DTE",
						LongName: "DATE",
						Value:    "060624,01",
					},
					Date:         time.Date(2024, time.June, 6, 0, 0, 0, 0, time.UTC),
					FlightNumber: 1,
				},
				&igc.BRecord{
					Time:          time.Date(2024, time.June, 6, 0, 0, 1, 0, time.UTC),
					Lat:           37 + 160743/6e5,
					Lon:           -(118 + 150070/6e5),
					Validity:      igc.Validity3D,
					AltBarometric: 4408,
					AltWGS84:      4675,
					Additions: map[string]int{
						"LAD": 3,
						"LOD": 0,
					},
				},
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := igc.ParseLines(tc.lines)
			assert.NoError(t, err)
			assertEqualErrors(t, tc.expectedErrs, actual.Errs)
			if tc.expectedRecords != nil {
				assert.Equal(t, tc.expectedRecords, actual.Records)
			}
		})
	}
}

func TestParseTestData(t *testing.T) {
	t.Parallel()
	expectedErrorsByName := map[string][]string{
		"2017_08_31_00_14_21_88GGB291.IGC": {
			`8722: "\x1A": unknown record type`,
		},
	}
	dirEntries, err := os.ReadDir("testdata")
	assert.NoError(t, err)
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			file, err := os.Open(filepath.Join("testdata", name))
			assert.NoError(t, err)
			defer file.Close()
			igc, err := igc.Parse(file)
			assert.NoError(t, err)
			assertEqualErrors(t, expectedErrorsByName[name], igc.Errs)
		})
	}
}

func assertEqualErrors(t *testing.T, expectedErrs []string, errs []error) {
	t.Helper()
	if expectedErrs == nil {
		expectedErrs = []string{}
	}
	actualErrs := make([]string, 0, len(errs))
	for _, err := range errs {
		actualErrs = append(actualErrs, err.Error())
	}
	assert.Equal(t, expectedErrs, actualErrs)
}
