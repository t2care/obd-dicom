package media

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/t2care/obd-dicom/dictionary/sopclass"
	"github.com/t2care/obd-dicom/dictionary/tags"
	"github.com/t2care/obd-dicom/dictionary/transfersyntax"
	"github.com/t2care/obd-dicom/media/transcoder"
	"github.com/t2care/obd-dicom/media/transcoder/jpeglib"
)

type DcmObj struct {
	Tags           []*DcmTag
	TransferSyntax *transfersyntax.TransferSyntax
	ExplicitVR     bool
	BigEndian      bool
	SQtag          *DcmTag
}

type ParseOptions struct {
	OnlyMetaHeader  bool // Group 0x0002. Useful for storeSCU for exemple
	UntilPatientTag bool // Until group 0x0010
	SkipPixelData   bool
	SkipFillTag     bool // Increase perf by skipping FillTag. Filltag is only useful for dumpTags
}

// NewEmptyDCMObj - Create as an interface to a new empty dcmObj
func NewEmptyDCMObj() *DcmObj {
	return &DcmObj{
		Tags:           make([]*DcmTag, 0),
		TransferSyntax: nil,
		ExplicitVR:     false,
		BigEndian:      false,
		SQtag:          &DcmTag{},
	}
}

// NewDCMObjFromFile - Read from a DICOM file into a DICOM Object
func NewDCMObjFromFile(fileName string, opt ...*ParseOptions) (*DcmObj, error) {
	if _, err := os.Stat(fileName); err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New("DcmObj::Read, file does not exist")
		}
		return nil, fmt.Errorf("DcmObj::Read %s", err.Error())
	}

	bufdata, err := NewBufDataFromFile(fileName)
	if err != nil {
		return nil, err
	}

	return parseBufData(bufdata, opt...)
}

// NewDCMObjFromBytes - Read from a DICOM bytes into a DICOM Object
func NewDCMObjFromBytes(data []byte) (*DcmObj, error) {
	return parseBufData(NewBufDataFromBytes(data))
}

func parseBufData(bufdata *BufData, opt ...*ParseOptions) (*DcmObj, error) {
	BigEndian := false

	transferSyntax, err := bufdata.ReadMeta()
	if err != nil {
		return nil, err
	}

	obj := &DcmObj{
		Tags:           make([]*DcmTag, 0),
		TransferSyntax: transferSyntax,
		ExplicitVR:     false,
		BigEndian:      false,
		SQtag:          &DcmTag{},
	}

	if obj.TransferSyntax == nil {
		return nil, fmt.Errorf("unable to read transfer syntax from data")
	}

	if obj.TransferSyntax == transfersyntax.ImplicitVRLittleEndian {
		obj.ExplicitVR = false
	} else {
		obj.ExplicitVR = true
	}
	if obj.TransferSyntax == transfersyntax.ExplicitVRBigEndian {
		BigEndian = true
	}
	bufdata.SetBigEndian(BigEndian)

	if err := bufdata.ReadObj(obj, opt...); err != nil {
		return nil, err
	}

	return obj, nil
}

func (obj *DcmObj) IsExplicitVR() bool {
	return obj.ExplicitVR
}

func (obj *DcmObj) SetExplicitVR(explicit bool) {
	obj.ExplicitVR = explicit
}

func (obj *DcmObj) IsBigEndian() bool {
	return obj.BigEndian
}

func (obj *DcmObj) SetBigEndian(bigEndian bool) {
	obj.BigEndian = bigEndian
}

// TagCount - return the Tags number
func (obj *DcmObj) TagCount() int {
	return len(obj.Tags)
}

// GetTagAt - return the Tag at position i
func (obj *DcmObj) GetTagAt(i int) *DcmTag {
	return obj.Tags[i]
}

func (obj *DcmObj) GetTag(tag *tags.Tag) *DcmTag {
	for _, t := range obj.Tags {
		if t.Group == tag.Group && t.Element == tag.Element {
			return t
		}
	}
	return nil
}

func (obj *DcmObj) GetTagGE(group uint16, element uint16) *DcmTag {
	for _, t := range obj.Tags {
		if t.Group == group && t.Element == element {
			return t
		}
	}
	return nil
}

func (obj *DcmObj) SetTag(i int, tag *DcmTag) {
	FillTag(tag)
	if i <= obj.TagCount() {
		obj.Tags[i] = tag
	}
}

func (obj *DcmObj) InsertTag(index int, tag *DcmTag) {
	FillTag(tag)
	obj.Tags = append(obj.Tags[:index+1], obj.Tags[index:]...)
	obj.Tags[index] = tag
}

func (obj *DcmObj) GetTags() []*DcmTag {
	return obj.Tags
}

