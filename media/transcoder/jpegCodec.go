//go:build cgo && jpeg

package transcoder

import (
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media/transcoder/jpeglib"
)

func init() {
	transfersyntax.RegisterCodec(transfersyntax.JPEGLosslessSV1.UID, jpegDecode, jpegEncode)
	transfersyntax.RegisterCodec(transfersyntax.JPEGLossless.UID, jpegDecode, jpegEncode)
	transfersyntax.RegisterCodec(transfersyntax.JPEGBaseline8Bit.UID, jpegDecode, jpegEncode)
	transfersyntax.RegisterCodec(transfersyntax.JPEGExtended12Bit.UID, jpeg12Decode, jpeg12Encode)
}

func jpegDecode(j uint32, bitsa uint16, in []byte, inSize uint32, out []byte, outSize uint32) error {
	offset := j * outSize
	if bitsa == 8 {
		return jpeglib.DIJG8decode(in, inSize, out[offset:], outSize)
	} else {
		return jpeglib.DIJG16decode(in, inSize, out[offset:], outSize)
	}
}

func jpegEncode(j uint32, RGB bool, img []byte, cols uint16, rows uint16, samples uint16, bitsa uint16, JPEGData *[]byte, JPEGBytes *int, mode int) error {
	offset := j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	if RGB {
		offset = 3 * offset
	}
	if bitsa == 8 {
		if RGB {
			return jpeglib.EIJG8encode(img[offset:], cols, rows, 3, JPEGData, JPEGBytes, mode)
		} else {
			return jpeglib.EIJG8encode(img[offset:], cols, rows, 1, JPEGData, JPEGBytes, mode)
		}
	} else {
		return jpeglib.EIJG16encode(img[offset/2:], cols, rows, 1, JPEGData, JPEGBytes, 0)
	}
}

func jpeg12Decode(j uint32, _ uint16, in []byte, inSize uint32, out []byte, outSize uint32) error {
	offset := j * outSize
	return jpeglib.DIJG12decode(in, inSize, out[offset:], outSize)
}

func jpeg12Encode(j uint32, _ bool, img []byte, cols uint16, rows uint16, _ uint16, bitsa uint16, JPEGData *[]byte, JPEGBytes *int, _ int) error {
	offset := j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	return jpeglib.EIJG12encode(img[offset/2:], cols, rows, 1, JPEGData, JPEGBytes, 0)
}
