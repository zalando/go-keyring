package keyring

import "testing"

func TestSecurityInteractiveQuote(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{in: `a\b`, want: `"a\\b"`},
		{in: `a'b`, want: `"a'b"`},
		{in: `a"b`, want: `"a\"b"`},
		{in: `a b`, want: `"a b"`},
		{in: "foo", want: `"foo"`},
		{in: "", want: `""`},
		{in: `a\`, want: `"a\\"`},
	}
	for _, tc := range cases {
		got := securityInteractiveQuote(tc.in)
		if got != tc.want {
			t.Errorf("securityInteractiveQuote(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}
