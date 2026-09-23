package config

import (
	"testing"
)

func TestParseFormatVersion(t *testing.T) {
	tests := []struct {
		name    string
		raw     string
		want    FormatVersion
		wantErr bool
	}{
		{name: "valid current", raw: "1.0.0", want: FormatVersion{Major: 1, Minor: 0, Patch: 0}},
		{name: "valid multi digit", raw: "12.34.56", want: FormatVersion{Major: 12, Minor: 34, Patch: 56}},
		{name: "two parts", raw: "1.0", wantErr: true},
		{name: "four parts", raw: "1.0.0.0", wantErr: true},
		{name: "non numeric", raw: "1.x.0", wantErr: true},
		{name: "negative", raw: "-1.0.0", wantErr: true},
		{name: "empty", raw: "", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseFormatVersion(tt.raw)
			if (err != nil) != tt.wantErr {
				t.Fatalf("ParseFormatVersion(%q) error = %v, wantErr %v", tt.raw, err, tt.wantErr)
			}

			if tt.wantErr {
				return
			}

			if got != tt.want {
				t.Errorf("ParseFormatVersion(%q) = %v, want %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestNormalizeFormatVersion(t *testing.T) {
	t.Run("empty maps to current", func(t *testing.T) {
		got, err := NormalizeFormatVersion("")
		if err != nil {
			t.Fatalf("NormalizeFormatVersion(\"\") error = %v", err)
		}

		if got != CurrentFormatVersion {
			t.Errorf("NormalizeFormatVersion(\"\") = %v, want %v", got, CurrentFormatVersion)
		}
	})

	t.Run("explicit version passes through", func(t *testing.T) {
		got, err := NormalizeFormatVersion("0.9.1")
		if err != nil {
			t.Fatalf("NormalizeFormatVersion error = %v", err)
		}

		want := FormatVersion{Major: 0, Minor: 9, Patch: 1}
		if got != want {
			t.Errorf("NormalizeFormatVersion = %v, want %v", got, want)
		}
	})

	t.Run("invalid version rejected", func(t *testing.T) {
		if _, err := NormalizeFormatVersion("banana"); err == nil {
			t.Error("NormalizeFormatVersion(\"banana\") expected error, got nil")
		}
	})
}

func TestFormatVersionCompare(t *testing.T) {
	tests := []struct {
		name string
		a, b FormatVersion
		want int
	}{
		{name: "equal", a: FormatVersion{1, 2, 3}, b: FormatVersion{1, 2, 3}, want: 0},
		{name: "major less", a: FormatVersion{1, 9, 9}, b: FormatVersion{2, 0, 0}, want: -1},
		{name: "major greater", a: FormatVersion{2, 0, 0}, b: FormatVersion{1, 9, 9}, want: 1},
		{name: "minor less", a: FormatVersion{1, 1, 9}, b: FormatVersion{1, 2, 0}, want: -1},
		{name: "patch less", a: FormatVersion{1, 2, 0}, b: FormatVersion{1, 2, 1}, want: -1},
		{name: "patch greater", a: FormatVersion{1, 2, 1}, b: FormatVersion{1, 2, 0}, want: 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.a.Compare(tt.b); got != tt.want {
				t.Errorf("Compare = %d, want %d", got, tt.want)
			}

			if got := tt.a.LessThan(tt.b); got != (tt.want < 0) {
				t.Errorf("LessThan = %v, want %v", got, tt.want < 0)
			}
		})
	}
}

func TestFormatVersionString(t *testing.T) {
	version := FormatVersion{Major: 3, Minor: 14, Patch: 159}
	if got := version.String(); got != "3.14.159" {
		t.Errorf("String() = %q, want %q", got, "3.14.159")
	}

	parsed, err := ParseFormatVersion(version.String())
	if err != nil {
		t.Fatalf("roundtrip parse error = %v", err)
	}

	if parsed != version {
		t.Errorf("roundtrip = %v, want %v", parsed, version)
	}
}
