package tui

import "testing"

func TestFormatSize(t *testing.T) {
	tests := []struct {
		bytes int64
		want  string
	}{
		{0, "0 B"},
		{1023, "1023 B"},
		{1024, "1.0 KiB"},
		{1536, "1.5 KiB"},
		{1024 * 1024, "1.0 MiB"},
	}
	for _, test := range tests {
		if got := formatSize(test.bytes); got != test.want {
			t.Errorf("formatSize(%d) = %q, want %q", test.bytes, got, test.want)
		}
	}
}
