package arinc

import (
	"fmt"
	"strconv"
)

const (
	SectionCodeNavaid  string = "D"
	SectionCodeEnroute string = "E"
	SectionCodeAirport string = "P"

	SubsectionCodeNavaidNDB         = "B"
	SubsectionCodeNavaidVHF         = " "
	SubsectionCodeEnrouteWaypoint   = "A"
	SubsectionCodeTerminalWaypoint  = "C"
	SubsectionCodeApproachProcedure = "F"
	SubsectionCodeLocGS             = "I"

	ContinuationRecordSimulation = "S"
)

// Record is a base struct that all ARINC records follow.
type Record struct {
	RecordType       string `fixed:"1,1"`
	CustomerAreaCode string `fixed:"2,4"`
	SectionCode      string `fixed:"5,5"`
	SubsectionCode   string `fixed:"6,6"`
	Data             string `fixed:"7,123"`
	FileRecordNumber string `fixed:"124,128"`
	CycleDate        string `fixed:"129,132"`
}

// VHFNavaidRecord is a record for a VHF navaid.
type VHFNavaidRecord struct {
	Record                   `fixed:"1,6"`
	AirportICAOID            string `fixed:"7,10"`
	ICAOCode1                string `fixed:"11,12"`
	VORID                    string `fixed:"14,17"`
	ICAOCode2                string `fixed:"20,21"`
	ContinuationRecordNumber string `fixed:"22,22"`
	VORFrequency             string `fixed:"23,27"`
	NavaidClass              string `fixed:"28,32"`
	VORLatitude              string `fixed:"33,41"`
	VORLongitude             string `fixed:"42,51"`
	DMEID                    string `fixed:"52,55"`
	DMELatitude              string `fixed:"56,64"`
	DMELongitude             string `fixed:"65,74"`
	StationDeclination       string `fixed:"75,79"`
	DMEElevation             string `fixed:"80,84"`
	FigureOfMerit            string `fixed:"85,85"`
	ILSDMEBias               string `fixed:"86,87"`
	FrequencyProtection      string `fixed:"88,90"`
	DatumCode                string `fixed:"91,93"`
	VORName                  string `fixed:"94,123"`
}

// NDBNavaidRecord is a record for an NDB.
type NDBNavaidRecord struct {
	Record                   `fixed:"1,6"`
	AirportID                string `fixed:"7,10"`
	ICAOCode1                string `fixed:"11,12"`
	NDBID                    string `fixed:"14,17"`
	ICAOCode2                string `fixed:"20,21"`
	ContinuationRecordNumber string `fixed:"22,22"`
	NDBFrequency             string `fixed:"23,27"`
	NDBClass                 string `fixed:"28,32"`
	NDBLatitude              string `fixed:"33,41"`
	NDBLongitude             string `fixed:"42,51"`
	MagneticVar              string `fixed:"75,79"`
	DatumCode                string `fixed:"91,93"`
	NDBName                  string `fixed:"94,123"`
}

// AirportEnrouteRecord is a record associated with an airport or enroute.
type AirportEnrouteRecord struct {
	Record         `fixed:"1,6"`
	AirportID      string `fixed:"7,10"`
	ICAOCode       string `fixed:"11,12"`
	SubsectionCode string `fixed:"13,13"`
	Data           string `fixed:"14,123"`
}

// WaypointPrimaryRecord is a record associated with a waypoint.
type WaypointPrimaryRecord struct {
	AirportEnrouteRecord     `fixed:"1,13"`
	WaypointID               string `fixed:"14,18"`
	ICAOCode                 string `fixed:"20,21"`
	ContinuationRecordNumber string `fixed:"22,22"`
	WaypointType             string `fixed:"27,29"`
	WaypointUsage            string `fixed:"30,31"`
	WaypointLatitude         string `fixed:"33,41"`
	WaypointLongitude        string `fixed:"42,51"`
	DynamicMagVar            string `fixed:"75,79"`
	DatumCode                string `fixed:"85,87"`
	NameFormatIndicator      string `fixed:"96,98"`
	WaypointNameDesc         string `fixed:"99,123"`
}

