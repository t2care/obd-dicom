package media

import (
	"testing"
)

func TestTagSorting(t *testing.T) {
	InitDict()

	type DcmTags struct {
		tagA DcmTag
		tagB DcmTag
	}

	tests := []struct {
		name     string
		args     DcmTags
		expected bool
	}{
		{
			name: "sorted tags, higher element id",
			args: DcmTags{tagA: DcmTag{Group: 0x0000, Element: 0x0000},
				tagB: DcmTag{Group: 0x0000, Element: 0x0001},
			},
			expected: true,
		},
		{
			name: "sorted tags, higher group & element ids",
			args: DcmTags{tagA: DcmTag{Group: 0x0000, Element: 0x0000},
				tagB: DcmTag{Group: 0x0009, Element: 0x0004},
			},
			expected: true,
		},
		{
			name: "unsorted tags",
			args: DcmTags{tagA: DcmTag{Group: 0x0100, Element: 0x0005},
				tagB: DcmTag{Group: 0x0000, Element: 0x0001},
			},
			expected: false,
		},
		{
			name: "same group, same element",
			args: DcmTags{tagA: DcmTag{Group: 0x0777, Element: 0x0042},
				tagB: DcmTag{Group: 0x0777, Element: 0x0042},
			},
			expected: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ret := tt.args.tagA.isBefore(&tt.args.tagB)

			if ret != tt.expected {
				t.Errorf("DcmTag.isBefore() returns %v, expected %v", ret, tt.expected)
				return
			}

		})
	}
}
