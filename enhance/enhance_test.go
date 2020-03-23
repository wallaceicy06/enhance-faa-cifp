package enhance

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"testing"

	"github.com/google/go-cmp/cmp"
	geo "github.com/kellydunn/golang-geo"
)

const (
	testDataFile    = "test_data.txt"
	testDataOutFile = "test_data_out.txt"
)

type badReader struct {
	io.Reader
}

func (*badReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("problem reading")
}

func TestProcessBadReader(t *testing.T) {
	for _, tt := range []struct {
		name    string
		in      io.Reader
		want    string
		wantErr bool
	}{
		{
			name: "Good",
			in:   bytes.NewReader([]byte("SUSAP KHWDK2CBOGRE K20    W     N37372195W122023769                       E0133     NAR           BOGRE                    107992002")),
			want: "SUSAP KHWDK2CBOGRE K20    W     N37372195W122023769                       E0133     NAR           BOGRE                    107992002\n",
		},
		{
			name:    "BadReader",
			in:      &badReader{},
			wantErr: true,
		},
		{
			name:    "BadData",
			in:      bytes.NewReader([]byte("SUSAP KHWDK2CBOGRE K20    W     NBADDATA!W122023769                       E0133     NAR           BOGRE                    107992002")),
			wantErr: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var out bytes.Buffer
			err := Process(tt.in, &out)
			if tt.wantErr {
				if err == nil {
					t.Fatal("Process() = <nil> want <non-nil>")
				}
				return
			}
			if err != nil {
				t.Fatalf("Process() = %v want <nil>", err)
			}
			if !bytes.Equal(out.Bytes(), []byte(tt.want)) {
				t.Errorf("Process() out = %q want %q", out, tt.want)
			}
		})
	}
}

type badWriter struct {
	io.Writer
}

func (*badWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("problem writing")
}

func TestProcessBadWriter(t *testing.T) {
	in := bytes.NewReader([]byte("SUSAP KHWDK2CBOGRE K20    W     N37372195W122023769                       E0133     NAR           BOGRE                    107992002"))
	if err := Process(in, &badWriter{}); err == nil {
		t.Fatalf("Process() = <nil> want <non-nil>")
	}
}

