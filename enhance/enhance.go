package enhance

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"log"

	fixedwidth "github.com/ianlopshire/go-fixedwidth"
	geo "github.com/kellydunn/golang-geo"
	"github.com/wallaceicy06/enhance-faa-cifp/arinc"
)

type airportData struct {
	Waypoints  map[string]*geo.Point
	Approaches map[string]*locApchData
}

// ApproachForLoc returns the first approach that specifies the given localizer
// ID as the recommended navaid. If there is no corresponding approach, nil is
// returned.
func (a *airportData) ApproachForLoc(LocalizerID string) *locApchData {
	for _, apch := range a.Approaches {
		if apch.LocalizerID == LocalizerID {
			return apch
		}
	}
	return nil
}

type locApchData struct {
	FinalApproachFix string
	LocalizerID      string
}

type Option func(p *processor)

// RemoveDuplicateLocalizers is an option that enables or disables removal of
// duplicate localizers in the data. If enabled, duplicate localizers that
// are specified as an LDA with or without glideslope will be removed from
// the output data. (e.g. KVNY includes a duplicate IBUR localizer for the
// LDA-C approach)
func RemoveDuplicateLocalizers(enabled bool) Option {
	return func(p *processor) {
		p.RemoveDuplicateLocalizers = enabled
	}
}

// Process reads ARINC data from in and writes the modified data to out. All
// localizers in the input data will be augmented with an extension field that
// includes a more accurate bearing for the localizer. This bearing is computed
// as the course of the leg from the final approach fix to the missed approach
// point.
func Process(in io.ReadSeeker, out io.Writer, opts ...Option) error {
	p := newProcessor(opts...)

	// If duplicate localizer removal is enabled, the data must be pre-processed
	// to collect all localizers.
	if p.RemoveDuplicateLocalizers {
		s := bufio.NewScanner(in)
		for s.Scan() {
			p.preProcess(s.Bytes())
		}
		if err := s.Err(); err != nil {
			return fmt.Errorf("problem parsing data: %v", err)
		}
	}

	if _, err := in.Seek(0, io.SeekStart); err != nil {
		return fmt.Errorf("could not seek to start of file: %v", err)
	}
	s := bufio.NewScanner(in)
	for s.Scan() {
		processed, err := p.processRecord(s.Bytes())
		if err != nil {
			return fmt.Errorf("could not process record: %v", err)
		}
		if _, err := out.Write(processed); err != nil {
			return fmt.Errorf("could not write processed data: %v", err)
		}
	}

	if err := s.Err(); err != nil {
		return fmt.Errorf("problem parsing data: %v", err)
	}
	return nil
}

type processor struct {
	Airports                  map[string]*airportData
	OtherWaypoints            map[string]*geo.Point
	DuplicateLocalizers       map[string]bool
	RemoveDuplicateLocalizers bool
}

func newProcessor(options ...Option) *processor {
	p := &processor{
		Airports:            make(map[string]*airportData),
		OtherWaypoints:      make(map[string]*geo.Point),
		DuplicateLocalizers: make(map[string]bool),
	}
	for _, o := range options {
		o(p)
	}
	return p
}

func (p *processor) preProcess(recordBytes []byte) error {
	r := arinc.Record{}
	dec := fixedwidth.NewDecoder(bytes.NewReader(recordBytes))
	if err := dec.Decode(&r); err != nil {
		return fmt.Errorf("problem unmarshalling data: %v", err)
	}
	if r.SectionCode == arinc.SectionCodeAirport {
		a := arinc.AirportEnrouteRecord{}
		if err := fixedwidth.Unmarshal(recordBytes, &a); err != nil {
			return fmt.Errorf("problem unmarshalling airport: %v", err)
		}
		if r.SectionCode == arinc.SectionCodeAirport && a.SubsectionCode == arinc.SubsectionCodeLocGS {
			loc := arinc.AirportLocGSPrimaryRecord{}
			if err := fixedwidth.Unmarshal(recordBytes, &loc); err != nil {
				return fmt.Errorf("problem unmarshalling data: %v", err)
			}

			if _, ok := p.DuplicateLocalizers[loc.LocalizerID]; ok {
				p.DuplicateLocalizers[loc.LocalizerID] = true
			} else {
				p.DuplicateLocalizers[loc.LocalizerID] = false
			}
		}
	}
	return nil
}

