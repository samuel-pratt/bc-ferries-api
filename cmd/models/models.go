package models

import "time"

// For shared structs

/**************/
/* V2 Structs */
/**************/

type CapacityRoute struct {
	RouteCode        string            `json:"routeCode"`
	FromTerminalCode string            `json:"fromTerminalCode"`
	ToTerminalCode   string            `json:"toTerminalCode"`
	SailingDuration  string            `json:"sailingDuration"`
	Sailings         []CapacitySailing `json:"sailings"`
	Notices          []ServiceNotice   `json:"notices,omitempty"`
	Terminal         *TerminalInfo     `json:"terminal,omitempty"`
	ObservedAt       time.Time         `json:"observedAt,omitempty"`
}

type CapacitySailing struct {
	DepartureTime string    `json:"time"`
	ArrivalTime   string    `json:"arrivalTime"`
	SailingStatus string    `json:"sailingStatus"`
	Fill          int       `json:"fill"`
	CarFill       int       `json:"carFill"`
	OversizeFill  int       `json:"oversizeFill"`
	VesselName    string    `json:"vesselName"`
	VesselStatus  string    `json:"vesselStatus"`
	Status        string    `json:"status,omitempty"`
	StatusText    string    `json:"statusText,omitempty"`
	DelayMinutes  *int      `json:"delayMinutes,omitempty"`
	ObservedAt    time.Time `json:"observedAt,omitempty"`
}

type ServiceNotice struct {
	ID            string    `json:"id"`
	Title         string    `json:"title"`
	URL           string    `json:"url"`
	Type          string    `json:"type"`
	Severity      string    `json:"severity"`
	TerminalCodes []string  `json:"terminalCodes,omitempty"`
	RouteCodes    []string  `json:"routeCodes,omitempty"`
	ObservedAt    time.Time `json:"observedAt"`
}

type Webcam struct {
	ID         string    `json:"id"`
	Label      string    `json:"label"`
	ImageURL   string    `json:"imageURL"`
	ObservedAt time.Time `json:"observedAt"`
}

type TerminalInfo struct {
	Code          string    `json:"code"`
	Name          string    `json:"name"`
	DetailsURL    string    `json:"detailsURL,omitempty"`
	DirectionsURL string    `json:"directionsURL,omitempty"`
	Webcams       []Webcam  `json:"webcams,omitempty"`
	ObservedAt    time.Time `json:"observedAt"`
}

type RouteOperationalData struct {
	Notices    []ServiceNotice `json:"notices,omitempty"`
	Terminal   *TerminalInfo   `json:"terminal,omitempty"`
	ObservedAt time.Time       `json:"observedAt"`
}

type NonCapacityResponse struct {
	Routes []NonCapacityRoute `json:"routes"`
	Meta   ResponseMeta       `json:"meta"`
}

type NonCapacityRoute struct {
	RouteCode        string               `json:"routeCode"`
	FromTerminalCode string               `json:"fromTerminalCode"`
	ToTerminalCode   string               `json:"toTerminalCode"`
	SailingDuration  string               `json:"sailingDuration"`
	Sailings         []NonCapacitySailing `json:"sailings"`
	Notices          []ServiceNotice      `json:"notices,omitempty"`
	Terminal         *TerminalInfo        `json:"terminal,omitempty"`
	ObservedAt       time.Time            `json:"observedAt,omitempty"`
}

type NonCapacitySailing struct {
	DepartureTime string    `json:"time"`
	ArrivalTime   string    `json:"arrivalTime"`
	VesselName    string    `json:"vesselName"`
	VesselStatus  string    `json:"vesselStatus"`
	Status        string    `json:"status,omitempty"`
	StatusText    string    `json:"statusText,omitempty"`
	DelayMinutes  *int      `json:"delayMinutes,omitempty"`
	ObservedAt    time.Time `json:"observedAt,omitempty"`
}

type SourceHealth struct {
	Status           string     `json:"status"`
	ObservedAt       *time.Time `json:"observedAt,omitempty"`
	LastSuccessfulAt *time.Time `json:"lastSuccessfulAt,omitempty"`
}

type ResponseMeta struct {
	SchemaVersion string                  `json:"schemaVersion"`
	GeneratedAt   time.Time               `json:"generatedAt"`
	Stale         bool                    `json:"stale"`
	Sources       map[string]SourceHealth `json:"sources"`
}

/**************/
/* V1 Structs */
/**************/

type Route struct {
	SailingDuration string    `json:"sailingDuration"`
	Sailings        []Sailing `json:"sailings"`
}

type Sailing struct {
	DepartureTime string `json:"time"`
	ArrivalTime   string `json:"arrivalTime"`
	IsCancelled   bool   `json:"isCancelled"`
	Fill          int    `json:"fill"`
	CarFill       int    `json:"carFill"`
	OversizeFill  int    `json:"oversizeFill"`
	VesselName    string `json:"vesselName"`
	VesselStatus  string `json:"vesselStatus"`
}