func (obj *DcmObj) DelTag(i int) {
	obj.Tags = append(obj.Tags[:i], obj.Tags[i+1:]...)
}

func (obj *DcmObj) DumpTags() error {
	for _, tag := range obj.Tags {
		if tag.VR == "SQ" {
			fmt.Printf("\t(%04X,%04X) %s - %s\n", tag.Group, tag.Element, tag.VR, tag.Description)
			seq, err := tag.ReadSeq(obj.IsExplicitVR())
			if err != nil {
				return err
			}
			seq.dumpSeq(1)
			continue
		}
		if tag.Length > 128 {
			fmt.Printf("\t(%04X,%04X) %s - %s : (Not displayed)\n", tag.Group, tag.Element, tag.VR, tag.Description)
			continue
		}
		switch tag.VR {
		case "US":
			fmt.Printf("\t(%04X,%04X) %s - %s : %d\n", tag.Group, tag.Element, tag.VR, tag.Description, binary.LittleEndian.Uint16(tag.Data))
		default:
			fmt.Printf("\t(%04X,%04X) %s - %s : %s\n", tag.Group, tag.Element, tag.VR, tag.Description, tag.Data)
		}
	}
	fmt.Println()
	return nil
}

func (obj *DcmObj) dumpSeq(indent int) error {
	tabs := "\t"
	for i := 0; i < indent; i++ {
		tabs += "\t"
	}

	for _, tag := range obj.Tags {
		if tag.isSequence() {
			fmt.Printf("%s(%04X,%04X) %s - %s\n", tabs, tag.Group, tag.Element, tag.VR, tag.Description)
			seq, err := tag.ReadSeq(obj.IsExplicitVR())
			if err != nil {
				return err
			}
			seq.dumpSeq(indent + 1)
			continue
		}
		if tag.Length > 128 {
			fmt.Printf("%s(%04X,%04X) %s - %s : (Not displayed)\n", tabs, tag.Group, tag.Element, tag.VR, tag.Description)
			continue
		}
		switch tag.VR {
		case "US":
			fmt.Printf("%s(%04X,%04X) %s - %s : %d\n", tabs, tag.Group, tag.Element, tag.VR, tag.Description, binary.LittleEndian.Uint16(tag.Data))
		default:
			fmt.Printf("%s(%04X,%04X) %s - %s : %s\n", tabs, tag.Group, tag.Element, tag.VR, tag.Description, tag.Data)
		}
	}
	return nil
}

func (obj *DcmObj) GetDate(tag *tags.Tag) time.Time {
	date := obj.GetString(tag)
	data, _ := time.Parse("20060102", date)
	return data
}

func (obj *DcmObj) GetUShort(tag *tags.Tag) uint16 {
	return obj.getUShortGE(tag.Group, tag.Element)
}

// GetUShortGE - return the Uint16 for this group & element
func (obj *DcmObj) getUShortGE(group uint16, element uint16) uint16 {
	for _, tag := range obj.GetTags() {
		if (tag.Group == group) && (tag.Element == element) {
			return tag.getUShort()
		}
	}
	return 0
}

func (obj *DcmObj) GetString(tag *tags.Tag) string {
	return obj.GetStringGE(tag.Group, tag.Element)
}

// GetStringGE - return the String for this group & element
func (obj *DcmObj) GetStringGE(group uint16, element uint16) string {
	for _, tag := range obj.GetTags() {
		if (tag.Group == group) && (tag.Element == element) {
			return tag.getString()
		}
	}
	return ""
}

// Add - add a new DICOM Tag to a DICOM Object
func (obj *DcmObj) Add(tag *DcmTag) {
	obj.Tags = append(obj.Tags, tag)
}

func (obj *DcmObj) WriteToBytes() []byte {
	bufdata := NewEmptyBufData()
	SOPClassUID := obj.GetStringGE(0x08, 0x16)
	SOPInstanceUID := obj.GetStringGE(0x08, 0x18)
	bufdata.WriteMeta(SOPClassUID, SOPInstanceUID, obj.TransferSyntax.UID)
	// Don't write metadata header in BigEndian (ReadMeta does not support )
	if obj.TransferSyntax.UID == transfersyntax.ExplicitVRBigEndian.UID {
		bufdata.SetBigEndian(true)
	}
	bufdata.WriteObj(obj)
	bufdata.SetPosition(0)
	return bufdata.GetAllBytes()
}

// Wrote - Write a DICOM Object to a DICOM File
func (obj *DcmObj) WriteToFile(fileName string) error {
	data := obj.WriteToBytes()
	return os.WriteFile(fileName, data, 0644)
}