func (p *processor) processRecord(recordBytes []byte) ([]byte, error) {
	writeRecord := func(buf *bytes.Buffer, records ...interface{}) ([]byte, error) {
		for _, record := range records {
			toWrite, err := fixedwidth.Marshal(record)
			if err != nil {
				return nil, fmt.Errorf("could not marshal record: %v", err)
			}
			fmt.Fprintf(buf, "%s\n", toWrite)
		}
		return buf.Bytes(), nil
	}
	out := &bytes.Buffer{}
	r := arinc.Record{}
	dec := fixedwidth.NewDecoder(bytes.NewReader(recordBytes))
	if err := dec.Decode(&r); err != nil {
		return nil, fmt.Errorf("problem unmarshalling data: %v", err)
	}

	if r.SectionCode == arinc.SectionCodeNavaid {
		switch r.SubsectionCode {
		case arinc.SubsectionCodeNavaidNDB:
			n := arinc.NDBNavaidRecord{}
			if err := fixedwidth.Unmarshal(recordBytes, &n); err != nil {
				return nil, fmt.Errorf("problem unmarshalling NDB: %v", err)
			}
			lat, lon, err := arinc.LatLon(n.NDBLatitude, n.NDBLongitude)
			if err != nil {
				return nil, fmt.Errorf("problem converting NDB latitude/longitude: %v", err)
			}
			p.OtherWaypoints[n.NDBID] = geo.NewPoint(lat, lon)
		case arinc.SubsectionCodeNavaidVHF:
			n := arinc.VHFNavaidRecord{}
			if err := fixedwidth.Unmarshal(recordBytes, &n); err != nil {
				return nil, fmt.Errorf("problem unmarshalling VOR: %v", err)
			}
			// Skip NDB/DME or DME with no corresponding VOR.
			if n.VORLatitude == "" || n.VORLongitude == "" {
				break
			}
			lat, lon, err := arinc.LatLon(n.VORLatitude, n.VORLongitude)
			if err != nil {
				return nil, fmt.Errorf("problem converting VOR %q latitude/longitude: %v", n.VORID, err)
			}
			p.OtherWaypoints[n.VORID] = geo.NewPoint(lat, lon)
		}
	} else if r.SectionCode == arinc.SectionCodeAirport || r.SectionCode == arinc.SectionCodeEnroute {
		a := arinc.AirportEnrouteRecord{}
		if err := fixedwidth.Unmarshal(recordBytes, &a); err != nil {
			return nil, fmt.Errorf("problem unmarshalling airport: %v", err)
		}
		if r.SectionCode == arinc.SectionCodeEnroute && r.SubsectionCode == arinc.SubsectionCodeEnrouteWaypoint {
			wpt := arinc.WaypointPrimaryRecord{}
			if err := fixedwidth.Unmarshal(recordBytes, &wpt); err != nil {
				return nil, fmt.Errorf("problem unmarshalling waypoint: %v", err)
			}
			lat, lon, err := arinc.LatLon(wpt.WaypointLatitude, wpt.WaypointLongitude)
			if err != nil {
				return nil, fmt.Errorf("problem converting waypoint latitude/longitude: %v", err)
			}
			p.OtherWaypoints[wpt.WaypointID] = geo.NewPoint(lat, lon)
		}
		if r.SectionCode == arinc.SectionCodeAirport {
			if _, ok := p.Airports[a.AirportID]; !ok {
				p.Airports[a.AirportID] = &airportData{
					Waypoints:  make(map[string]*geo.Point),
					Approaches: make(map[string]*locApchData),
				}
			}
			if a.SubsectionCode == arinc.SubsectionCodeTerminalWaypoint {
				wpt := arinc.WaypointPrimaryRecord{}
				if err := fixedwidth.Unmarshal(recordBytes, &wpt); err != nil {
					return nil, fmt.Errorf("problem unmarshalling waypoint: %v", err)
				}
				lat, lon, err := arinc.LatLon(wpt.WaypointLatitude, wpt.WaypointLongitude)
				if err != nil {
					return nil, fmt.Errorf("problem converting waypoint latitude/longitude: %v", err)
				}
				p.Airports[wpt.AirportID].Waypoints[wpt.WaypointID] = geo.NewPoint(lat, lon)
			}
			if a.SubsectionCode == arinc.SubsectionCodeApproachProcedure {
				apch := arinc.AirportProcedurePrimaryRecord{}
				if err := fixedwidth.Unmarshal(recordBytes, &apch); err != nil {
					return nil, fmt.Errorf("problem unmarshalling procedure: %v", err)
				}
				if apch.IsLocalizerFrontCourseApproach() {
					if apch.IsFinalApproachFix() {
						lc, ok := p.Airports[apch.AirportID].Approaches[apch.ProcedureID]
						if !ok {
							lc = &locApchData{}
							p.Airports[apch.AirportID].Approaches[apch.ProcedureID] = lc
						}
						lc.LocalizerID = apch.RecommendedNavaid
						lc.FinalApproachFix = apch.FixID
					}
				}
			}
			if a.SubsectionCode == arinc.SubsectionCodeLocGS {
				loc := arinc.AirportLocGSPrimaryRecord{}
				if err := fixedwidth.Unmarshal(recordBytes, &loc); err != nil {
					return nil, fmt.Errorf("problem unmarshalling data: %v", err)
				}

				if dup, ok := p.DuplicateLocalizers[loc.LocalizerID]; ok && dup {
					if loc.ILSCategory == "A" || loc.ILSCategory == "L" {
						log.Printf("Skipping duplicate localizer LDA facility: %q at %q", loc.LocalizerID, loc.AirportID)
						return nil, nil
					}
				}
				contRecord, err := p.processLocalizer(&loc)
				if err != nil {
					log.Printf("Skipping localizer %q at %q: %v", loc.LocalizerID, loc.AirportID, err)
					return writeRecord(out, r)
				}
				// There is some bug in the fixedwidth parser that causes these fields to not be parsed properly.
				// This is quick fix for the interim.
				contRecord.Data = loc.Data

				return writeRecord(out, loc, contRecord)
			}
		}
	}
	return writeRecord(out, r)
}

