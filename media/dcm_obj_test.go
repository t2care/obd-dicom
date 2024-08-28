package media

import (
	"testing"

	"github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"
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

func Test_dcmObj_ChangeTransferSynx(t *testing.T) {
	type args struct {
		outTS *transfersyntax.TransferSyntax
	}
	tests := []struct {
		name     string
		fileName string
		args     args
		wantErr  bool
	}{
		{
			name:     "Should change transfer synxtax to ImplicitVRLittleEndian",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.ImplicitVRLittleEndian},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax to ExplicitVRLittleEndian",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.ExplicitVRLittleEndian},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax to ExplicitVRBigEndian",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.ExplicitVRBigEndian},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax to RLELossless",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.RLELossless},
			wantErr:  true,
		},
		{
			name:     "Should change transfer synxtax to JPEGLosslessSV1",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.JPEGLosslessSV1},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax to JPEGBaseline8Bit",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.JPEGBaseline8Bit},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax to JPEGExtended12Bit",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.JPEGExtended12Bit},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax from JPEGLosslessSV1",
			fileName: "../samples/test-losslessSV1.dcm",
			args:     args{transfersyntax.ExplicitVRLittleEndian},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax from JPEGLosslessSV1 to ImplicitVRLittleEndian",
			fileName: "../samples/test-losslessSV1.dcm",
			args:     args{transfersyntax.ImplicitVRLittleEndian},
			wantErr:  false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := changeSyntax(tt.fileName, tt.args.outTS); (err != nil) != tt.wantErr {
				t.Errorf("changeSyntax() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
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
	if err = dcmObj.DumpTags(); err != nil {
		return
	}
	_, err = NewDCMObjFromBytes(dcmObj.WriteToBytes())
	return
}