func (obj *DcmObj) WriteDate(tag *tags.Tag, date time.Time) {
	obj.WriteString(tag, date.Format("20060102"))
}

func (obj *DcmObj) WriteDateRange(tag *tags.Tag, startDate time.Time, endDate time.Time) {
	obj.WriteString(tag, fmt.Sprintf("%s-%s", startDate.Format("20060102"), endDate.Format("20060102")))
}

func (obj *DcmObj) WriteTime(tag *tags.Tag, date time.Time) {
	obj.WriteString(tag, date.Format("150405"))
}

func (obj *DcmObj) WriteUint16(tag *tags.Tag, val uint16) {
	obj.WriteUint16GE(tag.Group, tag.Element, tag.VR, val)
}

func (obj *DcmObj) WriteUint32(tag *tags.Tag, val uint32) {
	obj.WriteUint32GE(tag.Group, tag.Element, tag.VR, val)
}

func (obj *DcmObj) WriteString(tag *tags.Tag, content string) {
	obj.WriteStringGE(tag.Group, tag.Element, tag.VR, content)
}

// WriteUint16GE - Writes a Uint16 to a DICOM tag
func (obj *DcmObj) WriteUint16GE(group uint16, element uint16, vr string, val uint16) {
	c := make([]byte, 2)
	if obj.BigEndian {
		binary.BigEndian.PutUint16(c, val)
	} else {
		binary.LittleEndian.PutUint16(c, val)
	}

	tag := &DcmTag{
		Group:     group,
		Element:   element,
		Length:    2,
		VR:        vr,
		Data:      c,
		BigEndian: obj.BigEndian,
	}
	FillTag(tag)
	obj.Tags = append(obj.Tags, tag)
}

// WriteUint32GE - Writes a Uint32 to a DICOM tag
func (obj *DcmObj) WriteUint32GE(group uint16, element uint16, vr string, val uint32) {
	c := make([]byte, 4)
	if obj.BigEndian {
		binary.BigEndian.PutUint32(c, val)
	} else {
		binary.LittleEndian.PutUint32(c, val)
	}

	tag := &DcmTag{
		Group:     group,
		Element:   element,
		Length:    4,
		VR:        vr,
		Data:      c,
		BigEndian: obj.BigEndian,
	}
	FillTag(tag)
	obj.Tags = append(obj.Tags, tag)
}

// WriteStringGE - Writes a String to a DICOM tag
func (obj *DcmObj) WriteStringGE(group uint16, element uint16, vr string, content string) {
	data := []byte(content)
	length := len(data)
	if length%2 == 1 {
		length++
		if vr == "UI" {
			data = append(data, 0x00)
		} else {
			data = append(data, 0x20)
		}
	}
	tag := &DcmTag{
		Group:     group,
		Element:   element,
		Length:    uint32(length),
		VR:        vr,
		Data:      data,
		BigEndian: false,
	}
	FillTag(tag)
	obj.Tags = append(obj.Tags, tag)
}

func (obj *DcmObj) GetTransferSyntax() *transfersyntax.TransferSyntax {
	return obj.TransferSyntax
}

func (obj *DcmObj) SetTransferSyntax(ts *transfersyntax.TransferSyntax) {
	obj.TransferSyntax = ts
	switch ts {
	case transfersyntax.ImplicitVRLittleEndian:
		obj.SetBigEndian(false)
		obj.SetExplicitVR(false)
	case transfersyntax.ExplicitVRBigEndian:
		obj.SetBigEndian(true)
		obj.SetExplicitVR(true)
	default:
		obj.SetBigEndian(false)
		obj.SetExplicitVR(true)
	}
}

