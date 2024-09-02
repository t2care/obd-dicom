package media

import (
	"testing"

	"github.com/one-byte-data/obd-dicom/dictionary/tags"
)

func TestTags(t *testing.T) {
	for _, tag := range tags.GetTags() {
		got := getDictionaryTag(tag.Group, tag.Element).Name
		if got != tag.Name {
			t.Errorf("Mismatch tag. Want %v, Got %v", tag.Name, got)
		}
	}
}