func TestProcessRecord(t *testing.T) {
	for _, tt := range []struct {
		name          string
		processor     *processor
		record        string
		want          string
		wantProcessor *processor
		wantErr       bool
	}{
		{
			name:      "NDB",
			processor: newProcessor(),
			record:    "SCANDB       ILI   PA004110H  W N59000000W155000000                       E0140           NARILIAMNA                       004122002",
			want:      "SCANDB       ILI   PA004110H  W N59000000W155000000                       E0140           NARILIAMNA                       004122002\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{},
				OtherWaypoints: map[string]*geo.Point{
					"ILI": geo.NewPoint(59, -155),
				},
			},
		},
		{
			name:      "NDBBadLatLon",
			processor: newProcessor(),
			record:    "SCANDB       ILI   PA004110H  W NBAD00000W155000000                       E0140           NARILIAMNA                       004122002",
			wantErr:   true,
		},
		{
			name:      "SkipsNDBDME",
			processor: newProcessor(),
			record:    "SCAND        ADK   PA011400 DUW                    ADK N51521587W176402739E0070003291     NARMOUNT MOFFETT                 002361703",
			want:      "SCAND        ADK   PA011400 DUW                    ADK N51521587W176402739E0070003291     NARMOUNT MOFFETT                 002361703\n",
			wantProcessor: &processor{
				Airports:       map[string]*airportData{},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name:      "VOR",
			processor: newProcessor(),
			record:    "SUSAD        PYE   K2011370VDHW N38000000W122000000    N38044712W122520418E0170013402     NARPOINT REYES                   236192002",
			want:      "SUSAD        PYE   K2011370VDHW N38000000W122000000    N38044712W122520418E0170013402     NARPOINT REYES                   236192002\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{},
				OtherWaypoints: map[string]*geo.Point{
					"PYE": geo.NewPoint(38, -122),
				},
			},
		},
		{
			name:      "VORBadLatLon",
			processor: newProcessor(),
			record:    "SUSAD        PYE   K2011370VDHW NBAD00000W122000000    N38044712W122520418E0170013402     NARPOINT REYES                   236192002",
			wantErr:   true,
		},
		{
			name:      "EnrouteWaypoint",
			processor: newProcessor(),
			record:    "SUSAEAENRT   SUNOL K20    C  RL N37000000W121000000                       E0132     NAR           SUNOL                    459212002",
			want:      "SUSAEAENRT   SUNOL K20    C  RL N37000000W121000000                       E0132     NAR           SUNOL                    459212002\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{},
				OtherWaypoints: map[string]*geo.Point{
					"SUNOL": geo.NewPoint(37, -121),
				},
			},
		},
		{
			name:      "EnrouteWaypointBadLatLon",
			processor: newProcessor(),
			record:    "SUSAEAENRT   SUNOL K20    C  RL NBAD00000W121000000                       E0132     NAR           SUNOL                    459212002",
			wantErr:   true,
		},
		{
			name:      "NewAirport",
			processor: newProcessor(),
			record:    "SUSAP KHWDK2AHWD     0     056YHN37393214W122071825E015000052         1800018000C    MNAR    HAYWARD EXECUTIVE             107981608",
			want:      "SUSAP KHWDK2AHWD     0     056YHN37393214W122071825E015000052         1800018000C    MNAR    HAYWARD EXECUTIVE             107981608\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints:  map[string]*geo.Point{},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name: "TerminalWaypoint",
			processor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints:  map[string]*geo.Point{},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record: "SUSAP KHWDK2CSUDGE K20    W     N37000000W121000000                       E0132     NAR           SUDGE                    108112002",
			want:   "SUSAP KHWDK2CSUDGE K20    W     N37000000W121000000                       E0132     NAR           SUDGE                    108112002\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"SUDGE": geo.NewPoint(37, -121),
						},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name: "TerminalWaypointBadLatLon",
			processor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints:  map[string]*geo.Point{},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record:  "SUSAP KHWDK2CSUDGE K20    W     NBAD00000W121000000                       E0132     NAR           SUDGE                    108112002",
			wantErr: true,
		},
		{
			name: "ApproachProcedureLocNotFAF",
			processor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(38, -122),
						},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record: "SUSAP KHWDK2FL28L  ASJC   010SJC  K2D 0V  A    IF                                             18000                 0 DS   108481212",
			want:   "SUSAP KHWDK2FL28L  ASJC   010SJC  K2D 0V  A    IF                                             18000                 0 DS   108481212\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(38, -122),
						},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name: "ApproachProcedureLocFAF",
			processor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(38, -122),
						},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record: "SUSAP KHWDK2FL28L  L      020FERNEK2PC0E  F    CF IHWDK2      1079007428800053PI  + 02500                 OAK   K2D 0 DS   108521310",
			want:   "SUSAP KHWDK2FL28L  L      020FERNEK2PC0E  F    CF IHWDK2      1079007428800053PI  + 02500                 OAK   K2D 0 DS   108521310\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(38, -122),
						},
						Approaches: map[string]*locApchData{
							"L28L": &locApchData{
								FinalApproachFix: "FERNE",
								LocalizerID:      "IHWD",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name: "Localizer",
			processor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(37.59, -121.99),
						},
						Approaches: map[string]*locApchData{
							"L28L": &locApchData{
								FinalApproachFix: "FERNE",
								LocalizerID:      "IHWD",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record: "SUSAP KHWDK2IIHWD0   111150RW28LN37394620W1220746752879                   0109     0500   E0150                            108901212",
			want:   "SUSAP KHWDK2IIHWD0   111150RW28LN37394620W1220746752879                   0109     0500   E0150                            108901212\nSUSAP KHWDK2IIHWD0   2S                            30341N                                                                  108901212\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(37.59, -121.99),
						},
						Approaches: map[string]*locApchData{
							"L28L": &locApchData{
								FinalApproachFix: "FERNE",
								LocalizerID:      "IHWD",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name: "SkipLocalizerNoApproach",
			processor: &processor{
				Airports:       map[string]*airportData{},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record: "SUSAP KHWDK2IIHWD0   111150RW28LN37394620W1220746752879                   0109     0500   E0150                            108901212",
			want:   "SUSAP KHWDK2IIHWD0   111150RW28LN37394620W1220746752879                   0109     0500   E0150                            108901212\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints:  map[string]*geo.Point{},
						Approaches: map[string]*locApchData{},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name: "SkipLocalizerBadLatLon",
			processor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(37.59, -121.99),
						},
						Approaches: map[string]*locApchData{
							"L28L": &locApchData{
								FinalApproachFix: "FERNE",
								LocalizerID:      "IHWD",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record: "SUSAP KHWDK2IIHWD0   111150RW28LNBAD94620W1220746752879                   0109     0500   E0150                            108901212",
			want:   "SUSAP KHWDK2IIHWD0   111150RW28LNBAD94620W1220746752879                   0109     0500   E0150                            108901212\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KHWD": &airportData{
						Waypoints: map[string]*geo.Point{
							"FERNE": geo.NewPoint(37.59, -121.99),
						},
						Approaches: map[string]*locApchData{
							"L28L": &locApchData{
								FinalApproachFix: "FERNE",
								LocalizerID:      "IHWD",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
		{
			name: "LocalizerOtherWaypointFAF",
			processor: &processor{
				Airports: map[string]*airportData{
					"KSAC": &airportData{
						Approaches: map[string]*locApchData{
							"I02": &locApchData{
								FinalApproachFix: "SAC",
								LocalizerID:      "ISAC",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{
					"SAC": geo.NewPoint(38.44, -121.55),
				},
			},
			record: "SUSAP KSACK2IISAC1   111030RW02 N38311332W1212917310191N38302558W1212950951089 10860600300E01405700020                     973081402",
			want:   "SUSAP KSACK2IISAC1   111030RW02 N38311332W1212917310191N38302558W1212950951089 10860600300E01405700020                     973081402\nSUSAP KSACK2IISAC1   2S                            03105N                                                                  973081402\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KSAC": &airportData{
						Approaches: map[string]*locApchData{
							"I02": &locApchData{
								FinalApproachFix: "SAC",
								LocalizerID:      "ISAC",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{
					"SAC": geo.NewPoint(38.44, -121.55),
				},
			},
		},

		{
			name: "SkipLocalizerMissingOtherWaypointFAF",
			processor: &processor{
				Airports: map[string]*airportData{
					"KSAC": &airportData{
						Approaches: map[string]*locApchData{
							"I02": &locApchData{
								FinalApproachFix: "SAC",
								LocalizerID:      "ISAC",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
			record: "SUSAP KSACK2IISAC1   111030RW02 N38311332W1212917310191N38302558W1212950951089 10860600300E01405700020                     973081402",
			want:   "SUSAP KSACK2IISAC1   111030RW02 N38311332W1212917310191N38302558W1212950951089 10860600300E01405700020                     973081402\n",
			wantProcessor: &processor{
				Airports: map[string]*airportData{
					"KSAC": &airportData{
						Approaches: map[string]*locApchData{
							"I02": &locApchData{
								FinalApproachFix: "SAC",
								LocalizerID:      "ISAC",
							},
						},
					},
				},
				OtherWaypoints: map[string]*geo.Point{},
			},
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.processor.processRecord([]byte(tt.record))
			if tt.wantErr {
				if err == nil {
					t.Fatalf("processRecord(%q) = _, <nil> want <non-nil>", tt.record)
				}
				return
			}
			if err != nil {
				t.Fatalf("processRecord(%q) = _, %v want <nil>", tt.record, err)
			}
			if !bytes.Equal(got, []byte(tt.want)) {
				t.Errorf("processRecord(%q) = %q, _ want %q, _", tt.record, got, tt.want)
			}
			if diff := cmp.Diff(tt.wantProcessor, tt.processor, cmp.AllowUnexported(geo.Point{})); diff != "" {
				t.Errorf("processor had diffs (-got +want): %s", diff)
			}
		})
	}
}

func TestProcessCompleteFile(t *testing.T) {
	testData, err := ioutil.ReadFile(testDataFile)
	if err != nil {
		t.Fatalf("Could not read test data file: %v", err)
	}
	testDataOut, err := ioutil.ReadFile(testDataOutFile)
	if err != nil {
		t.Fatalf("Could not read test data file: %v", err)
	}

	in := bytes.NewReader(testData)
	var out bytes.Buffer
	if err := Process(in, &out); err != nil {
		t.Fatalf("Process() = %v want <nil>", err)
	}
	if !bytes.Equal(out.Bytes(), testDataOut) {
		t.Errorf("Process() out content not as expected")
	}
}
