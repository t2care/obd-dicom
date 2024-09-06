package media

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
)

// DcmTag DICOM tag structure
type DcmTag struct {
	Name        string
	Description string
	Group       uint16
	Element     uint16
	Length      uint32
	VR          string
	VM          string
	Data        []byte
	BigEndian   bool
}

// getUShort convert tag.Data to uint16
func (tag *DcmTag) getUShort() uint16 {
	if tag.Length == 2 {
		if tag.BigEndian {
			return binary.BigEndian.Uint16(tag.Data)
		}
		return binary.LittleEndian.Uint16(tag.Data)
	}
	return 0
}

// getString convert tag.Data to string
func (tag *DcmTag) getString() string {
	n := bytes.IndexByte(tag.Data, 0)
	if n == -1 {
		n = int(tag.Length)
	}
	return strings.TrimSpace(string(tag.Data[:n]))
}

// writeSeq - Create an SQ tag from a DICOM Object
func (tag *DcmTag) writeSeq(group uint16, element uint16, seq *DcmObj) {
	bufdata := &BufData{
		BigEndian: false,
		MS:        NewEmptyMemoryStream(),
	}

	bufdata.BigEndian = seq.IsBigEndian()
	tag.BigEndian = seq.IsBigEndian()
	tag.Group = group
	tag.Element = element
	if tag.Group == 0xFFFE {
		tag.VR = ""
	} else {
		tag.VR = "SQ"
	}
	for i := 0; i < seq.TagCount(); i++ {
		temptag := seq.GetTagAt(i)
		bufdata.WriteTag(temptag, seq.IsExplicitVR())
	}
	tag.Length = uint32(bufdata.GetSize())
	if tag.Length%2 == 1 {
		tag.Length++
		bufdata.MS.Write([]byte{0x00}, 1)
	}
	if tag.Length > 0 {
		bufdata.SetPosition(0)
		data, _ := bufdata.MS.Read(int(tag.Length))
		tag.Data = data
	}
}

// ReadSeq - reads a dicom sequence
func (tag *DcmTag) ReadSeq(ExplicitVR bool) (*DcmObj, error) {
	seq := NewEmptyDCMObj()
	bufdata := &BufData{
		BigEndian: false,
		MS:        NewEmptyMemoryStream(),
	}

	bufdata.Write(tag.Data, int(tag.Length))
	bufdata.MS.SetPosition(0)
	var tempTags *DcmObj
	haveItem := false
	for bufdata.MS.GetPosition() < bufdata.MS.GetSize() {
		temptag, err := bufdata.ReadTag(ExplicitVR)
		if err != nil {
			return seq, fmt.Errorf("cannot read (%04X,%04X). Error: %s", tag.Group, tag.Element, err.Error())
		}

		if !ExplicitVR {
			temptag.VR = getDictionaryVR(tag.Group, tag.Element)
		}
		switch temptag.Element {
		case 0xE000:
			if temptag.Length != 0xFFFFFFFF {
				seq.Add(temptag)
				continue
			}
			haveItem = true
			tempTags = new(DcmObj)
		case 0xE00D:
			item := new(DcmTag)
			item.writeItem(tempTags)
			seq.Add(item)
		default:
			if haveItem {
				tempTags.Add(temptag)
			} else {
				seq.Add(temptag)
			}
		}
	}
	return seq, nil
}

func (tag *DcmTag) writeItem(obj *DcmObj) {
	tag.writeSeq(0xFFFE, 0xE000, obj)
}

func (tag *DcmTag) transcode(explicitVR bool, outTS *transfersyntax.TransferSyntax) error {
	seq := new(DcmObj)
	seq.SetTransferSyntax(outTS)
	if (explicitVR != seq.IsExplicitVR()) && tag.isSequence() {
		seq, err := tag.ReadSeq(explicitVR)
		if err != nil {
			return err
		}
		for _, item := range seq.GetTags() {
			item.transcode(explicitVR, outTS)
		}
		seq.SetTransferSyntax(outTS)
		tag.writeSeq(tag.Group, tag.Element, seq)
	}
	return nil
}

func (tag *DcmTag) isSequence() bool {
	return tag.VR == "SQ" || (tag.Group == 0xFFFE && tag.Element == 0xE000)
}

// Length is undefined 0xFFFFFFFF
func (tag *DcmTag) isSequenceUndefined() bool {
	return tag.isSequence() && tag.Length == 0xFFFFFFFF
}

func (tag *DcmTag) isSequenceEnd() bool {
	return (tag.Group == 0xFFFE && tag.Element == 0xE00D) || (tag.Group == 0xFFFE && tag.Element == 0xE0DD)
}
