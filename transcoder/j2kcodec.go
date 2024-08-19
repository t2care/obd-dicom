//go:build jpeg2000

package transcoder

import (
	"github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"
	"github.com/one-byte-data/obd-dicom/openjpeg"
)

func init() {
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000Lossless.UID, openjpeg.J2Kdecode, openjpeg.J2Kencode)
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000.UID, openjpeg.J2Kdecode, openjpeg.J2Kencode)
}
