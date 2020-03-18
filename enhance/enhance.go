package enhance

import (
	"bufio"
	"bytes"
	"fmt"
	"io"

	"github.com/wallaceicy06/enhance-faa-cifp/arinc"
	fixedwidth "github.com/wallaceicy06/go-fixedwidth"
)

// Process reads ARINC data from in and writes the modified data to out. All
// localizers in the input data will be augmented with an extension field that
// includes a more accurate bearing for the localizer. This bearing is computed
// as the course of the leg from the final approach fix to the missed approach
// point.
func Process(in io.Reader, out io.Writer) error {
	s := bufio.NewScanner(in)
	for s.Scan() {
		p := arinc.Record{}
		dec := fixedwidth.NewDecoder(bytes.NewReader(s.Bytes()))
		dec.SetTrimSpace(false)
		if err := dec.Decode(&p); err != nil {
			return fmt.Errorf("problem unmarshalling data: %v", err)
		}
		if p.SectionCode == "P" {
			a := arinc.AirportRecord{}
			if err := fixedwidth.Unmarshal(s.Bytes(), &a); err != nil {
				return fmt.Errorf("problem unmarshalling data: %v", err)
			}
			if a.SubsectionCode == "I" || a.SubsectionCode == "L" {
				loc := arinc.AirportLocGSPrimaryRecord{}
				if err := fixedwidth.Unmarshal(s.Bytes(), &loc); err != nil {
					return fmt.Errorf("problem unmarshalling data: %v", err)
				}
			}
		}
		toWrite, err := fixedwidth.Marshal(p)
		if err != nil {
			return fmt.Errorf("could not marshal record: %v", err)
		}
		fmt.Fprintf(out, "%s\n", toWrite)
	}
	if err := s.Err(); err != nil {
		return fmt.Errorf("problem parsing data: %v", err)
	}
	return nil
}
