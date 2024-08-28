package media

import (
	"testing"

	"github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"
)

func TestTranscodeSeq(t *testing.T) {
	tests := []struct {
		name    string
		inTS    *transfersyntax.TransferSyntax
		outTS   *transfersyntax.TransferSyntax
		wantErr bool
	}{
		{
			name:    "Should transcode seq from ImplicitVRLittleEndian to ExplicitVRLittleEndian",
			inTS:    transfersyntax.ImplicitVRLittleEndian,
			outTS:   transfersyntax.ExplicitVRLittleEndian,
			wantErr: false,
		},
		{
			name:    "Should transcode seq from ExplicitVRLittleEndian to ImplicitVRLittleEndian",
			inTS:    transfersyntax.ExplicitVRLittleEndian,
			outTS:   transfersyntax.ImplicitVRLittleEndian,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := transcodeSeq(tt.inTS, tt.outTS); (err != nil) != tt.wantErr {
				t.Errorf("transcodeSeq() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func transcodeSeq(fromTS, toTS *transfersyntax.TransferSyntax) error {
	o := NewEmptyDCMObj()
	o.SetTransferSyntax(fromTS)
	o.AddConceptNameSeq(0x40, 0xA043, "123", "Test")

	if err := o.ChangeTransferSynx(toTS); err != nil {
		return err
	}
	for _, tag := range o.GetTags() {
		seq, err := tag.ReadSeq(o.IsExplicitVR())
		if err != nil {
			return err
		}
		for _, item := range seq.GetTags() {
			_, err := item.ReadSeq(o.IsExplicitVR())
			if err != nil {
				return err
			}
		}
	}
	return nil
}
