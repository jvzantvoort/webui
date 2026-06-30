package utils

import (
	"os"
	"testing"
)

func TestFileExists(t *testing.T) {
	f, err := os.CreateTemp("", "webui-test-*")
	if err != nil {
		t.Fatal(err)
	}
	f.Close()
	defer os.Remove(f.Name())

	if !FileExists(f.Name()) {
		t.Errorf("FileExists(%q) = false, want true", f.Name())
	}
	if FileExists("/nonexistent/path/xyz") {
		t.Error("FileExists('/nonexistent/path/xyz') = true, want false")
	}
}

func TestAbsPath(t *testing.T) {
	tests := []struct {
		base, rel, want string
	}{
		{"/base", "rel/path", "/base/rel/path"},
		{"/base", "/abs/path", "/abs/path"},
		{"/base", ".", "/base"},
	}
	for _, tt := range tests {
		got := AbsPath(tt.base, tt.rel)
		if got != tt.want {
			t.Errorf("AbsPath(%q, %q) = %q, want %q", tt.base, tt.rel, got, tt.want)
		}
	}
}
