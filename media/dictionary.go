package media

import (
	"encoding/xml"
	"os"
	"strconv"

	"github.com/one-byte-data/obd-dicom/dictionary/tags"
)

type dictionary struct {
	XMLName xml.Name `xml:"dictionary"`
	Tags    []xmlTag `xml:"tag"`
}

type xmlTag struct {
	Group       string `xml:"group,attr"`
	Element     string `xml:"element,attr"`
	Name        string `xml:"keyword,attr"`
	VR          string `xml:"vr,attr"`
	VM          string `xml:"vm,attr"`
	Description string `xml:",chardata"`
}
type tagKey struct {
	group   uint16
	element uint16
}

var codes map[tagKey]*tags.Tag

// FillTag - Populates with data from dictionary
func FillTag(tag *DcmTag) {
	dt := getDictionaryTag(tag.Group, tag.Element)
	if tag.Name == "" {
		tag.Name = dt.Name
	}
	if tag.Description == "" {
		tag.Description = dt.Description
	}
	if tag.VR == "" {
		tag.VR = dt.VR
	}
	if tag.VM == "" {
		tag.VM = dt.VM
	}
}

// getDictionaryTag - get tag from Dictionary
func getDictionaryTag(group uint16, element uint16) *tags.Tag {
	if codes == nil {
		InitDict()
	}
	if t, ok := codes[tagKey{group: group, element: element}]; ok {
		return t
	}
	return &tags.Tag{
		Group:       0,
		Element:     0,
		VR:          "UN",
		VM:          "",
		Name:        "Unknown",
		Description: "Unknown",
	}
}

// getDictionaryVR - get info from Dictionary
func getDictionaryVR(group uint16, element uint16) string {
	if codes == nil {
		InitDict()
	}
	if t, ok := codes[tagKey{group: group, element: element}]; ok {
		return t.VR
	}
	return "UN"
}

func loadPrivateDictionary() {
	privateDictionaryFile := "./private.xml"
	data, err := os.ReadFile(privateDictionaryFile)
	if err != nil {
		return
	}

	dict := new(dictionary)
	err = xml.Unmarshal(data, dict)
	if err != nil {
		return
	}

	for _, t := range dict.Tags {
		g, err := strconv.Atoi(t.Group)
		if err != nil {
			continue
		}
		e, err := strconv.Atoi(t.Element)
		if err != nil {
			continue
		}
		group := uint16(g)
		element := uint16(e)
		codes[tagKey{group: group, element: element}] = &tags.Tag{
			Group:       group,
			Element:     element,
			Name:        t.Name,
			Description: t.Description,
			VR:          t.VR,
			VM:          t.VM,
		}
	}
}

// InitDict Initialize Dictionary
func InitDict() {
	tagList := tags.GetTags()
	codes = make(map[tagKey]*tags.Tag, len(tagList))
	for _, t := range tagList {
		codes[tagKey{group: t.Group, element: t.Element}] = t
	}
	loadPrivateDictionary()
}
