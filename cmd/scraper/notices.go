package scraper

import "strings"

// A route page can include notices for other routes leaving the same terminal.
// Only attach explicitly named terminals/routes to the directions they affect.
var noticeTerminalNames = map[string][]string{
	"TSA": {"tsawwassen"},
	"SWB": {"swartz bay"},
	"HSB": {"horseshoe bay"},
	"DUK": {"duke point"},
	"NAN": {"departure bay"},
	"LNG": {"langdale"},
	"BOW": {"snug cove", "bowen island"},
	"POB": {"otter bay", "pender island"},
	"PVB": {"village bay", "mayne island"},
	"PLH": {"long harbour"},
	"PSB": {"sturdies bay", "galiano island"},
	"PST": {"lyall harbour", "saturna island"},
	"FUL": {"fulford harbour"},
	"VES": {"vesuvius"},
	"PEN": {"penelakut", "telegraph harbour"},
	"THT": {"thetis island"},
	"CHM": {"chemainus"},
	"CFT": {"crofton"},
	"NAH": {"nanaimo harbour"},
	"GAB": {"gabriola", "descanso bay"},
	"CMX": {"little river"},
	"PWR": {"westview"},
	"TEX": {"blubber bay", "texada"},
	"SLT": {"saltery bay"},
	"ERL": {"earls cove"},
	"BTW": {"brentwood bay"},
	"MIL": {"mill bay"},
	"BKY": {"buckley bay"},
	"DNM": {"denman island west"},
	"DNE": {"denman island east"},
	"HRN": {"hornby", "shingle spit"},
	"CAM": {"campbell river"},
	"QDR": {"quathiaski cove"},
	"HRB": {"heriot bay"},
	"COR": {"cortes", "whaletown"},
	"MCN": {"port mcneill"},
	"ALR": {"alert bay"},
	"SOI": {"sointula"},
	"PPH": {"port hardy", "bear cove"},
	"PPR": {"prince rupert"},
	"BEC": {"bella coola"},
	"PBB": {"bella bella", "mcloughlin bay"},
	"POF": {"ocean falls"},
	"SHW": {"shearwater"},
	"KLE": {"klemtu"},
	"PSK": {"skidegate"},
	"ALF": {"alliford bay"},
	"SGI": {"southern gulf islands"},
}

func noticeScope(title, routeCode, origin string) ([]string, bool) {
	lower := strings.ToLower(title)
	var codes []string
	for code, names := range noticeTerminalNames {
		for _, name := range names {
			if strings.Contains(lower, name) {
				codes = append(codes, code)
				break
			}
		}
	}
	if len(codes) == 0 {
		return []string{origin}, true
	} // Keep unscoped operational notices supplied on this route page.
	route := strings.ReplaceAll(routeCode, "-", "")
	if len(route) != 6 {
		return codes, false
	}
	from, to := route[:3], route[3:]
	matches := func(endpoint string) bool {
		for _, code := range codes {
			if endpoint == code {
				return true
			}
			if endpoint == "SGI" && strings.Contains("|POB|PVB|PLH|PSB|PST|", "|"+code+"|") {
				return true
			}
			if code == "SGI" && strings.Contains("|POB|PVB|PLH|PSB|PST|", "|"+endpoint+"|") {
				return true
			}
		}
		return false
	}
	if len(codes) > 1 {
		return codes, matches(from) && matches(to)
	}
	return codes, matches(from) || matches(to)
}
