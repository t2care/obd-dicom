//go:build jpeg2000

package openjpeg

import "github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"

func init() {
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000Lossless.UID, J2Kdecode, J2Kencode)
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000.UID, J2Kdecode, J2Kencode)
}
