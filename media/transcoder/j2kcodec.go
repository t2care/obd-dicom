//go:build jpeg2000

package transcoder

import (
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media/transcoder/openjpeg"
)

func init() {
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000Lossless.UID, decode, encode)
	transfersyntax.RegisterCodec(transfersyntax.JPEG2000.UID, decode, encode)
}

func decode(frame uint32, bitsa uint16, j2kData []byte, j2kSize uint32, outputData []byte, outputSize uint32) error {
	offset := frame * outputSize
	return openjpeg.J2Kdecode(j2kData, j2kSize, outputData[offset:])
}

func encode(frame uint32, RGB bool, img []byte, cols uint16, rows uint16, samples uint16, bitsa uint16, JPEGData *[]byte, JPEGBytes *int, ratio int) error {
	offset := frame * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	if RGB {
		offset = 3 * offset
		return openjpeg.J2Kencode(img[offset:], cols, rows, 3, bitsa, JPEGData, JPEGBytes, ratio)
	} else {
		return openjpeg.J2Kencode(img[offset:], cols, rows, 1, bitsa, JPEGData, JPEGBytes, ratio)
	}
}
