package staticdata

import "sort"

func Pairs(from []string, to [][]string) []string {
	seen := map[string]bool{}
	for i, a := range from {
		for _, b := range to[i] {
			seen[a+"-"+b] = true
		}
	}
	out := []string{}
	for p := range seen {
		out = append(out, p)
	}
	sort.Strings(out)
	return out
}
func SchedulePairs() []string {
	return Pairs(GetNonCapacityDepartureTerminals(), GetNonCapacityDestinationTerminals())
}
func CapacityPairs() []string {
	return Pairs(GetCapacityDepartureTerminals(), GetCapacityDestinationTerminals())
}
func ConditionPairs() []string {
	scheduleOnly := map[string]bool{"PLH-SWB": true, "SWB-POB": true, "SWB-PSB": true, "SWB-PST": true, "SWB-PVB": true, "TSA-PLH": true, "TSA-POB": true, "TSA-PSB": true, "TSA-PST": true, "TSA-PVB": true}
	out := []string{"TSA-SGI", "SWB-SGI"}
	for _, p := range SchedulePairs() {
		if !scheduleOnly[p] {
			out = append(out, p)
		}
	}
	sort.Strings(out)
	return out
}
