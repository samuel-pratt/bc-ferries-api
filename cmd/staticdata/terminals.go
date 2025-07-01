package staticdata

/*
 * GetCapacityDepartureTerminals
 *
 * Returns an array of departure terminals for capacity routes
 *
 * @return []string
 */
func GetCapacityDepartureTerminals() []string {
	departureTerminals := [6]string{
		"TSA",
		"SWB",
		"HSB",
		"DUK",
		"LNG",
		"NAN",
	}

	return departureTerminals[:]
}

/*
 * GetCapacityDestinationTerminals
 *
 * Returns an array of destination terminals for capacity routes
 *
 * @return [][]string
 */
func GetCapacityDestinationTerminals() [][]string {
	destinationTerminals := [6][]string{
		{"SWB", "SGI", "DUK"},
		{"TSA", "FUL", "SGI"},
		{"NAN", "LNG", "BOW"},
		{"TSA"},
		{"HSB"},
		{"HSB"},
	}

	return destinationTerminals[:]
}

/*
 * GetNonCapacityDepartureTerminals
 *
 * Returns an array of departure terminals for non-capacity routes
 *
 * @return []string
 */
func GetNonCapacityDepartureTerminals() []string {
	departureTerminals := [45]string{
		"TSA", "HSB", "SWB", "NAN", "DUK",
		"NAH", "CMX", "PPH", "BTW", "BKY",
		"CAM", "CHM", "CFT", "MIL", "MCN",
		"LNG", "PWR", "SLT", "ERL", "TEX",
		"POB", "PSB", "PVB", "PST", "GAB",
		"PEN", "PLH", "VES", "FUL", "THT",
		"ALR", "DNM", "DNE", "HRN", "SOI",
		"HRB", "QDR", "BEC", "PBB", "POF",
		"SHW", "KLE", "PPR", "PSK", "ALF",
	}

	return departureTerminals[:]
}

/*
 * GetNonCapacityDestinationTerminals
 *
 * Returns an array of destination terminals for non-capacity routes
 *
 * @return [][]string
 */
func GetNonCapacityDestinationTerminals() [][]string {
	destinationTerminals := [45][]string{
		{"PSB", "PVB", "DUK", "POB", "PLH", "PST", "SWB"},
		{"BOW", "NAN", "LNG"},
		{"PSB", "PVB", "POB", "FUL", "PST", "TSA", "PSB", "PVB", "POB", "FUL", "PST"},
		{"HSB"},
		{"TSA"},
		{"GAB"},
		{"PWR"},
		{"PBB", "BEC", "KLE", "POF", "PPR", "SHW", "PBB", "BEC", "KLE", "POF", "PPR", "SHW"},
		{"MIL"},
		{"DNM"},
		{"QDR"},
		{"PEN", "THT", "PEN", "THT"},
		{"VES"},
		{"BTW"},
		{"ALR", "SOI"},
		{"HSB"},
		{"CMX", "TEX"},
		{"ERL"},
		{"SLT"},
		{"PWR"},
		{"PSB", "PVB", "PLH", "PST", "TSA", "SWB"},
		{"PVB", "POB", "PLH", "PST", "TSA", "SWB"},
		{"PSB", "POB", "PLH", "PST", "TSA", "SWB"},
		{"PSB", "PVB", "POB", "PLH", "TSA", "SWB"},
		{"NAH"},
		{"CHM", "THT"},
		{"PSB", "PVB", "POB", "PST", "TSA", "SWB"},
		{"CFT"},
		{"SWB"},
		{"CHM", "PEN"},
		{"SOI", "MCN"},
		{"BKY"},
		{"HRN"},
		{"DNE"},
		{"ALR", "MCN"},
		{"COR"},
		{"CAM"},
		{"PBB", "POF", "PPH", "SHW"},
		{"BEC", "KLE", "POF", "PPH", "PPR", "SHW"},
		{"PBB", "BEC", "PPH", "SHW"},
		{"PBB", "BEC", "POF", "PPH"},
		{"PBB", "PPH", "PPR", "PBB", "PPH"},
		{"PBB", "PSK", "KLE", "PPH", "PSK"},
		{"ALF", "PPR"},
		{"PSK"},
	}

	return destinationTerminals[:]
}
