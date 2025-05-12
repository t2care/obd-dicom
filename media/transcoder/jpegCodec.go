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

func jpegEncode(j uint32, RGB bool, img []byte, cols uint16, rows uint16, samples uint16, bitsa uint16, bitss uint16, ww, wc, rs, ri float64, JPEGData *[]byte, JPEGBytes *int, mode int) error {
	offset := calcOffset(j, RGB, cols, rows, bitsa)
	if RGB {
		offset = 3 * offset
	}

	if bitsa == 16 {
		var err error
		img, err = scale16to8Bits(img, bitss, wc, ww, rs, ri)
		if err != nil {
			return err
		}
		offset = calcOffset(j, RGB, cols, rows, 8)
	}

	if RGB {
		return jpeglib.EIJG8encode(img[offset:], cols, rows, 3, JPEGData, JPEGBytes, mode)
	} else {
		return jpeglib.EIJG8encode(img[offset:], cols, rows, 1, JPEGData, JPEGBytes, mode)
	}
}

func calcOffset(j uint32, RGB bool, cols uint16, rows uint16, bitsa uint16) uint32 {
	offset := j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	if RGB {
		offset = 3 * offset
	}
	return offset
}

func scale16to8Bits(img16 []byte, bitss uint16, wc, ww, rs, ri float64) ([]byte, error) {
	n := len(img16)
	if n%2 != 0 {
		return nil, fmt.Errorf("invalid 16-bit buffer size (%d bytes): the buffer size must be even", n)
	}
	out := make([]byte, n/2)

	if wc <= 0 || ww <= 0 {
		maxRaw := float64(int(1) << bitss)
		wc = maxRaw / 2.0
		ww = maxRaw
	}

	for i := 0; i < n/2; i++ {
		raw := float64(int(img16[2*i+1])<<8 | int(img16[2*i]))

		//Apply modality rescale if provided
		if rs != 0 && ri != 0 {
			raw = raw*rs + ri
		}

		co := wc - 0.5
		hw := (ww - 1) / 2.0

		//Linear conversion
		var y float64
		if raw <= co-hw {
			y = 0
		} else if raw > co+hw {
			y = 255
		} else {
			y = ((raw-co)/(ww-1) + 0.5) * 255
		}
		out[i] = byte(y)
	}
	return out, nil
}

func jpeg12Decode(j uint32, _ uint16, in []byte, inSize uint32, out []byte, outSize uint32) error {
	offset := j * outSize
	return jpeglib.DIJG12decode(in, inSize, out[offset:], outSize)
}

func jpeg12Encode(j uint32, _ bool, img []byte, cols uint16, rows uint16, _ uint16, bitsa uint16, bitss uint16, ww, wc, rs, ri float64, JPEGData *[]byte, JPEGBytes *int, _ int) error {
	offset := j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	return jpeglib.EIJG12encode(img[offset/2:], cols, rows, 1, JPEGData, JPEGBytes, 0)
}
