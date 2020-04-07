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
	SubsectionCodeNavaidVHF         = ""
	SubsectionCodeEnrouteWaypoint   = "A"
	SubsectionCodeTerminalWaypoint  = "C"
	SubsectionCodeApproachProcedure = "F"
	SubsectionCodeLocGS             = "I"

	ContinuationRecordSimulation  = "S"
	LocalizerBearingSourceNotGovt = "N"
)

// Record is a base struct that all ARINC records follow.
type Record struct {
	RecordType       string `fixed:"1,1,left"`
	CustomerAreaCode string `fixed:"2,4,left"`
	SectionCode      string `fixed:"5,5,left"`
	SubsectionCode   string `fixed:"6,6,left"`
	Data             string `fixed:"7,123,left"`
	FileRecordNumber string `fixed:"124,128,left"`
	CycleDate        string `fixed:"129,132,left"`
}

// VHFNavaidRecord is a record for a VHF navaid.
type VHFNavaidRecord struct {
	Record                   `fixed:"1,6,left"`
	AirportICAOID            string `fixed:"7,10,left"`
	ICAOCode1                string `fixed:"11,12,left"`
	VORID                    string `fixed:"14,17,left"`
	ICAOCode2                string `fixed:"20,21,left"`
	ContinuationRecordNumber string `fixed:"22,22,left"`
	VORFrequency             string `fixed:"23,27,left"`
	NavaidClass              string `fixed:"28,32,left"`
	VORLatitude              string `fixed:"33,41,left"`
	VORLongitude             string `fixed:"42,51,left"`
	DMEID                    string `fixed:"52,55,left"`
	DMELatitude              string `fixed:"56,64,left"`
	DMELongitude             string `fixed:"65,74,left"`
	StationDeclination       string `fixed:"75,79,left"`
	DMEElevation             string `fixed:"80,84,left"`
	FigureOfMerit            string `fixed:"85,85,left"`
	ILSDMEBias               string `fixed:"86,87,left"`
	FrequencyProtection      string `fixed:"88,90,left"`
	DatumCode                string `fixed:"91,93,left"`
	VORName                  string `fixed:"94,123,left"`
}

// NDBNavaidRecord is a record for an NDB.
type NDBNavaidRecord struct {
	Record                   `fixed:"1,6,left"`
	AirportID                string `fixed:"7,10,left"`
	ICAOCode1                string `fixed:"11,12,left"`
	NDBID                    string `fixed:"14,17,left"`
	ICAOCode2                string `fixed:"20,21,left"`
	ContinuationRecordNumber string `fixed:"22,22,left"`
	NDBFrequency             string `fixed:"23,27,left"`
	NDBClass                 string `fixed:"28,32,left"`
	NDBLatitude              string `fixed:"33,41,left"`
	NDBLongitude             string `fixed:"42,51,left"`
	MagneticVar              string `fixed:"75,79,left"`
	DatumCode                string `fixed:"91,93,left"`
	NDBName                  string `fixed:"94,123,left"`
}

// AirportEnrouteRecord is a record associated with an airport or enroute.
type AirportEnrouteRecord struct {
	Record         `fixed:"1,6,left"`
	AirportID      string `fixed:"7,10,left"`
	ICAOCode       string `fixed:"11,12,left"`
	SubsectionCode string `fixed:"13,13,left"`
	Data           string `fixed:"14,123,left"`
}