// parsePoint parses a latitude or longitude string into degrees, minutes, and
// seconds. All numerical values are negative if the direction is 'S' or 'W'.
func parsePoint(point string) (int, int, float64, error) {
	if len(point) < 9 || len(point) > 10 {
		return 0, 0, 0.0, fmt.Errorf("invalid length for latitude or longitude string: %v", point)
	}
	// If the string is a latitude string (length 9), add an extra 0 to make
	// math easier.
	if len(point) == 9 {
		point = fmt.Sprintf("%s0%s", point[0:1], point[1:9])
	}
	dir := 1
	if point[0] == 'S' || point[0] == 'W' {
		dir *= -1
	}
	degrees, err := strconv.Atoi(point[1:4])
	if err != nil {
		return 0, 0.0, 0.0, fmt.Errorf("invalid degrees: %q", point[1:4])
	}

	minutes, err := strconv.Atoi(point[4:6])
	if err != nil {
		return 0, 0.0, 0.0, fmt.Errorf("invalid minutes: %q", point[4:6])
	}
	seconds, err := strconv.ParseFloat(fmt.Sprintf("%s.%s", point[6:8], point[8:10]), 64)
	if err != nil {
		return 0, 0.0, 0.0, fmt.Errorf("invalid seconds: %q", point[6:10])
	}
	return dir * degrees, dir * minutes, float64(dir) * seconds, nil
}

// LatLon calculates the numerical latitude and longitude for the provided
// latitude and longitude strings. If the data is invalid, an error is returned.
func LatLon(latitude, longitude string) (float64, float64, error) {
	latDeg, latMin, latSec, err := parsePoint(latitude)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("could not calculate latitude: %v", err)
	}
	lonDeg, lonMin, lonSec, err := parsePoint(longitude)
	if err != nil {
		return 0.0, 0.0, fmt.Errorf("could not calculate longitude: %v", err)
	}
	lat := float64(latDeg) + (float64(latMin) / 60.0) + (latSec / 3600.0)
	lon := float64(lonDeg) + (float64(lonMin) / 60.0) + (lonSec / 3600.0)
	return lat, lon, nil
}

// AirportLocGSPrimaryRecord ia a record for a glideslope or localizer at an airport.
// See 4.1.11.1 Airport and Heliport Localizer and Glide Slope Primary Records
type AirportLocGSPrimaryRecord struct {
	AirportEnrouteRecord             `fixed:"1,13"`
	LocalizerID                      string `fixed:"14,17"`
	ILSCategory                      string `fixed:"18,18"`
	ContinuationRecordNumber         string `fixed:"22,22"`
	LocalizerFrequency               string `fixed:"23,27"`
	RunwayIdentifier                 string `fixed:"28,32"`
	LocalizerLatitude                string `fixed:"33,41"`
	LocalizerLongitude               string `fixed:"42,51"`
	LocalizerBearing                 string `fixed:"52,55"`
	GlideSlopeLatitude               string `fixed:"56,64"`
	GlideSlopeLongitude              string `fixed:"65,74"`
	LocalizerPosition                string `fixed:"75,78"`
	LocalizerPositionReference       string `fixed:"79,79"`
	GlideSlopePosition               string `fixed:"80,83"`
	LocalizerWidth                   string `fixed:"84,87"`
	GlideSlopeAngle                  string `fixed:"88,90"`
	StationDeclination               string `fixed:"91,95"`
	GlideSlopeHeightAtThreshold      string `fixed:"96,97"`
	GlideSlopeElevation              string `fixed:"98,102"`
	SupportingFacilityID             string `fixed:"103,106"`
	SupportingFacilityICAOCode       string `fixed:"107,108"`
	SupportingFacilitySectionCode    string `fixed:"109,109"`
	SupportingFacilitySubsectionCode string `fixed:"110,110"`
}

