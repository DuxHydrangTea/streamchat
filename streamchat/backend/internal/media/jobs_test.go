package media

import "testing"

func TestTaskNameIsStable(t *testing.T) {
	if TaskDownloadVideo != "media:download-video" {
		t.Fatal("task name changed")
	}
}

func TestParseRange(t *testing.T) {
	cases := []struct {
		name           string
		value          string
		start, end     int64
		partial, valid bool
	}{
		{"full", "", 0, 9, false, true},
		{"bounded", "bytes=2-5", 2, 5, true, true},
		{"open ended", "bytes=7-", 7, 9, true, true},
		{"suffix", "bytes=-3", 7, 9, true, true},
		{"suffix larger than file", "bytes=-20", 0, 9, true, true},
		{"out of range", "bytes=10-", 0, 0, false, false},
		{"reversed", "bytes=5-2", 0, 0, false, false},
		{"multiple", "bytes=0-1,4-5", 0, 0, false, false},
		{"empty file", "", 0, 0, false, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			start, end, partial, valid := parseRange(tc.value, 10)
			if tc.name == "empty file" {
				start, end, partial, valid = parseRange("", 0)
			}
			if start != tc.start || end != tc.end || partial != tc.partial || valid != tc.valid {
				t.Fatalf("got (%d,%d,%t,%t), want (%d,%d,%t,%t)", start, end, partial, valid, tc.start, tc.end, tc.partial, tc.valid)
			}
		})
	}
}