func (obj *DcmObj) GetPixelData(frame int) ([]byte, error) {
	var i int
	var rows, cols, bitsa, planar uint16
	var PhotoInt string
	sq := 0
	frames := uint32(0)
	RGB := false
	icon := false

	if !transfersyntax.SupportedTransferSyntax(obj.TransferSyntax.UID) {
		return nil, fmt.Errorf("unsupported transfer synxtax %s", obj.TransferSyntax.Name)
	}

	for i = 0; i < len(obj.Tags); i++ {
		tag := obj.GetTagAt(i)
		if tag.isSequenceUndefined() {
			sq++
		}
		if sq == 0 {
			if (tag.Group == 0x0028) && (!icon) {
				switch tag.Element {
				case 0x04:
					PhotoInt = tag.getString()
					if !strings.Contains(PhotoInt, "MONO") {
						RGB = true
					}
				case 0x06:
					planar = tag.getUShort()
				case 0x08:
					uframes, err := strconv.Atoi(tag.getString())
					if err != nil {
						frames = 0
					} else {
						frames = uint32(uframes)
					}
				case 0x10:
					rows = tag.getUShort()
				case 0x11:
					cols = tag.getUShort()
				case 0x0100:
					bitsa = tag.getUShort()
				}
			}
			if (tag.Group == 0x0088) && (tag.Element == 0x0200) && (tag.Length == 0xFFFFFFFF) {
				icon = true
			}
			if (tag.Group == 0x6003) && (tag.Element == 0x1010) && (tag.Length == 0xFFFFFFFF) {
				icon = true
			}
			if (tag.Group == 0x7FE0) && (tag.Element == 0x0010) && (!icon) {
				size := uint32(cols) * uint32(rows) * uint32(bitsa) / 8
				if RGB {
					size = 3 * size
				}
				if frames > 0 {
					size = uint32(frames) * size
				} else {
					frames = 1
				}
				if size == 0 {
					return nil, errors.New("DcmObj::ConvertTransferSyntax, size=0")
				}

				if frame > int(frames) {
					return nil, errors.New("invalid frame")
				}

				if tag.Length == 0xFFFFFFFF {
					return obj.GetTagAt(i + 2 + frame).Data, nil
				} else {
					if RGB && (planar == 1) {
						var img_offset, img_size uint32
						img_size = size / frames
						img := make([]byte, img_size)
						for f := uint32(0); f < frames; f++ {
							img_offset = img_size * f
							for j := uint32(0); j < img_size/3; j++ {
								img[3*j] = tag.Data[j+img_offset]
								img[3*j+1] = tag.Data[j+img_size/3+img_offset]
								img[3*j+2] = tag.Data[j+2*img_size/3+img_offset]
							}
							if f == uint32(frame) {
								return img, nil
							}
						}
						planar = 0
					} else {
						return tag.Data, nil
					}
				}
			}
		}
		if tag.isSequenceEnd() {
			sq--
		}
	}
	return nil, fmt.Errorf("there was an error getting pixel data")
}

func (obj *DcmObj) ChangeTransferSynx(outTS *transfersyntax.TransferSyntax) error {
	flag := false

	var i int
	var rows, cols, bitss, bitsa, planar, pixelrep uint16
	var PhotoInt string
	sq := 0
	frames := uint32(0)
	RGB := false
	icon := false

	if obj.TransferSyntax.UID == outTS.UID {
		return nil
	}

	if !transfersyntax.SupportedTransferSyntax(outTS.UID) {
		return fmt.Errorf("unsupported transfer synxtax %s", outTS.Name)
	}

	for i = 0; i < len(obj.Tags); i++ {
		tag := obj.GetTagAt(i)
		if tag.isSequence() {
			if tag.Length == 0xFFFFFFFF {
				sq++
			} else {
				if err := tag.transcode(obj.IsExplicitVR(), outTS); err != nil {
					return err
				}
				flag = true
			}
		}
		if sq == 0 {
			if (tag.Group == 0x0028) && (!icon) {
				switch tag.Element {
				case 0x04:
					PhotoInt = tag.getString()
					if !strings.Contains(PhotoInt, "MONO") {
						RGB = true
					}
				case 0x06:
					planar = tag.getUShort()
				case 0x08:
					uframes, err := strconv.Atoi(tag.getString())
					if err != nil {
						frames = 0
					} else {
						frames = uint32(uframes)
					}
				case 0x10:
					rows = tag.getUShort()
				case 0x11:
					cols = tag.getUShort()
				case 0x0100:
					bitsa = tag.getUShort()
				case 0x0101:
					bitss = tag.getUShort()
				case 0x0103:
					pixelrep = tag.getUShort()
				}
			}
			if (tag.Group == 0x0088) && (tag.Element == 0x0200) && (tag.Length == 0xFFFFFFFF) {
				icon = true
			}
			if (tag.Group == 0x6003) && (tag.Element == 0x1010) && (tag.Length == 0xFFFFFFFF) {
				icon = true
			}
			if (tag.Group == 0x7FE0) && (tag.Element == 0x0010) && (!icon) {
				size := uint32(cols) * uint32(rows) * uint32(bitsa) / 8
				if RGB {
					size = 3 * size
				}
				if frames > 0 {
					size = uint32(frames) * size
				} else {
					frames = 1
				}
				if size == 0 {
					return errors.New("DcmObj::ConvertTransferSyntax, size=0")
				}
				img := make([]byte, size)
				if tag.Length == 0xFFFFFFFF {
					obj.uncompress(i, img, size, frames, bitsa, PhotoInt)
				} else { // Uncompressed
					if RGB && (planar == 1) { // change from planar=1 to planar=0
						var img_offset, img_size uint32
						img_size = size / frames
						for f := uint32(0); f < frames; f++ {
							img_offset = img_size * f
							for j := uint32(0); j < img_size/3; j++ {
								img[3*j+img_offset] = tag.Data[j+img_offset]
								img[3*j+1+img_offset] = tag.Data[j+img_size/3+img_offset]
								img[3*j+2+img_offset] = tag.Data[j+2*img_size/3+img_offset]
							}
						}
						planar = 0
					} else {
						copy(img, tag.Data)
					}
				}
				if err := obj.compress(&i, img, RGB, cols, rows, bitss, bitsa, pixelrep, planar, frames, outTS.UID); err != nil {
					return err
				} else {
					flag = true
				}
			}
		}
		if tag.isSequenceEnd() {
			sq--
		}
	}
	if flag {
		obj.SetTransferSyntax(outTS)
		return nil
	}
	return fmt.Errorf("there was an error changing the transfer synxtax")
}

