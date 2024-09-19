package utils

import (
	"reflect"
	"testing"

	"github.com/t2care/obd-dicom/media"
)

func TestMapDicomDataToStruct(t *testing.T) {
	obj, _ := media.NewDCMObjFromFile("../samples/test.dcm")
	type instance struct {
		BitsAllocated uint8 `dicom:"0028,0100"`
	}
	type series struct {
		SeriesNumber string `dicom:"0020,0011"`
		Instance     []instance
	}
	type study struct {
		PatientName string `dicom:"0010,0010"`
		Series      []series
	}
	type dicomweb struct {
		PatientName string `json:"0010,0010"`
	}
	type args struct {
		dicomDataset *media.DcmObj
		targetStruct any
	}
	tests := []struct {
		name       string
		args       args
		wantErr    bool
		wantTarget any
		keywork    string
	}{
		{
			name:       "Parse error",
			args:       args{obj, study{}},
			wantErr:    true,
			wantTarget: study{},
		},
		{
			name:    "Parse structure",
			args:    args{obj, &study{}},
			wantErr: false,
			wantTarget: &study{
				PatientName: "ACR PHANTOM",
				Series: []series{{
					SeriesNumber: "301",
					Instance:     []instance{{BitsAllocated: 16}},
				}},
			},
		},
		{
			name:       "Parse with other keywork",
			args:       args{obj, &dicomweb{}},
			keywork:    "json",
			wantErr:    false,
			wantTarget: &dicomweb{PatientName: "ACR PHANTOM"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var err error
			if tt.keywork != "" {
				err = MapDicomDataToStruct(tt.args.dicomDataset, tt.args.targetStruct, tt.keywork)
			} else {
				err = MapDicomDataToStruct(tt.args.dicomDataset, tt.args.targetStruct)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("MapDicomDataToStruct() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.targetStruct, tt.wantTarget) {
				t.Errorf("MapDicomDataToStruct() = %v, want %v", tt.args.targetStruct, tt.wantTarget)
			}
		})
	}
}