// AirportLocGSSimContinuationRecord is a continuation record for an AirportLocGSPrimaryRecord.
// See 4.1.11.3 Airport and Heliport Localizer and Glide Slope Simulation Continuation Records
type AirportLocGSSimContinuationRecord struct {
	AirportEnrouteRecord     `fixed:"1,13"`
	LocalizerID              string `fixed:"14,17"`
	ILSCategory              string `fixed:"18,18"`
	ContinuationRecordNumber string `fixed:"22,22"`
	ApplicationType          string `fixed:"23,23"`
	FacilityCharacteristics  string `fixed:"24,27"`
	LocalizerTrueBearing     string `fixed:"52,56"`
	LocalizerBearingSource   string `fixed:"57,57"`
	GlideSlopeBeamWidth      string `fixed:"88,90"`
	ApproachRouteIdent1      string `fixed:"91,96"`
	ApproachRouteIdent2      string `fixed:"97,102"`
	ApproachRouteIdent3      string `fixed:"103,108"`
	ApproachRouteIdent4      string `fixed:"109,114"`
	ApproachRouteIdent5      string `fixed:"115,120"`
}

// AirportProcedurePrimaryRecord is a record for a SID, STAR, or approach procedure
// at an airport.
// See 4.1.9.1 Airport SID/STAR/Approach Primary Records
type AirportProcedurePrimaryRecord struct {
	AirportEnrouteRecord         `fixed:"1,13"`
	ProcedureID                  string `fixed:"14,19"`
	RouteType                    string `fixed:"20,20"`
	TransitionID                 string `fixed:"21,25"`
	SequenceNumber               string `fixed:"27,29"`
	FixID                        string `fixed:"30,34"`
	ProcedureICAOCode            string `fixed:"35,36"`
	ProcedureSectionCode         string `fixed:"37,37"`
	ProcedureSubsectionCode      string `fixed:"38,38"`
	ContinuationRecordNumber     string `fixed:"39,39"`
	WaypointDescriptionCode      string `fixed:"40,43"`
	TurnDirection                string `fixed:"44,44"`
	RNP                          string `fixed:"45,47"`
	PathAndTermination           string `fixed:"48,49"`
	TurnDirectionValid           string `fixed:"50,50"`
	RecommendedNavaid            string `fixed:"51,54"`
	RecommendedNavaidICAOCode    string `fixed:"55,56"`
	ArcRadius                    string `fixed:"57,62"`
	Theta                        string `fixed:"63,66"`
	Rho                          string `fixed:"67,70"`
	MagneticCourse               string `fixed:"71,74"`
	RouteOrHoldingDistanceOrTime string `fixed:"75,78"`
	RecommendedNavSection        string `fixed:"79,79"`
	RecommendedNavSubsection     string `fixed:"80,80"`
	AltitudeDescription          string `fixed:"83,83"`
	ATCIndicator                 string `fixed:"84,84"`
	Altitude1                    string `fixed:"85,89"`
	Altitude2                    string `fixed:"90,94"`
	TransitionAltitude           string `fixed:"95,99"`
	SpeedLimit                   string `fixed:"100,102"`
	VerticalAngle                string `fixed:"103,106"`
	CenterFixOrTAASectorID       string `fixed:"107,111"`
	MultipleCodeOrTAASectorID    string `fixed:"112,112"`
	CenterFixICAOCode            string `fixed:"113,114"`
	CenterFixSectionCode         string `fixed:"115,115"`
	CenterFixSubsectionCode      string `fixed:"116,116"`
	GpsFmsIndication             string `fixed:"117,117"`
	SpeedLimitDescription        string `fixed:"118,118"`
	ApproachRouteQualifier1      string `fixed:"119,119"`
	ApproachRouteQualifier2      string `fixed:"120,120"`
}

// IsLocalizerFrontCourseApproach returns true if the approach procedure is a
// LOC, SDF, ILS, or LDA approach.
func (p *AirportProcedurePrimaryRecord) IsLocalizerFrontCourseApproach() bool {
	id := p.ProcedureID
	if len(id) >= 1 && (id[0] == 'I' || id[0] == 'L' || id[0] == 'U' || id[0] == 'X') {
		return true
	}
	return false
}

// IsFinalApproachFix returns true if the record is for the final approach
// fix on an approach procedure.
func (p *AirportProcedurePrimaryRecord) IsFinalApproachFix() bool {
	d := p.WaypointDescriptionCode
	if len(d) >= 4 && d[3] == 'F' {
		return true
	}
	return false
}