// AddConceptNameSeq - Concept Name Sequence for DICOM SR
func (obj *DcmObj) AddConceptNameSeq(group uint16, element uint16, CodeValue string, CodeMeaning string) {
	item := &DcmObj{
		Tags:           make([]*DcmTag, 0),
		TransferSyntax: nil,
		ExplicitVR:     false,
		BigEndian:      false,
		SQtag:          new(DcmTag),
	}
	seq := &DcmObj{
		Tags:           make([]*DcmTag, 0),
		TransferSyntax: nil,
		ExplicitVR:     false,
		BigEndian:      false,
		SQtag:          new(DcmTag),
	}
	tag := new(DcmTag)

	item.BigEndian = obj.BigEndian
	item.ExplicitVR = obj.ExplicitVR
	seq.BigEndian = obj.BigEndian
	seq.ExplicitVR = obj.ExplicitVR

	item.WriteString(tags.CodeValue, CodeValue)
	item.WriteString(tags.CodingSchemeDesignator, "odb")
	item.WriteString(tags.CodeMeaning, CodeMeaning)
	tag.writeItem(item)
	seq.Add(tag)
	tag = new(DcmTag)
	tag.writeSeq(group, element, seq)
	obj.Add(tag)
}

// AddSRText - add Text to SR
func (obj *DcmObj) AddSRText(text string) {
	item := &DcmObj{
		Tags:           make([]*DcmTag, 0),
		TransferSyntax: nil,
		ExplicitVR:     false,
		BigEndian:      false,
		SQtag:          new(DcmTag),
	}
	seq := &DcmObj{
		Tags:           make([]*DcmTag, 0),
		TransferSyntax: nil,
		ExplicitVR:     false,
		BigEndian:      false,
		SQtag:          new(DcmTag),
	}
	tag := new(DcmTag)

	item.BigEndian = obj.BigEndian
	item.ExplicitVR = obj.ExplicitVR
	seq.BigEndian = obj.BigEndian
	seq.ExplicitVR = obj.ExplicitVR

	item.WriteString(tags.RelationshipType, "CONTAINS")
	item.WriteString(tags.ValueType, "TEXT")
	item.AddConceptNameSeq(0x40, 0xA043, "2222", "Report Text")
	item.WriteString(tags.TextValue, text)
	tag.writeSeq(0xFFFE, 0xE000, item)
	seq.Add(tag)
	tag.writeSeq(0x40, 0xA730, seq)
	obj.Add(tag)
}

// CreateSR - Create a DICOM SR object
func (obj *DcmObj) CreateSR(study DCMStudy, SeriesInstanceUID string, SOPInstanceUID string) {
	obj.WriteString(tags.InstanceCreationDate, time.Now().Format("20060102"))
	obj.WriteString(tags.InstanceCreationTime, time.Now().Format("150405"))
	obj.WriteString(tags.SOPClassUID, sopclass.BasicTextSRStorage.UID)
	obj.WriteString(tags.SOPInstanceUID, SOPInstanceUID)
	obj.WriteString(tags.AccessionNumber, study.AccessionNumber)
	obj.WriteString(tags.Modality, "SR")
	obj.WriteString(tags.InstitutionName, study.InstitutionName)
	obj.WriteString(tags.ReferringPhysicianName, study.ReferringPhysician)
	obj.WriteString(tags.StudyDescription, study.Description)
	obj.WriteString(tags.SeriesDescription, "REPORT")
	obj.WriteString(tags.PatientName, study.PatientName)
	obj.WriteString(tags.PatientID, study.PatientID)
	obj.WriteString(tags.PatientBirthDate, study.PatientBD)
	obj.WriteString(tags.PatientSex, study.PatientSex)
	obj.WriteString(tags.StudyInstanceUID, study.StudyInstanceUID)
	obj.WriteString(tags.SeriesInstanceUID, SeriesInstanceUID)
	obj.WriteString(tags.SeriesNumber, "200")
	obj.WriteString(tags.InstanceNumber, "1")
	obj.WriteString(tags.ValueType, "CONTAINER")
	obj.AddConceptNameSeq(0x0040, 0xA043, "1111", "Radiology Report")
	obj.WriteString(tags.ContinuityOfContent, "SEPARATE")
	obj.WriteString(tags.VerifyingObserverName, study.ObserverName)
	obj.WriteString(tags.CompletionFlag, "COMPLETE")
	obj.WriteString(tags.VerificationFlag, "VERIFIED")
	obj.AddSRText(study.ReportText)
}

