package transfersyntax

type TransferSyntax struct {
	UID         string
	Name        string
	Description string
	Type        string
}

var SupportedTransferSyntaxes = []*TransferSyntax{
	ImplicitVRLittleEndian,
	ExplicitVRLittleEndian,
	ExplicitVRBigEndian,
	JPEGLosslessSV1,
	JPEGBaseline8Bit,
	JPEGExtended12Bit,
}

var tsMap map[string]*TransferSyntax

func init() {
	tsMap = make(map[string]*TransferSyntax, len(transferSyntaxes))
	for _, ts := range transferSyntaxes {
		tsMap[ts.UID] = ts
	}
}

func GetTransferSyntaxFromName(name string) *TransferSyntax {
	for _, ts := range transferSyntaxes {
		if ts.Name == name {
			return ts
		}
	}
	return nil
}

func GetTransferSyntaxFromUID(uid string) *TransferSyntax {
	if ts, ok := tsMap[uid]; ok {
		return ts
	}
	// Extra loop to fix old bug
	uid = string([]rune(uid)[:len(uid)-1])
	for _, ts := range transferSyntaxes {
		if ts.UID == uid {
			return ts
		}
	}
	return nil
}

func SupportedTransferSyntax(uid string) bool {
	for _, ts := range SupportedTransferSyntaxes {
		if ts.UID == uid {
			return true
		}
	}
	return false
}

type decodeFunc func(frame uint32, bitsa uint16, j2kData []byte, j2kSize uint32, outputData []byte, outputSize uint32) error
type encodeFunc func(frame uint32, RGB bool, rawData []byte, width uint16, height uint16, samples uint16, bitsa uint16, outData *[]byte, outSize *int, ratio int) error

var decodes = make(map[string]decodeFunc)
var encodes = make(map[string]encodeFunc)

func RegisterCodec(uid string, decode decodeFunc, encode encodeFunc) {
	decodes[uid] = decode
	SupportedTransferSyntaxes = append(SupportedTransferSyntaxes, GetTransferSyntaxFromUID(uid))
}

func (ts *TransferSyntax) Decode(frame uint32, bitsa uint16, j2kData []byte, j2kSize uint32, outputData []byte, outputSize uint32) error {
	if fn, ok := decodes[ts.UID]; ok {
		return fn(frame, bitsa, j2kData, j2kSize, outputData, outputSize)
	}
	return nil
}

func (ts *TransferSyntax) Encode(frame uint32, RGB bool, rawData []byte, width uint16, height uint16, samples uint16, bitsa uint16, outData *[]byte, outSize *int, ratio int) error {
	if fn, ok := encodes[ts.UID]; ok {
		return fn(frame, RGB, rawData, width, height, samples, bitsa, outData, outSize, ratio)
	}
	return nil
}
