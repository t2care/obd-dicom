//go:build jpeg2000

package media

import (
	"testing"

	"github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"
)

func Test_dcmObj_jpeg2000_ChangeTransferSynx(t *testing.T) {
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
			name:     "Should change transfer synxtax to JPEG2000Lossless",
			fileName: "../samples/jpeg8.dcm",
			args:     args{transfersyntax.JPEG2000Lossless},
			wantErr:  false,
		},
		{
			name:     "Should change transfer synxtax to JPEG2000",
			fileName: "../samples/test2.dcm",
			args:     args{transfersyntax.JPEG2000},
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
