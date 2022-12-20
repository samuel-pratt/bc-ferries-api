package main

func GetDepartureTerminals() []string {
	departureTerminals := [8]string{
		"TSA",
		"SWB",
		"HSB",
		"DUK",
		"LNG",
		"NAN",
		"FUL",
		"BOW",
	}

	return departureTerminals[:]
}

func GetDestinationTerminals() [][]string {
	destinationTerminals := [8][]string{
		{"SWB", "SGI", "DUK"},
		{"TSA", "FUL", "SGI"},
		{"NAN", "LNG", "BOW"},
		{"TSA"},
		{"HSB"},
		{"HSB"},
		{"SWB"},
		{"HSB"},
	}

	return destinationTerminals[:]
}
