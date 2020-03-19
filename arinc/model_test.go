package arinc

import "testing"

func TestIsLocalizerFrontCourseApproach(t *testing.T) {
	for _, tt := range []struct {
		name   string
		record *AirportProcedurePrimaryRecord
		want   bool
	}{
		{
			name:   "ILS",
			record: &AirportProcedurePrimaryRecord{ProcedureID: "I28R"},
			want:   true,
		},
		{
			name:   "LocalizerFrontCourse",
			record: &AirportProcedurePrimaryRecord{ProcedureID: "L28L"},
			want:   true,
		},
		{
			name:   "LDA",
			record: &AirportProcedurePrimaryRecord{ProcedureID: "X19R"},
			want:   true,
		},
		{
			name:   "SDF",
			record: &AirportProcedurePrimaryRecord{ProcedureID: "U05"},
			want:   true,
		},
		{
			name:   "LocalizerBackCourse",
			record: &AirportProcedurePrimaryRecord{ProcedureID: "B28L"},
			want:   false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.record.IsLocalizerFrontCourseApproach()
			if got != tt.want {
				t.Errorf("IsLocalizerFrontCourseApproach() = %t want %t", got, tt.want)
			}
		})
	}
}

func TestIsFinalApproachFix(t *testing.T) {
	for _, tt := range []struct {
		name   string
		record *AirportProcedurePrimaryRecord
		want   bool
	}{
		{
			name:   "Good",
			record: &AirportProcedurePrimaryRecord{WaypointDescriptionCode: "   F"},
			want:   true,
		},
		{
			name:   "NotFinalApproachFix",
			record: &AirportProcedurePrimaryRecord{WaypointDescriptionCode: "   M"},
			want:   false,
		},
		{
			name:   "InvalidData",
			record: &AirportProcedurePrimaryRecord{WaypointDescriptionCode: ""},
			want:   false,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := tt.record.IsFinalApproachFix()
			if got != tt.want {
				t.Errorf("IsLocalizerFrontCourseApproach() = %t want %t", got, tt.want)
			}
		})
	}
}
