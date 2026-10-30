package conversations

import "testing"

func TestNormalizeUsernames(t *testing.T) {
	got := NormalizeUsernames([]string{" Casey_1 ", "casey_1", "", "  ", "Jordan"})
	want := []string{"casey_1", "jordan"}
	if len(got) != len(want) {
		t.Fatalf("NormalizeUsernames() = %#v, want %#v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("NormalizeUsernames() = %#v, want %#v", got, want)
		}
	}
}

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
