package igc

type Manufacturer struct {
	Name string
	TLC  string
	SCC  byte
}

var ApprovedManufacturers = []Manufacturer{
	{Name: "Aircotec", TLC: "ACT", SCC: 'I'},
	{Name: "Avionix", TLC: "AVX"},
	{Name: "Cambridge Aero Instruments", TLC: "CAM", SCC: 'C'},
	{Name: "ClearNav Instruments", TLC: "CNI"},
	{Name: "Data Swan/DSX", TLC: "DSX", SCC: 'D'},
	{Name: "EW Avionics", TLC: "EWA", SCC: 'E'},
	{Name: "Filser", TLC: "FIL", SCC: 'F'},
	{Name: "Flarm (Flight Alarm)", TLC: "FLA", SCC: 'G'},
	{Name: "Flytech", TLC: "FLY"},
	{Name: "Garrecht", TLC: "GCS", SCC: 'A'},
	{Name: "IMI Gliding Equipment", TLC: "IMI", SCC: 'M'},
	{Name: "Logstream", TLC: "LGS"},
	{Name: "LX Navigation", TLC: "LXN", SCC: 'L'},
	{Name: "LXNAV d.o.o.", TLC: "LXV", SCC: 'V'},
	{Name: "Naviter", TLC: "NAV", SCC: 0},
	{Name: "New Technologies s.r.l.", TLC: "NTE", SCC: 'N'},
	{Name: "Nielsen Kellerman", TLC: "NKL", SCC: 'K'},
	{Name: "Peschges", TLC: "PES", SCC: 'P'},
	{Name: "PressFinish Electronics", TLC: "PFE"},
	{Name: "Print Technik", TLC: "PRT", SCC: 'R'},
	{Name: "RC Electronics", TLC: "RCE"},
	{Name: "Scheffel", TLC: "SCH", SCC: 'H'},
	{Name: "Streamline Data Instruments", TLC: "SDI", SCC: 'S'},
	{Name: "Triadis Engineering GmbH", TLC: "TRI", SCC: 'T'},
	{Name: "Zander", TLC: "ZAN", SCC: 'Z'},
}

var ApprovedManufacturersByThreeCharacterCode map[string]*Manufacturer

func init() {
	ApprovedManufacturersByThreeCharacterCode = make(map[string]*Manufacturer, len(ApprovedManufacturers))
	for i, manufacturer := range ApprovedManufacturers {
		ApprovedManufacturersByThreeCharacterCode[manufacturer.TLC] = &ApprovedManufacturers[i]
	}
}
