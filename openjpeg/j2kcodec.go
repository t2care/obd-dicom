//go:build jpeg2000

package openjpeg

import "github.com/one-byte-data/obd-dicom/dictionary/transfersyntax"

func init() {
	transfersyntax.Register(transfersyntax.JPEG2000Lossless, J2Kdecode, J2Kencode)
	transfersyntax.Register(transfersyntax.JPEG2000, J2Kdecode, J2Kencode)
}
