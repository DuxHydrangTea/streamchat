package conversations

import "testing"

func TestValidateCreate(t *testing.T) {
	valid := []struct {
		kind, name string
		members    []string
	}{{"direct", "", []string{"u2"}}, {"group", "Team", nil}}
	for _, v := range valid {
		if err := ValidateCreate(v.kind, v.name, v.members); err != nil {
			t.Errorf("valid input rejected: %v", err)
		}
	}
	invalid := []struct {
		kind, name string
		members    []string
	}{{"unknown", "", nil}, {"group", "", nil}, {"group", string(make([]byte, 101)), nil}, {"direct", "", make([]string, 100)}}
	for _, v := range invalid {
		if err := ValidateCreate(v.kind, v.name, v.members); err == nil {
			t.Errorf("invalid input accepted: %+v", v)
		}
	}
}
