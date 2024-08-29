package transfersyntax

type TransferSyntax struct {
	UID         string
	Name        string
	Description string
	Type        string
}

var supportedTransferSyntaxes = []*TransferSyntax{
	ImplicitVRLittleEndian,
	ExplicitVRLittleEndian,
	DeflatedExplicitVRLittleEndian,
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
	return nil
}

func SupportedTransferSyntax(uid string) bool {
	for _, ts := range supportedTransferSyntaxes {
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
	supportedTransferSyntaxes = append(supportedTransferSyntaxes, GetTransferSyntaxFromUID(uid))
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
