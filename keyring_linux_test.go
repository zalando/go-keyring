//go:build linux

package keyring

import (
	"os"
	"reflect"
	"testing"
)

func TestEscapePowerShellString(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"no quotes", "abc", "abc"},
		{"single quote", "a'b", "a''b"},
		{"multiple quotes", "'''", "''''''"},
		{"leading and trailing", "'x'", "''x''"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := escapePowerShellString(tc.in); got != tc.want {
				t.Errorf("escapePowerShellString(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestUTF16LERoundTrip(t *testing.T) {
	cases := []string{
		"",
		"hello",
		"pässwörd",
		"line1\nline2",
		"emoji: \U0001F600",
		"mixed 日本語 text",
	}
	for _, in := range cases {
		t.Run(in, func(t *testing.T) {
			encoded := utf16LEEncode(in)
			if len(encoded)%2 != 0 {
				t.Fatalf("encoded length %d is not even", len(encoded))
			}
			if got := utf16LEDecode(encoded); got != in {
				t.Errorf("round trip = %q, want %q", got, in)
			}
		})
	}
}

func TestUTF16LEEncodeKnown(t *testing.T) {
	// "AB" -> 0x41 0x00 0x42 0x00 in little-endian UTF-16.
	got := utf16LEEncode("AB")
	want := []byte{0x41, 0x00, 0x42, 0x00}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("utf16LEEncode(\"AB\") = %v, want %v", got, want)
	}
}

func TestUTF16LEDecodeOddByteIgnored(t *testing.T) {
	// A trailing odd byte should be ignored rather than cause a panic.
	in := []byte{0x41, 0x00, 0x42} // "A" + dangling byte
	if got := utf16LEDecode(in); got != "A" {
		t.Errorf("utf16LEDecode(%v) = %q, want %q", in, got, "A")
	}
}

func TestCleanPowerShellStderr(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want string
	}{
		{"empty", "", ""},
		{"whitespace only", "   \n\t ", ""},
		{"trimmed real error", "  boom  ", "boom"},
		{"clixml discarded", "#< CLIXML\n<Objs>...</Objs>", ""},
		{"clixml with leading space", "   #< CLIXML stuff", ""},
		{"not clixml", "error #< CLIXML", "error #< CLIXML"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := cleanPowerShellStderr(tc.in); got != tc.want {
				t.Errorf("cleanPowerShellStderr(%q) = %q, want %q", tc.in, got, tc.want)
			}
		})
	}
}

func TestCredName(t *testing.T) {
	var k wslKeychain
	cases := []struct {
		service, username, want string
	}{
		{"svc", "user", "svc:user"},
		{"svc", "", "svc:"},
		{"", "", ":"},
		{"a:b", "c", "a:b:c"},
	}
	for _, tc := range cases {
		if got := k.credName(tc.service, tc.username); got != tc.want {
			t.Errorf("credName(%q, %q) = %q, want %q", tc.service, tc.username, got, tc.want)
		}
	}
}

func TestIsWSLEnvVars(t *testing.T) {
	// Save and clear the interop env vars so the detection is deterministic.
	for _, key := range []string{"WSL_DISTRO_NAME", "WSL_INTEROP"} {
		t.Setenv(key, "")
		os.Unsetenv(key)
	}

	t.Run("WSL_DISTRO_NAME set", func(t *testing.T) {
		t.Setenv("WSL_DISTRO_NAME", "Ubuntu")
		if !IsWSL() {
			t.Error("expected IsWSL() to be true when WSL_DISTRO_NAME is set")
		}
	})

	t.Run("WSL_INTEROP set", func(t *testing.T) {
		t.Setenv("WSL_INTEROP", "/run/WSL/1_interop")
		if !IsWSL() {
			t.Error("expected IsWSL() to be true when WSL_INTEROP is set")
		}
	})
}
