package models

import "time"

// For shared structs

/**************/
/* V2 Structs */
/**************/

type CapacityRoute struct {
	DataState
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
	SailingDetails
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
	ID             string     `json:"id"`
	Title          string     `json:"title"`
	Body           string     `json:"body,omitempty"`
	BodyObservedAt *time.Time `json:"bodyObservedAt,omitempty"`
	URL            string     `json:"url"`
	Type           string     `json:"type"`
	Severity       string     `json:"severity"`
	TerminalCodes  []string   `json:"terminalCodes,omitempty"`
	RouteCodes     []string   `json:"routeCodes,omitempty"`
	ObservedAt     time.Time  `json:"observedAt"`
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
	DataState
	Notices    []ServiceNotice `json:"notices,omitempty"`
	Terminal   *TerminalInfo   `json:"terminal,omitempty"`
	ObservedAt time.Time       `json:"observedAt"`
}

type NonCapacityResponse struct {
	Routes          []NonCapacityRoute `json:"routes"`
	Meta            ResponseMeta       `json:"meta"`
	SpecialServices []SpecialService   `json:"specialServices"`
}

type NonCapacityRoute struct {
	DataState
	CurrentConditions *ConditionsSnapshot  `json:"currentConditions,omitempty"`
	RouteCode         string               `json:"routeCode"`
	FromTerminalCode  string               `json:"fromTerminalCode"`
	ToTerminalCode    string               `json:"toTerminalCode"`
	SailingDuration   string               `json:"sailingDuration"`
	Sailings          []NonCapacitySailing `json:"sailings"`
	Notices           []ServiceNotice      `json:"notices,omitempty"`
	Terminal          *TerminalInfo        `json:"terminal,omitempty"`
	ObservedAt        time.Time            `json:"observedAt,omitempty"`
}

type NonCapacitySailing struct {
	SailingDetails
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
	Complete      bool                    `json:"complete"`
	Issues        []string                `json:"issues,omitempty"`
}

// Additive fields only: legacy sailing values and the five-column route tables
// remain compatible with older clients and rollback binaries.
type DataState struct {
	DataStatus  string     `json:"dataStatus,omitempty"`
	ServiceDate string     `json:"serviceDate,omitempty"`
	CheckedAt   *time.Time `json:"checkedAt,omitempty"`
	SourceURL   string     `json:"sourceURL,omitempty"`
	Warnings    []string   `json:"warnings,omitempty"`
}

type SailingDetails struct {
	ServiceDate           string     `json:"serviceDate,omitempty"`
	ArrivalDate           string     `json:"arrivalDate,omitempty"`
	ScheduledTime         string     `json:"scheduledTime,omitempty"`
	ActualDepartureTime   string     `json:"actualDepartureTime,omitempty"`
	EstimatedArrivalTime  string     `json:"estimatedArrivalTime,omitempty"`
	ActualArrivalTime     string     `json:"actualArrivalTime,omitempty"`
	Duration              string     `json:"duration,omitempty"`
	Stops                 []Stop     `json:"stops,omitempty"`
	CapacityKnown         *bool      `json:"capacityKnown,omitempty"`
	CarCapacityKnown      *bool      `json:"carCapacityKnown,omitempty"`
	OversizeCapacityKnown *bool      `json:"oversizeCapacityKnown,omitempty"`
	StatusObservedAt      *time.Time `json:"statusObservedAt,omitempty"`
}

type Stop struct {
	Type         string `json:"type"`
	TerminalName string `json:"terminalName,omitempty"`
}

type ConditionsSnapshot struct {
	DataState
	ObservedAt time.Time         `json:"observedAt"`
	Sailings   []CapacitySailing `json:"sailings"`
	Notices    []ServiceNotice   `json:"notices,omitempty"`
	Terminal   *TerminalInfo     `json:"terminal,omitempty"`
}

// The independently operated service has named stops, not standard BCF terminal
// codes. Preserve its complete published matrix and request/connection legends.
type SpecialService struct {
	DataState
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	ObservedAt   time.Time  `json:"observedAt"`
	ValidFrom    string     `json:"validFrom"`
	ValidThrough string     `json:"validThrough"`
	Columns      []string   `json:"columns"`
	Rows         [][]string `json:"rows"`
	Notes        []string   `json:"notes"`
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
