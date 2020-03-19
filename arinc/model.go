package arinc

const (
	SectionCodeAirport string = "P"

	SubsectionCodeApproachProcedure = "F"
	SubsectionCodeLocGS = "I"
)

// Record is a base struct that all ARINC records follow.
type Record struct {
	RecordType       string `fixed:"1,1"`
	CustomerAreaCode string `fixed:"2,4"`
	SectionCode      string `fixed:"5,5"`
	Data             string `fixed:"6,123"`
	FileRecordNumber string `fixed:"124,128"`
	CycleDate        string `fixed:"129,132"`
}

// AirportRecord is a record associated with an airport.
type AirportRecord struct {
	Record            `fixed:"1,6"`
	AirportIdentifier string `fixed:"7,10"`
	ICAOCode          string `fixed:"11,12"`
	SubsectionCode    string `fixed:"13,13"`
	Data              string `fixed:"14,123"`
}

// AirportLocGSPrimaryRecord ia a record for a glideslope or localizer at an airport.
// See 4.1.11.1 Airport and Heliport Localizer and Glide Slope Primary Records
type AirportLocGSPrimaryRecord struct {
	AirportRecord                    `fixed:"1,13"`
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
	AirportRecord            `fixed:"1,13"`
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
	AirportRecord                `fixed:"1,13"`
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

// IsMissedApproachPoint returns true if the record is for the missed approach
// point on an approach procedure.
func (p *AirportProcedurePrimaryRecord) IsMissedApproachPoint() bool {
	d := p.WaypointDescriptionCode
	if len(d) >= 4 && d[3] == 'M' {
		return true
	}
	return false
}
