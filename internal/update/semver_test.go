package update

import "testing"

func TestParseVersion(t *testing.T) {
	tests := []struct {
		input                    string
		major, minor, patch      int
	}{
		{"0.1.0", 0, 1, 0},
		{"1.2.3", 1, 2, 3},
		{"v0.1.0", 0, 1, 0},
		{"v10.20.30", 10, 20, 30},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			maj, min, pat, err := parseVersion(tt.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if maj != tt.major || min != tt.minor || pat != tt.patch {
				t.Errorf("got %d.%d.%d, want %d.%d.%d", maj, min, pat, tt.major, tt.minor, tt.patch)
			}
		})
	}
}

func TestParseVersion_Invalid(t *testing.T) {
	tests := []string{
		"",
		"1",
		"1.2",
		"1.2.3.4",
		"a.b.c",
		"1.2.x",
	}

	for _, input := range tests {
		t.Run(input, func(t *testing.T) {
			_, _, _, err := parseVersion(input)
			if err == nil {
				t.Error("expected error, got nil")
			}
		})
	}
}

func TestIsNewer(t *testing.T) {
	tests := []struct {
		name    string
		latest  string
		current string
		want    bool
	}{
		{"major bump", "2.0.0", "1.0.0", true},
		{"minor bump", "0.2.0", "0.1.0", true},
		{"patch bump", "0.1.1", "0.1.0", true},
		{"same version", "0.1.0", "0.1.0", false},
		{"older major", "0.1.0", "1.0.0", false},
		{"older minor", "0.1.0", "0.2.0", false},
		{"older patch", "0.1.0", "0.1.1", false},
		{"with v prefix", "v0.2.0", "0.1.0", true},
		{"both with v prefix", "v0.2.0", "v0.1.0", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := isNewer(tt.latest, tt.current)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("isNewer(%q, %q) = %v, want %v", tt.latest, tt.current, got, tt.want)
			}
		})
	}
}
