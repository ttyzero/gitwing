package ui

import (
	"strings"
	"testing"

	"github.com/mattn/go-runewidth"
)

func TestCompactInt(t *testing.T) {
	cases := map[int]string{
		0:       "0",
		12:      "12",
		999:     "999",
		1200:    "1.2k",
		12000:   "12k",
		1500000: "1.5m",
	}
	for n, want := range cases {
		if got := compactInt(n); got != want {
			t.Errorf("compactInt(%d) = %q want %q", n, got, want)
		}
	}
}

func TestEllipsize(t *testing.T) {
	if got := ellipsize("hello", 10); got != "hello" {
		t.Fatalf("%q", got)
	}
	got := ellipsize("hello world", 8)
	if runewidthCheck := len([]rune(got)); runewidthCheck < 1 {
		t.Fatal("empty")
	}
	if !stringsHasEllipsis(got) {
		t.Fatalf("expected ellipsis, got %q", got)
	}
}

func stringsHasEllipsis(s string) bool {
	for _, r := range s {
		if r == '…' {
			return true
		}
	}
	return false
}

func TestWrap(t *testing.T) {
	lines := wrap("A TUI bus for TUIs and CLIs to talk.", 16, 3)
	if len(lines) != 3 {
		t.Fatalf("%q", lines)
	}
	joined := strings.Join(lines, " ")
	if !strings.Contains(joined, "talk") {
		t.Fatalf("lost tail: %q", lines)
	}
	for _, ln := range lines {
		if runewidth.StringWidth(ln) > 16 {
			t.Fatalf("wide line %q", ln)
		}
	}
}
