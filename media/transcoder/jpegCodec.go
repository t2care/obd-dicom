package transcoder

import (
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media/transcoder/jpeglib"
)

func init() {
	transfersyntax.RegisterCodec(transfersyntax.JPEGLosslessSV1.UID, jpegDecode, jpegEncode)
	transfersyntax.RegisterCodec(transfersyntax.JPEGLossless.UID, jpegDecode, jpegEncode)
}

func jpegDecode(j uint32, bitsa uint16, j2kData []byte, j2kSize uint32, outputData []byte, single uint32) error {
	offset := j * single
	if bitsa == 8 {
		return jpeglib.DIJG8decode(j2kData, j2kSize, outputData[offset:], single)
	} else {
		return jpeglib.DIJG16decode(j2kData, j2kSize, outputData[offset:], single)
	}
}

func jpegEncode(j uint32, RGB bool, img []byte, cols uint16, rows uint16, samples uint16, bitsa uint16, JPEGData *[]byte, JPEGBytes *int, ratio int) error {
	offset := j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	if RGB {
		offset = 3 * offset
	}
	if bitsa == 8 {
		if RGB {
			return jpeglib.EIJG8encode(img[offset:], cols, rows, 3, JPEGData, JPEGBytes, ratio)
		} else {
			return jpeglib.EIJG8encode(img[offset:], cols, rows, 1, JPEGData, JPEGBytes, ratio)
		}
	} else {
		return jpeglib.EIJG16encode(img[offset/2:], cols, rows, 1, JPEGData, JPEGBytes, 0)
	}
}