// WaypointPrimaryRecord is a record associated with a waypoint.
type WaypointPrimaryRecord struct {
	AirportEnrouteRecord     `fixed:"1,13,left"`
	WaypointID               string `fixed:"14,18,left"`
	ICAOCode                 string `fixed:"20,21,left"`
	ContinuationRecordNumber string `fixed:"22,22,left"`
	WaypointType             string `fixed:"27,29,left"`
	WaypointUsage            string `fixed:"30,31,left"`
	WaypointLatitude         string `fixed:"33,41,left"`
	WaypointLongitude        string `fixed:"42,51,left"`
	DynamicMagVar            string `fixed:"75,79,left"`
	DatumCode                string `fixed:"85,87,left"`
	NameFormatIndicator      string `fixed:"96,98,left"`
	WaypointNameDesc         string `fixed:"99,123,left"`
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

// EncodeBearing encodes the specified bearing into a five character string,
// where the decimal point is implied after the third character. If the provided
// bearing is negative or greater than 360, then the output is undefined.
// Example: EncodeBearing(190.123) = "19012"
func EncodeBearing(bearing float64) string {
	s := fmt.Sprintf("%06.2f", bearing)
	return s[:3] + s[4:]
}

// AirportLocGSPrimaryRecord ia a record for a glideslope or localizer at an airport.
// See 4.1.11.1 Airport and Heliport Localizer and Glide Slope Primary Records
type AirportLocGSPrimaryRecord struct {
	AirportEnrouteRecord             `fixed:"1,13,left"`
	LocalizerID                      string `fixed:"14,17,left"`
	ILSCategory                      string `fixed:"18,18,left"`
	ContinuationRecordNumber         string `fixed:"22,22,left"`
	LocalizerFrequency               string `fixed:"23,27,left"`
	RunwayIdentifier                 string `fixed:"28,32,left"`
	LocalizerLatitude                string `fixed:"33,41,left"`
	LocalizerLongitude               string `fixed:"42,51,left"`
	LocalizerBearing                 string `fixed:"52,55,left"`
	GlideSlopeLatitude               string `fixed:"56,64,left"`
	GlideSlopeLongitude              string `fixed:"65,74,left"`
	LocalizerPosition                string `fixed:"75,78,left"`
	LocalizerPositionReference       string `fixed:"79,79,left"`
	GlideSlopePosition               string `fixed:"80,83,left"`
	LocalizerWidth                   string `fixed:"84,87,left"`
	GlideSlopeAngle                  string `fixed:"88,90,left"`
	StationDeclination               string `fixed:"91,95,left"`
	GlideSlopeHeightAtThreshold      string `fixed:"96,97,left"`
	GlideSlopeElevation              string `fixed:"98,102,left"`
	SupportingFacilityID             string `fixed:"103,106,left"`
	SupportingFacilityICAOCode       string `fixed:"107,108,left"`
	SupportingFacilitySectionCode    string `fixed:"109,109,left"`
	SupportingFacilitySubsectionCode string `fixed:"110,110,left"`
	Data                             string `fixed:"124,132,left"`
}

// AirportLocGSSimContinuationRecord is a continuation record for an AirportLocGSPrimaryRecord.
// See 4.1.11.3 Airport and Heliport Localizer and Glide Slope Simulation Continuation Records
type AirportLocGSSimContinuationRecord struct {
	AirportEnrouteRecord     `fixed:"1,13,left"`
	LocalizerID              string `fixed:"14,17,left"`
	ILSCategory              string `fixed:"18,18,left"`
	ContinuationRecordNumber string `fixed:"22,22,left"`
	ApplicationType          string `fixed:"23,23,left"`
	FacilityCharacteristics  string `fixed:"24,27,left"`
	LocalizerTrueBearing     string `fixed:"52,56,left"`
	LocalizerBearingSource   string `fixed:"57,57,left"`
	GlideSlopeBeamWidth      string `fixed:"88,90,left"`
	ApproachRouteIdent1      string `fixed:"91,96,left"`
	ApproachRouteIdent2      string `fixed:"97,102,left"`
	ApproachRouteIdent3      string `fixed:"103,108,left"`
	ApproachRouteIdent4      string `fixed:"109,114,left"`
	ApproachRouteIdent5      string `fixed:"115,120,left"`
	Data                     string `fixed:"124,132,left"`
}

// AirportProcedurePrimaryRecord is a record for a SID, STAR, or approach procedure
// at an airport.
// See 4.1.9.1 Airport SID/STAR/Approach Primary Records
type AirportProcedurePrimaryRecord struct {
	AirportEnrouteRecord         `fixed:"1,13,left"`
	ProcedureID                  string `fixed:"14,19,left"`
	RouteType                    string `fixed:"20,20,left"`
	TransitionID                 string `fixed:"21,25,left"`
	SequenceNumber               string `fixed:"27,29,left"`
	FixID                        string `fixed:"30,34,left"`
	ProcedureICAOCode            string `fixed:"35,36,left"`
	ProcedureSectionCode         string `fixed:"37,37,left"`
	ProcedureSubsectionCode      string `fixed:"38,38,left"`
	ContinuationRecordNumber     string `fixed:"39,39,left"`
	WaypointDescriptionCode      string `fixed:"40,43,left"`
	TurnDirection                string `fixed:"44,44,left"`
	RNP                          string `fixed:"45,47,left"`
	PathAndTermination           string `fixed:"48,49,left"`
	TurnDirectionValid           string `fixed:"50,50,left"`
	RecommendedNavaid            string `fixed:"51,54,left"`
	RecommendedNavaidICAOCode    string `fixed:"55,56,left"`
	ArcRadius                    string `fixed:"57,62,left"`
	Theta                        string `fixed:"63,66,left"`
	Rho                          string `fixed:"67,70,left"`
	MagneticCourse               string `fixed:"71,74,left"`
	RouteOrHoldingDistanceOrTime string `fixed:"75,78,left"`
	RecommendedNavSection        string `fixed:"79,79,left"`
	RecommendedNavSubsection     string `fixed:"80,80,left"`
	AltitudeDescription          string `fixed:"83,83,left"`
	ATCIndicator                 string `fixed:"84,84,left"`
	Altitude1                    string `fixed:"85,89,left"`
	Altitude2                    string `fixed:"90,94,left"`
	TransitionAltitude           string `fixed:"95,99,left"`
	SpeedLimit                   string `fixed:"100,102,left"`
	VerticalAngle                string `fixed:"103,106,left"`
	CenterFixOrTAASectorID       string `fixed:"107,111,left"`
	MultipleCodeOrTAASectorID    string `fixed:"112,112,left"`
	CenterFixICAOCode            string `fixed:"113,114,left"`
	CenterFixSectionCode         string `fixed:"115,115,left"`
	CenterFixSubsectionCode      string `fixed:"116,116,left"`
	GpsFmsIndication             string `fixed:"117,117,left"`
	SpeedLimitDescription        string `fixed:"118,118,left"`
	ApproachRouteQualifier1      string `fixed:"119,119,left"`
	ApproachRouteQualifier2      string `fixed:"120,120,left"`
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
