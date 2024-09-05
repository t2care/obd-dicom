//go:build jpeg2000

package transcoder

import (
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media/transcoder/openjpeg"
)

func init() {
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000Lossless.UID, openjpeg.J2Kdecode, openjpeg.J2Kencode)
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000.UID, openjpeg.J2Kdecode, openjpeg.J2Kencode)
}