// CreatePDF - Create a DICOM SR object
func (obj *DcmObj) CreatePDF(study DCMStudy, SeriesInstanceUID string, SOPInstanceUID string, fileName string) {
	obj.WriteString(tags.InstanceCreationDate, time.Now().Format("20060102"))
	obj.WriteString(tags.InstanceCreationTime, time.Now().Format("150405"))
	obj.WriteString(tags.SOPClassUID, sopclass.EncapsulatedPDFStorage.UID)
	obj.WriteString(tags.SOPInstanceUID, SOPInstanceUID)
	obj.WriteString(tags.AccessionNumber, study.AccessionNumber)
	obj.WriteString(tags.Modality, "OT")
	obj.WriteString(tags.InstitutionName, study.InstitutionName)
	obj.WriteString(tags.ReferringPhysicianName, study.ReferringPhysician)
	obj.WriteString(tags.StudyDescription, study.Description)
	obj.WriteString(tags.PatientName, study.PatientName)
	obj.WriteString(tags.PatientID, study.PatientID)
	obj.WriteString(tags.PatientBirthDate, study.PatientBD)
	obj.WriteString(tags.PatientSex, study.PatientSex)
	obj.WriteString(tags.StudyInstanceUID, study.StudyInstanceUID)
	obj.WriteString(tags.SeriesInstanceUID, SeriesInstanceUID)
	obj.WriteString(tags.SeriesNumber, "300")
	obj.WriteString(tags.InstanceNumber, "1")

	mstream, _ := NewMemoryStreamFromFile(fileName)

	mstream.SetPosition(0)
	size := uint32(mstream.GetSize())
	if size%2 == 1 {
		size++
		mstream.Append([]byte{0x00})
	}
	obj.WriteString(tags.DocumentTitle, fileName)
	obj.Add(&DcmTag{
		Group:     0x42,
		Element:   0x11,
		Length:    size,
		VR:        "OB",
		Data:      mstream.GetData(),
		BigEndian: obj.BigEndian,
	})
	obj.WriteString(tags.MIMETypeOfEncapsulatedDocument, "application/pdf")
}

