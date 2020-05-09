package arinc

import (
	"math"
	"testing"
)

func TestLatLon(t *testing.T) {
	const tolerance = 0.0001
	for _, tt := range []struct {
		name    string
		latStr  string
		lonStr  string
		wantLat float64
		wantLon float64
		wantErr bool
	}{
		{
			name:    "GoodNE",
			latStr:  "N39513881",
			lonStr:  "E104450794",
			wantLat: 39.860781,
			wantLon: 104.752206,
		},
		{
			name:    "GoodNW",
			latStr:  "N39513881",
			lonStr:  "W104450794",
			wantLat: 39.860781,
			wantLon: -104.752206,
		},
		{
			name:    "GoodSE",
			latStr:  "S39513881",
			lonStr:  "E104450794",
			wantLat: -39.860781,
			wantLon: 104.752206,
		},
		{
			name:    "GoodSW",
			latStr:  "S39513881",
			lonStr:  "W104450794",
			wantLat: -39.860781,
			wantLon: -104.752206,
		},
		{
			name:    "InvalidLatLen",
			latStr:  "881",
			lonStr:  "W104450794",
			wantErr: true,
		},
		{
			name:    "InvalidLatDeg",
			latStr:  "N3F513881",
			lonStr:  "W104450794",
			wantErr: true,
		},
		{
			name:    "InvalidLatMin",
			latStr:  "N39F13881",
			lonStr:  "W104450794",
			wantErr: true,
		},
		{
			name:    "InvalidLatSec",
			latStr:  "N395138F1",
			lonStr:  "W104450794",
			wantErr: true,
		},
		{
			name:    "InvalidLonLen",
			latStr:  "S39513881",
			lonStr:  "W1044507945",
			wantErr: true,
		},
		{
			name:    "InvalidLonDeg",
			latStr:  "S39513881",
			lonStr:  "W10F450794",
			wantErr: true,
		},
		{
			name:    "InvalidLonLen",
			latStr:  "S39513881",
			lonStr:  "W104F50794",
			wantErr: true,
		},
		{
			name:    "InvalidLonSec",
			latStr:  "S39513881",
			lonStr:  "W104450F945",
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			gotLat, gotLon, err := LatLon(tt.latStr, tt.lonStr)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("LatLon(%q, %q) = _, _, <nil> want _, _, <non-nil>", tt.latStr, tt.lonStr)
				}
				return
			}
			if err != nil {
				t.Fatalf("LatLon(%q, %q) = _, _, %v want _, _, <nil>", tt.latStr, tt.lonStr, err)
			}
			if diff := math.Abs(gotLat - tt.wantLat); diff > tolerance {
				t.Errorf("latitude = %f want %f", gotLat, tt.wantLat)
			}
			if diff := math.Abs(gotLon - tt.wantLon); diff > tolerance {
				t.Errorf("longitude = %f want %f", gotLon, tt.wantLon)
			}
		})
	}
}

func TestEncodeBearing(t *testing.T) {
	for _, tt := range []struct {
		name    string
		bearing float64
		want    string
	}{
		{
			name:    "Simple",
			bearing: 123.45,
			want:    "12345",
		},
		{
			name:    "LongDecimal",
			bearing: 123.45678,
			want:    "12346",
		},
		{
			name:    "LessThan100Degrees",
			bearing: 23.45678,
			want:    "02346",
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := EncodeBearing(tt.bearing); got != tt.want {
				t.Errorf("EncodeBearing(%f) = %q want %q", tt.bearing, got, tt.want)
			}
		})
	}
}

func TestParseMagneticVar(t *testing.T) {
	for _, tt := range []struct {
		name     string
		magVar   string
		want     float64
		wantTrue bool
		wantErr  bool
	}{
		{
			name:   "West",
			magVar: "W0140",
			want:   14.0,
		},
		{
			name:   "East",
			magVar: "E0135",
			want:   -13.5,
		},
		{
			name:     "True",
			magVar:   "T0000",
			want:     0,
			wantTrue: true,
		},
		{
			name:    "InvalidDirection",
			magVar:  "Y0140",
			wantErr: true,
		},
		{
			name:    "InvalidLength",
			magVar:  "Y",
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, isTrue, err := ParseMagneticVar(tt.magVar)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("ParseMagneticVar(%q) = _, _, <nil> want _, _, <non-nil>", tt.magVar)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseMagneticVar(%q) = _, _, %v want _, _, <nil>", tt.magVar, err)
			}
			if isTrue != tt.wantTrue {
				t.Errorf("ParseMagneticVar(%q) = _, %t, _ want _, %t, _", tt.magVar, isTrue, tt.wantTrue)
			}
			if got != tt.want {
				t.Errorf("ParseMagneticVar(%q) = %f, _, _ want %f, _, _", tt.magVar, got, tt.want)
			}
		})
	}
}

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
			got := tt.record.IsFinalApproachFix()
			if got != tt.want {
				t.Errorf("IsLocalizerFrontCourseApproach() = %t want %t", got, tt.want)
			}
		})
	}
}
