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

func getTransferSyntaxFromName(name string) *TransferSyntax {
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

type decodeFunc func(j2kData []byte, j2kSize uint32, outputData []byte) error
type encodeFunc func(rawData []byte, width uint16, height uint16, samples uint16, bitsa uint16, outData *[]byte, outSize *int, ratio int) error

var decodes = make(map[string]decodeFunc)
var encodes = make(map[string]encodeFunc)

func RegisterCodec(uid string, decode decodeFunc, encode encodeFunc) {
	decodes[uid] = decode
	SupportedTransferSyntaxes = append(SupportedTransferSyntaxes, GetTransferSyntaxFromUID(uid))
}

func (ts *TransferSyntax) Decode(j2kData []byte, j2kSize uint32, outputData []byte) error {
	if fn, ok := decodes[ts.UID]; ok {
		return fn(j2kData, j2kSize, outputData)
	}
	return nil
}

func (ts *TransferSyntax) Encode(rawData []byte, width uint16, height uint16, samples uint16, bitsa uint16, outData *[]byte, outSize *int, ratio int) error {
	if fn, ok := encodes[ts.UID]; ok {
		return fn(rawData, width, height, samples, bitsa, outData, outSize, ratio)
	}
	return nil
}
