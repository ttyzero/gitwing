package ui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"

	"github.com/ttyzero/commons/theme"
	"github.com/ttyzero/gitwing/internal/github"
	"github.com/ttyzero/gitwing/internal/gitinfo"
)

func TestRenderFits20(t *testing.T) {
	m := model{
		pal:    theme.New(true),
		width:  20,
		height: 24,
		busOK:  true,
		git: gitinfo.Info{
			Root:   "/tmp/ttybus",
			Branch: "main",
			SHA:    "a5af5fe",
			Dirty:  true,
			Folder: "ttybus",
			Owner:  "tjstebbing",
			Repo:   "ttybus",
			File: gitinfo.File{
				InRepo:  true,
				Name:    "README.md",
				Rel:     "README.md",
				Status:  "M ",
				Added:   8,
				Deleted: 2,
			},
		},
		gh: github.Repo{
			FullName:    "tjstebbing/ttybus",
			Name:        "ttybus",
			Description: "A TUI bus for TUIs and CLIs to talk.",
			Language:    "Go",
			Stars:       12,
			Forks:       3,
			HTMLURL:     "https://github.com/ttyzero/ttybus",
		},
	}
	out := m.render()
	t.Log("\n" + out)
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 20 {
			t.Errorf("line %d width %d > 20: %q", i, w, line)
		}
	}
	plain := out
	for _, need := range []string{"gitwing", "ttybus", "main", "README.md", "+8"} {
		if !strings.Contains(plain, need) {
			t.Errorf("missing %q", need)
		}
	}
}

func TestRenderBorderless(t *testing.T) {
	m := model{
		pal:        theme.New(true),
		borderless: true,
		width:      20,
		height:     16,
		busOK:      true,
	}
	out := m.render()
	if strings.Contains(out, "╭") || strings.Contains(out, "╰") {
		t.Fatalf("unexpected border:\n%s", out)
	}
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 20 {
			t.Errorf("line %d width %d > 20", i, w)
		}
	}
}

func TestRenderIdle(t *testing.T) {
	m := model{
		pal:    theme.New(true),
		width:  20,
		height: 16,
		busOK:  true,
	}
	out := m.render()
	if !strings.Contains(out, "listening") {
		t.Fatalf("%s", out)
	}
	for i, line := range strings.Split(out, "\n") {
		if w := lipgloss.Width(line); w > 20 {
			t.Errorf("line %d width %d > 20: %q", i, w, line)
		}
	}
}
