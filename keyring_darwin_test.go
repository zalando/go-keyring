//go:build darwin

package keyring

import (
	"strings"
	"testing"
)

// setCommand must accept commands up to 4095 bytes and reject longer ones:
// security -i reads at most 4095 bytes of a line, newline included (#153).
func TestSetCommandMaxLength(t *testing.T) {
	command, err := setCommandOfLen(t, 4095)
	if err != nil || len(command) != 4095 {
		t.Fatalf("got %d bytes, %v; want 4095 bytes, no error", len(command), err)
	}
	if _, err := setCommandOfLen(t, 4096); err != ErrSetDataTooBig {
		t.Fatalf("4096 bytes: got %v, want ErrSetDataTooBig", err)
	}
}

// setCommandOfLen calls setCommand with a service name padded so that the
// command is n bytes long.
func setCommandOfLen(t *testing.T, n int) (string, error) {
	t.Helper()
	short, err := setCommand("s", "user", "password")
	if err != nil {
		t.Fatal(err)
	}
	// Everything in the command except the one-character service name.
	overhead := len(short) - 1
	return setCommand(strings.Repeat("s", n-overhead), "user", "password")
}
