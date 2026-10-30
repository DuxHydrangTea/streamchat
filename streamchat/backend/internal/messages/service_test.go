package messages

import "testing"

func TestValidateText(t *testing.T) {
	cases := []struct {
		name, text string
		wantErr    bool
	}{
		{"normal", "hello", false}, {"trimmed empty", "  \n", true}, {"too long", string(make([]byte, 4001)), true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := ValidateText(tc.text)
			if (err != nil) != tc.wantErr {
				t.Fatalf("err=%v wantErr=%v", err, tc.wantErr)
			}
		})
	}
}