func (p *processor) processLocalizer(loc *arinc.AirportLocGSPrimaryRecord) (*arinc.AirportLocGSSimContinuationRecord, error) {
	a, ok := p.Airports[loc.AirportID]
	if !ok {
		// This case is pretty much impossible because the airport should have been added before this is called.
		return nil, fmt.Errorf("found localizer %q without corresponding airport %q", loc.LocalizerID, loc.AirportID)
	}
	apch := a.ApproachForLoc(loc.LocalizerID)
	if apch == nil {
		return nil, fmt.Errorf("could not find corresponding approach for localizer %q", loc.LocalizerID)
	}
	fapWaypoint, ok := a.Waypoints[apch.FinalApproachFix]
	if !ok {
		eWpt, ok := p.OtherWaypoints[apch.FinalApproachFix]
		if !ok {
			return nil, fmt.Errorf("could not find corresponding waypoint for final approach fix %q,", apch.FinalApproachFix)
		}
		fapWaypoint = eWpt
	}
	lat, lon, err := arinc.LatLon(loc.LocalizerLatitude, loc.LocalizerLongitude)
	if err != nil {
		return nil, fmt.Errorf("could not calculate latitude/longitude for localizer %q: %v", loc.LocalizerID, err)
	}
	locPosition := geo.NewPoint(lat, lon)
	bearing := fapWaypoint.BearingTo(locPosition)
	// This corects a bug in the golang-geo library that causes negative bearings.
	if bearing < 0 {
		bearing = 360 + bearing
	}

	loc.ContinuationRecordNumber = "1"
	contRecord := &arinc.AirportLocGSSimContinuationRecord{
		AirportEnrouteRecord:     loc.AirportEnrouteRecord,
		LocalizerID:              loc.LocalizerID,
		ILSCategory:              loc.ILSCategory,
		ContinuationRecordNumber: "2",
		ApplicationType:          arinc.ContinuationRecordSimulation,
		LocalizerTrueBearing:     arinc.EncodeBearing(bearing),
		LocalizerBearingSource:   arinc.LocalizerBearingSourceNotGovt,
	}
	return contRecord, nil
}
