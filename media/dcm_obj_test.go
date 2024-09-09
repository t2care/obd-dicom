package media

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
)

func TestNewDCMObjFromFile(t *testing.T) {
	InitDict()

	type args struct {
		fileName string
	}
	tests := []struct {
		name          string
		args          args
		wantTagsCount int
		wantErr       bool
	}{
		{
			name:          "Should load DICOM file from bugged DICOM written by us",
			args:          args{fileName: "../samples/test2-2.dcm"},
			wantTagsCount: 116,
			wantErr:       false,
		},
		{
			name:          "Should load DICOM file from post bugged DICOM written by us",
			args:          args{fileName: "../samples/test2-3.dcm"},
			wantTagsCount: 116,
			wantErr:       false,
		},
		{
			name:          "Should load DICOM file",
			args:          args{fileName: "../samples/test2.dcm"},
			wantTagsCount: 116,
			wantErr:       false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dcmObj, err := NewDCMObjFromFile(tt.args.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewDCMObjFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(dcmObj.GetTags()) != tt.wantTagsCount {
				t.Errorf("NewDCMObjFromFile() count = %v, wantTagsCount %v", len(dcmObj.GetTags()), tt.wantTagsCount)
				return
			}
		})
	}
}

func TestChangeTransferSynx(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
	}{
		{
			name:     "Should change transfer synxtax from ExplicitVRLittleEndian",
			fileName: "../samples/test2.dcm",
		},
	}
	for _, tt := range tests {
		for _, ts := range transfersyntax.SupportedTransferSyntaxes {
			assert.NoError(t, changeSyntax(tt.fileName, ts), fmt.Sprintf("%s to %s", tt.name, ts.Name))
		}
	}
}

func changeSyntax(filename string, ts *transfersyntax.TransferSyntax) (err error) {
	dcmObj, err := NewDCMObjFromFile(filename)
	if err != nil {
		return
	}
	if err = dcmObj.ChangeTransferSynx(ts); err != nil {
		return
	}
	out := "tmp"
	if err = dcmObj.WriteToFile(out); err != nil {
		return
	}
	return dcmtk_dump(out)
}

func dcmtk_dump(dcm string) error {
	if out, err := exec.Command("dcmdump", dcm).CombinedOutput(); err != nil {
		return fmt.Errorf("%s", string(out))
	}
	return nil
}

func BenchmarkOBD(b *testing.B) {
	for i := 0; i < b.N; i++ {
		NewDCMObjFromFile("../samples/test.dcm")
	}
}

func TestParseOptions(t *testing.T) {
	tests := []struct {
		name         string
		opt          *ParseOptions
		tagCount     int
		protocolName string
	}{
		{
			name:         "No options",
			opt:          &ParseOptions{},
			tagCount:     99,
			protocolName: "SAG T1 ACR",
		},
		{
			name:         "Skip pixel",
			opt:          &ParseOptions{SkipPixelData: true},
			tagCount:     98,
			protocolName: "SAG T1 ACR",
		},
		{
			name:         "Only meta header",
			opt:          &ParseOptions{OnlyMetaHeader: true},
			tagCount:     0,
			protocolName: "",
		},
		{
			name:         "Until patient tags",
			opt:          &ParseOptions{UntilPatientTag: true},
			tagCount:     30,
			protocolName: "",
		},
		{
			name:         "Skip FillTag",
			opt:          &ParseOptions{SkipPixelData: true, SkipFillTag: true},
			tagCount:     98,
			protocolName: "SAG T1 ACR",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, _ := NewDCMObjFromFile("../samples/test.dcm", tt.opt)
			if len(o.GetTags()) != tt.tagCount {
				t.Errorf("TestParseOptions() count = %v, want %v", len(o.GetTags()), tt.tagCount)
			}
			if pn := o.GetString(tags.ProtocolName); pn != tt.protocolName {
				t.Errorf("TestParseOptions() syntax = %v, want %v", pn, tt.protocolName)
			}
		})
	}
}

func TestGetString(t *testing.T) {
	tests := []struct {
		name  string
		tag   *tags.Tag
		value string
	}{
		{
			name:  "Get patient name",
			tag:   tags.PatientName,
			value: "ACR PHANTOM",
		},
		{
			name:  "Get SeriesNumber",
			tag:   tags.SeriesNumber,
			value: "301",
		},
		{
			name:  "Get AITDeviceType",
			tag:   tags.AITDeviceType,
			value: "",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, _ := NewDCMObjFromFile("../samples/test.dcm", &ParseOptions{SkipPixelData: true})
			if pn := o.GetString(tt.tag); pn != tt.value {
				t.Errorf("TestGetString() get = %v, want %v", pn, tt.value)
			}
		})
	}
}

func TestGetUShort(t *testing.T) {
	tests := []struct {
		name  string
		tag   *tags.Tag
		value uint16
	}{
		{
			name:  "Get SamplesPerPixel",
			tag:   tags.SamplesPerPixel,
			value: 1,
		},
		{
			name:  "Get Rows",
			tag:   tags.Rows,
			value: 256,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			o, _ := NewDCMObjFromFile("../samples/test.dcm", &ParseOptions{SkipPixelData: true})
			if pn := o.GetUShort(tt.tag); pn != tt.value {
				t.Errorf("TestGetUInt() get = %v, want %v", pn, tt.value)
			}
		})
	}
}
