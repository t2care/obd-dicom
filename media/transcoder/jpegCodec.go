//go:build cgo && jpeg

package transcoder

import (
	"fmt"

	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media/transcoder/jpeglib"
)

func init() {
	transfersyntax.RegisterCodec(transfersyntax.JPEGLosslessSV1.UID, jpegDecode, jpegEncode)
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

func jpegEncode(j uint32, RGB bool, img []byte, cols uint16, rows uint16, samples uint16, bitsa uint16, bitss uint16, ww, wc float64, JPEGData *[]byte, JPEGBytes *int, mode int) error {
	offset := j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	if RGB {
		offset = 3 * offset
	}

	if bitsa == 16 {
		img, _ = scale16to8(img, bitss, wc, ww)
	}

	if RGB {
		return jpeglib.EIJG8encode(img[offset:], cols, rows, 3, JPEGData, JPEGBytes, mode)
	} else {
		return jpeglib.EIJG8encode(img[offset:], cols, rows, 1, JPEGData, JPEGBytes, mode)
	}
}

func scale16to8(img16 []byte, bitss uint16, wc, ww float64) ([]byte, error) {
	n := len(img16)
	if n%2 != 0 {
		return nil, fmt.Errorf("buffer 16 bits invalide (%d octets)", n)
	}
	out := make([]byte, n/2)

	L := wc - ww/2.0
	U := wc + ww/2.0
	if L < 0 {
		L = 0
	}
	maxRaw := float64(int(1) << bitss)
	if U > maxRaw {
		U = maxRaw
	}
	delta := U - L
	if delta <= 0 {
		delta = 1
	}

	for i := 0; i < n/2; i++ {
		low := img16[2*i]
		high := img16[2*i+1]
		raw := float64(int(high)<<8 | int(low))

		if raw < L {
			raw = L
		} else if raw > U {
			raw = U
		}

		norm := (raw - L) * 255.0 / delta
		out[i] = byte(norm)
	}
	return out, nil
}

func jpeg12Decode(j uint32, _ uint16, in []byte, inSize uint32, out []byte, outSize uint32) error {
	offset := j * outSize
	return jpeglib.DIJG12decode(in, inSize, out[offset:], outSize)
}

func jpeg12Encode(j uint32, _ bool, img []byte, cols uint16, rows uint16, _ uint16, bitsa uint16, bitss uint16, ww, wc float64, JPEGData *[]byte, JPEGBytes *int, _ int) error {
	offset := j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	return jpeglib.EIJG12encode(img[offset/2:], cols, rows, 1, JPEGData, JPEGBytes, 0)
}
