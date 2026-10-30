package attachments

import "testing"

func TestValidateUpload(t *testing.T) {
	valid := []struct {
		name, mime string
		size       int64
	}{{"photo.png", "image/png", 10}, {"clip.mp4", "video/mp4", 100}}
	for _, v := range valid {
		if err := ValidateUpload(v.name, v.mime, v.size); err != nil {
			t.Errorf("valid upload rejected: %v", err)
		}
	}
	invalid := []struct {
		name, mime string
		size       int64
	}{{"../x.png", "image/png", 10}, {"x.exe", "application/x-executable", 10}, {"x.png", "image/png", 0}, {"x.png", "image/png", MaxUploadBytes + 1}}
	for _, v := range invalid {
		if err := ValidateUpload(v.name, v.mime, v.size); err == nil {
			t.Errorf("invalid upload accepted: %+v", v)
		}
	}
}
