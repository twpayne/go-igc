package igc_test

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
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
			name: "c_record_declaration",
			line: "C110524093545000000000502",
			expectedRecord: &igc.CRecordDeclaration{
				DeclarationTime:    time.Date(2024, time.May, 11, 9, 35, 45, 0, time.UTC),
				TaskNumber:         5,
				NumberOfTurnpoints: 2,
			},
		},
		{
			name: "c_record_declaration_xctrack",
			line: "C0110231200370000000000-1 Competition task",
			expectedRecord: &igc.CRecordDeclaration{
				DeclarationTime:    time.Date(2023, time.October, 1, 12, 0, 37, 0, time.UTC),
				NumberOfTurnpoints: -1,
				Text:               " Competition task",
			},
		},
		{
			name: "c_record_waypoint",
			line: "C4415173N00604205ET-296-Trainon",
			expectedRecord: &igc.CRecordWaypoint{
				Lat:  44 + 15173/6e4,
				Lon:  6 + 4205/6e4,
				Text: "T-296-Trainon",
			},
		},
		{
			name: "c_record_waypoint_sw",
			line: "C4415173S00604205WT-296-Trainon",
			expectedRecord: &igc.CRecordWaypoint{
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
				Additions: []igc.RecordAddition{
					{StartColumn: 36, FinishColumn: 38, TLC: "TAS"},
				},
			},
		},
		{
			name: "i_record_two_additions",
			line: "I023638FXA3940SIU",
			expectedRecord: &igc.IRecord{
				Additions: []igc.RecordAddition{
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
				Additions: []igc.RecordAddition{
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
			expectedErr: "1: \"\\x00\": unknown record type\n'\\x00': invalid character",
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
					Additions: []igc.RecordAddition{
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
					Additions: []igc.RecordAddition{
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
					Additions: []igc.RecordAddition{
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
					Additions: []igc.RecordAddition{
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
			name: "b_record_invalid",
			lines: []string{
				"HFDTE310824",
				"B08590616423468S00332739EA02570-3574100224255",
			},
			expectedErrs: []string{
				"2: invalid B record",
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
					Additions: []igc.RecordAddition{
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
			name: "m_record",
			lines: []string{
				"HFDTEDATE:040624,01",
				"M020810HRT1113OXY",
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
				&igc.MRecord{
					Additions: []igc.RecordAddition{
						{StartColumn: 8, FinishColumn: 10, TLC: "HRT"},
						{StartColumn: 11, FinishColumn: 13, TLC: "OXY"},
					},
				},
			},
		},
		{
			name: "n_record",
			lines: []string{
				"HFDTEDATE:040624,01",
				"M020810HRT1113OXY",
				"N123456112098",
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
				&igc.MRecord{
					Additions: []igc.RecordAddition{
						{StartColumn: 8, FinishColumn: 10, TLC: "HRT"},
						{StartColumn: 11, FinishColumn: 13, TLC: "OXY"},
					},
				},
				&igc.NRecord{
					Time: time.Date(2024, time.June, 4, 12, 34, 56, 0, time.UTC),
					Additions: map[string]int{
						"HRT": 112,
						"OXY": 98,
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
					Additions: []igc.RecordAddition{
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
					Additions: []igc.RecordAddition{
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
					Additions: []igc.RecordAddition{
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
					Additions: []igc.RecordAddition{
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
		"MD_85ugkjj1.IGC": {
			`18: invalid C record`,
		},
		"igc-PkwnUq.igc": {
			"48: \"7-00\": syntax error",
			"53: \"8-03\": syntax error",
			"54: \"2-13\": syntax error",
			"55: \"8-05\": syntax error",
			"56: \"2-01\": syntax error",
			"77: \"8-03\": syntax error",
			"78: \"2-13\": syntax error",
			"79: \"8-05\": syntax error",
			"80: \"2-01\": syntax error",
			"133: \"7-00\": syntax error",
			"157: \"7-00\": syntax error",
			"162: \"8-03\": syntax error",
			"163: \"2-13\": syntax error",
			"164: \"8-05\": syntax error",
			"165: \"2-01\": syntax error",
			"197: invalid B record",
			"198: invalid B record",
			"199: invalid B record",
			"200: invalid B record",
			"201: invalid B record",
			"202: invalid B record",
			"203: invalid B record",
			"204: invalid B record",
			"205: invalid B record",
			"206: invalid B record",
			"207: invalid B record",
			"208: invalid B record",
			"209: invalid B record",
			"210: invalid B record",
			"211: invalid B record",
			"212: invalid B record",
			"213: invalid B record",
			"214: invalid B record",
			"215: invalid B record",
			"216: invalid B record",
			"217: invalid B record",
			"218: invalid B record",
			"219: invalid B record",
			"220: invalid B record",
			"221: invalid B record",
			"222: invalid B record",
			"223: invalid B record",
			"224: invalid B record",
			"225: invalid B record",
			"226: invalid B record",
			"227: invalid B record",
			"228: invalid B record",
			"229: invalid B record",
			"230: invalid B record",
			"231: invalid B record",
			"232: invalid B record",
			"233: invalid B record",
			"234: invalid B record",
			"235: invalid B record",
			"236: invalid B record",
			"237: invalid B record",
			"238: invalid B record",
			"239: invalid B record",
			"240: invalid B record",
			"241: invalid B record",
			"242: invalid B record",
			"243: invalid B record",
			"244: invalid B record",
			"245: invalid B record",
			"246: invalid B record",
			"247: invalid B record",
			"248: invalid B record",
			"249: invalid B record",
			"250: invalid B record",
			"251: invalid B record",
			"252: invalid B record",
			"253: invalid B record",
			"254: invalid B record",
			"255: invalid B record",
			"256: invalid B record",
			"266: \"7-00\": syntax error",
			"271: \"8-03\": syntax error",
			"272: \"2-13\": syntax error",
			"273: \"8-05\": syntax error",
			"274: \"2-01\": syntax error",
			"351: \"7-00\": syntax error",
			"356: \"8-03\": syntax error",
			"357: \"2-13\": syntax error",
			"358: \"8-05\": syntax error",
			"359: \"2-01\": syntax error",
			"380: \"8-03\": syntax error",
			"381: \"2-13\": syntax error",
			"382: \"8-05\": syntax error",
			"383: \"2-01\": syntax error",
			"436: \"7-00\": syntax error",
			"460: \"7-00\": syntax error",
			"465: \"8-03\": syntax error",
			"466: \"2-13\": syntax error",
			"467: \"8-05\": syntax error",
			"468: \"2-01\": syntax error",
			"545: \"7-00\": syntax error",
			"550: \"8-03\": syntax error",
			"551: \"2-13\": syntax error",
			"552: \"8-05\": syntax error",
			"553: \"2-01\": syntax error",
			"569: \"7-00\": syntax error",
			"574: \"8-03\": syntax error",
			"575: \"2-13\": syntax error",
			"576: \"8-05\": syntax error",
			"577: \"2-01\": syntax error",
			"654: \"7-00\": syntax error",
			"659: \"8-03\": syntax error",
			"660: \"2-13\": syntax error",
			"661: \"8-05\": syntax error",
			"662: \"2-01\": syntax error",
			"678: \"7-00\": syntax error",
			"683: \"8-03\": syntax error",
			"684: \"2-13\": syntax error",
			"685: \"8-05\": syntax error",
			"686: \"2-01\": syntax error",
			"763: \"7-00\": syntax error",
			"768: \"8-03\": syntax error",
			"769: \"2-13\": syntax error",
			"770: \"8-05\": syntax error",
			"771: \"2-01\": syntax error",
			"848: \"7-00\": syntax error",
			"853: \"8-03\": syntax error",
			"854: \"2-13\": syntax error",
			"855: \"8-05\": syntax error",
			"856: \"2-01\": syntax error",
			"872: \"7-00\": syntax error",
			"877: \"8-03\": syntax error",
			"878: \"2-13\": syntax error",
			"879: \"8-05\": syntax error",
			"880: \"2-01\": syntax error",
		},
	}
	dirEntries, err := os.ReadDir("testdata")
	assert.NoError(t, err)
	for _, dirEntry := range dirEntries {
		name := dirEntry.Name()
		if strings.ToLower(filepath.Ext(dirEntry.Name())) != ".igc" {
			continue
		}
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			file, err := os.Open(filepath.Join("testdata", name))
			assert.NoError(t, err)
			defer file.Close()
			igc, err := igc.Parse(file,
				igc.WithAllowInvalidChars(true),
			)
			assert.NoError(t, err)
			assertEqualErrors(t, expectedErrorsByName[name], igc.Errs)
		})
	}
}

func TestTypes(t *testing.T) {
	for _, value := range []igc.Record{
		&igc.ARecord{},
		&igc.BRecord{},
		&igc.CRecordWaypoint{},
		&igc.CRecordDeclaration{},
		&igc.DRecord{},
		&igc.ERecord{},
		&igc.ERecordWithoutTLC{},
		&igc.FRecord{},
		&igc.GRecord{},
		&igc.HRecord{},
		&igc.HFDTERecord{},
		&igc.HRecordWithInvalidSource{},
		&igc.IRecord{},
		&igc.JRecord{},
		&igc.KRecord{},
		&igc.LRecord{},
		&igc.LRecordWithoutTLC{},
		&igc.MRecord{},
		&igc.NRecord{},
	} {
		_, name, _ := strings.Cut(reflect.TypeOf(value).String(), ".")
		t.Run(name, func(t *testing.T) {
			assert.Equal(t, name[0], value.Type())
		})
	}
}

func FuzzParse(f *testing.F) {
	dirEntries, err := os.ReadDir("testdata")
	assert.NoError(f, err)
	for _, dirEntry := range dirEntries {
		data, err := os.ReadFile(filepath.Join("testdata", dirEntry.Name()))
		assert.NoError(f, err)
		f.Add(data)
	}
	f.Fuzz(func(_ *testing.T, data []byte) {
		_, _ = igc.Parse(bytes.NewReader(data))
	})
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