func (obj *DcmObj) compress(i *int, img []byte, RGB bool, cols uint16, rows uint16, bitss uint16, bitsa uint16, pixelrep uint16, planar uint16, frames uint32, outTS string) error {
	var offset, size, jpeg_size, j uint32
	var JPEGData []byte
	var JPEGBytes, index int

	single := uint32(cols) * uint32(rows) * uint32(bitsa) / 8
	size = single * frames
	if RGB {
		size = 3 * size
	}

	index = *i
	tag := obj.GetTagAt(index)

	switch outTS {
	case transfersyntax.JPEGLosslessSV1.UID:
		tag.VR = "OB"
		tag.Length = 0xFFFFFFFF
		if tag.Data != nil {
			tag.Data = nil
		}
		obj.SetTag(index, tag)
		index++
		newtag := &DcmTag{
			Group:     0xFFFE,
			Element:   0xE000,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		for j = 0; j < frames; j++ {
			index++
			offset = j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
			if RGB {
				offset = 3 * offset
			}
			if bitsa == 8 {
				if RGB {
					if err := jpeglib.EIJG8encode(img[offset:], cols, rows, 3, &JPEGData, &JPEGBytes, 4); err != nil {
						return err
					}
				} else {
					if err := jpeglib.EIJG8encode(img[offset:], cols, rows, 1, &JPEGData, &JPEGBytes, 4); err != nil {
						return err
					}
				}
			} else {
				if err := jpeglib.EIJG16encode(img[offset/2:], cols, rows, 1, &JPEGData, &JPEGBytes, 0); err != nil {
					return err
				}
			}
			newtag = &DcmTag{
				Group:     0xFFFE,
				Element:   0xE000,
				Length:    uint32(JPEGBytes),
				VR:        "DL",
				Data:      JPEGData,
				BigEndian: obj.IsBigEndian(),
			}
			obj.InsertTag(index, newtag)
			JPEGData = nil
		}
		index++
		newtag = &DcmTag{
			Group:     0xFFFE,
			Element:   0xE0DD,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		*i = index
	case transfersyntax.JPEGBaseline8Bit.UID:
		tag.VR = "OB"
		tag.Length = 0xFFFFFFFF
		if tag.Data != nil {
			tag.Data = nil
		}
		obj.SetTag(index, tag)
		index++
		newtag := &DcmTag{
			Group:     0xFFFE,
			Element:   0xE000,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		jpeg_size = 0
		for j = 0; j < frames; j++ {
			index++
			offset = j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
			if RGB {
				offset = 3 * offset
				if err := jpeglib.EIJG8encode(img[offset:], cols, rows, 3, &JPEGData, &JPEGBytes, 0); err != nil {
					return err
				}
			} else {
				if bitsa == 8 {
					if err := jpeglib.EIJG8encode(img[offset:], cols, rows, 1, &JPEGData, &JPEGBytes, 0); err != nil {
						return err
					}
				} else {
					if err := jpeglib.EIJG12encode(img[offset:], cols, rows, 1, &JPEGData, &JPEGBytes, 0); err != nil {
						return err
					}
				}
			}
			newtag = &DcmTag{
				Group:     0xFFFE,
				Element:   0xE000,
				Length:    uint32(JPEGBytes),
				VR:        "DL",
				Data:      JPEGData,
				BigEndian: obj.IsBigEndian(),
			}
			obj.InsertTag(index, newtag)
			JPEGData = nil
			jpeg_size = jpeg_size + uint32(JPEGBytes)
		}
		index++
		newtag = &DcmTag{
			Group:     0xFFFE,
			Element:   0xE0DD,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		*i = index
	case transfersyntax.JPEGExtended12Bit.UID:
		tag.VR = "OB"
		tag.Length = 0xFFFFFFFF
		if tag.Data != nil {
			tag.Data = nil
		}
		obj.SetTag(index, tag)
		index++
		newtag := &DcmTag{
			Group:     0xFFFE,
			Element:   0xE000,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		jpeg_size = 0
		for j = 0; j < frames; j++ {
			index++
			offset = j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
			if err := jpeglib.EIJG12encode(img[offset/2:], cols, rows, 1, &JPEGData, &JPEGBytes, 0); err != nil {
				return err
			}
			newtag = &DcmTag{
				Group:     0xFFFE,
				Element:   0xE000,
				Length:    uint32(JPEGBytes),
				VR:        "DL",
				Data:      JPEGData,
				BigEndian: obj.IsBigEndian(),
			}
			obj.InsertTag(index, newtag)
			JPEGData = nil
			jpeg_size = jpeg_size + uint32(JPEGBytes)
		}
		index++
		newtag = &DcmTag{
			Group:     0xFFFE,
			Element:   0xE0DD,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		*i = index
	case transfersyntax.JPEG2000Lossless.UID:
		tag.VR = "OB"
		tag.Length = 0xFFFFFFFF
		if tag.Data != nil {
			tag.Data = nil
		}
		obj.SetTag(index, tag)
		index++
		newtag := &DcmTag{
			Group:     0xFFFE,
			Element:   0xE000,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		for j = 0; j < frames; j++ {
			index++
			offset = j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
			if RGB {
				offset = 3 * offset
				if err := transfersyntax.JPEG2000Lossless.Encode(img[offset:], cols, rows, 3, bitsa, &JPEGData, &JPEGBytes, 0); err != nil {
					return err
				}
			} else {
				if err := transfersyntax.JPEG2000Lossless.Encode(img[offset:], cols, rows, 1, bitsa, &JPEGData, &JPEGBytes, 0); err != nil {
					return err
				}
			}
			newtag = &DcmTag{
				Group:     0xFFFE,
				Element:   0xE000,
				Length:    uint32(JPEGBytes),
				VR:        "DL",
				Data:      JPEGData,
				BigEndian: obj.IsBigEndian(),
			}
			obj.InsertTag(index, newtag)
			JPEGData = nil
		}
		index++
		newtag = &DcmTag{
			Group:     0xFFFE,
			Element:   0xE0DD,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		*i = index
	case transfersyntax.JPEG2000.UID:
		tag.VR = "OB"
		tag.Length = 0xFFFFFFFF
		if tag.Data != nil {
			tag.Data = nil
		}
		obj.SetTag(index, tag)
		index++
		newtag := &DcmTag{
			Group:     0xFFFE,
			Element:   0xE000,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		jpeg_size = 0
		for j = 0; j < frames; j++ {
			index++
			offset = j * uint32(cols) * uint32(rows) * uint32(bitsa) / 8
			if RGB {
				offset = 3 * offset
				if err := transfersyntax.JPEG2000.Encode(img[offset:], cols, rows, 3, bitsa, &JPEGData, &JPEGBytes, 10); err != nil {
					return err
				}
			} else {
				if err := transfersyntax.JPEG2000.Encode(img[offset:], cols, rows, 1, bitsa, &JPEGData, &JPEGBytes, 10); err != nil {
					return err
				}
			}
			newtag = &DcmTag{
				Group:     0xFFFE,
				Element:   0xE000,
				Length:    uint32(JPEGBytes),
				VR:        "DL",
				Data:      JPEGData,
				BigEndian: obj.IsBigEndian(),
			}
			obj.InsertTag(index, newtag)
			JPEGData = nil
			jpeg_size = jpeg_size + uint32(JPEGBytes)
		}
		index++
		newtag = &DcmTag{
			Group:     0xFFFE,
			Element:   0xE0DD,
			Length:    0,
			VR:        "DL",
			Data:      nil,
			BigEndian: obj.IsBigEndian(),
		}
		obj.InsertTag(index, newtag)
		*i = index
	default:
		if bitss == 8 {
			tag.VR = "OB"
		} else {
			tag.VR = "OW"
		}
		tag.Length = size
		if tag.Data != nil {
			tag.Data = nil
		}
		tag.Data = make([]byte, tag.Length)
		copy(tag.Data, img)
		obj.SetTag(index, tag)
	}
	return nil
}

func (obj *DcmObj) uncompress(i int, img []byte, size uint32, frames uint32, bitsa uint16, PhotoInt string) error {
	var j, offset, single uint32
	single = size / frames

	obj.DelTag(i + 1) // Delete offset table.
	switch obj.TransferSyntax.UID {
	case transfersyntax.RLELossless.UID:
		for j = 0; j < frames; j++ {
			offset = j * single
			tag := obj.GetTagAt(i + 1)
			if err := transcoder.RLEdecode(tag.Data, img[offset:], tag.Length, single, PhotoInt); err != nil {
				return err
			}
			obj.DelTag(i + 1)
		}
		obj.DelTag(i + 1)
	case transfersyntax.JPEGLosslessSV1.UID:
		fallthrough
	case transfersyntax.JPEGLossless.UID:
		for j = 0; j < frames; j++ {
			offset = j * single
			tag := obj.GetTagAt(i + 1)
			if bitsa == 8 {
				if err := jpeglib.DIJG8decode(tag.Data, tag.Length, img[offset:], single); err != nil {
					return err
				}
			} else {
				if err := jpeglib.DIJG16decode(tag.Data, tag.Length, img[offset:], single); err != nil {
					return err
				}
			}
			obj.DelTag(i + 1)
		}
		obj.DelTag(i + 1)
	case transfersyntax.JPEGBaseline8Bit.UID:
		for j = 0; j < frames; j++ {
			offset = j * single
			tag := obj.GetTagAt(i + 1)
			if bitsa == 8 {
				if err := jpeglib.DIJG8decode(tag.Data, tag.Length, img[offset:], single); err != nil {
					return err
				}
			} else {
				if err := jpeglib.DIJG12decode(tag.Data, tag.Length, img[offset:], single); err != nil {
					return err
				}
			}
			obj.DelTag(i + 1)
		}
		obj.DelTag(i + 1)
	case transfersyntax.JPEGExtended12Bit.UID:
		for j = 0; j < frames; j++ {
			offset = j * single
			tag := obj.GetTagAt(i + 1)
			if err := jpeglib.DIJG12decode(tag.Data, tag.Length, img[offset:], single); err != nil {
				return err
			}
			obj.DelTag(i + 1)
		}
		obj.DelTag(i + 1)
	case transfersyntax.JPEG2000Lossless.UID:
		for j = 0; j < frames; j++ {
			offset = j * single
			tag := obj.GetTagAt(i + 1)

			if err := transfersyntax.JPEG2000Lossless.Decode(tag.Data, tag.Length, img[offset:]); err != nil {
				return err
			}
			obj.DelTag(i + 1)
		}
		obj.DelTag(i + 1)
	case transfersyntax.JPEG2000.UID:
		for j = 0; j < frames; j++ {
			offset = j * single
			tag := obj.GetTagAt(i + 1)
			if err := transfersyntax.JPEG2000.Decode(tag.Data, tag.Length, img[offset:]); err != nil {
				return err
			}
			obj.DelTag(i + 1)
		}
		obj.DelTag(i + 1)
	}
	return nil
}
