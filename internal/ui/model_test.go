package ui

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"

	"github.com/ttyzero/commons/theme"
	"github.com/ttyzero/gitwing/internal/github"
	"github.com/ttyzero/gitwing/internal/gitinfo"
)

func TestInspectSwitchClearsStaleGitHub(t *testing.T) {
	m := model{
		pal:    theme.New(true),
		width:  20,
		height: 24,
		path:   "/tmp/navehnet/README.md",
		busOK:  true,
		git: gitinfo.Info{
			Root:   "/tmp/ttybus",
			Folder: "ttybus",
			Owner:  "tjstebbing",
			Repo:   "ttybus",
			Branch: "main",
		},
		gh: github.Repo{
			FullName:    "tjstebbing/ttybus",
			Name:        "ttybus",
			Description: "A TUI bus",
			Stars:       12,
			HTMLURL:     "https://github.com/ttyzero/ttybus",
		},
		ghKey: "tjstebbing/ttybus",
	}
	next, cmd := m.Update(inspectMsg{
		path: "/tmp/navehnet/README.md",
		info: gitinfo.Info{
			Root:   "/tmp/navehnet",
			Folder: "navehnet",
			Owner:  "tjstebbing",
			Repo:   "navehnet",
			Branch: "hermes/hop-service-env",
		},
	})
	got, ok := next.(model)
	if !ok {
		t.Fatalf("got %T", next)
	}
	if got.gh.Name != "" || got.gh.Stars != 0 {
		t.Fatalf("stale github still present: %+v", got.gh)
	}
	if got.git.Repo != "navehnet" {
		t.Fatalf("git repo %q", got.git.Repo)
	}
	if !got.ghLoad || got.ghKey != "tjstebbing/navehnet" {
		t.Fatalf("expected navehnet fetch, key=%q load=%v", got.ghKey, got.ghLoad)
	}
	if cmd == nil {
		t.Fatal("expected fetch cmd")
	}
	out := got.render()
	if strings.Contains(out, "ttybus") {
		t.Fatalf("still showing ttybus:\n%s", out)
	}
	if !strings.Contains(out, "navehnet") {
		t.Fatalf("missing navehnet:\n%s", out)
	}
}

func TestGitHubFetchErrorDoesNotKeepPrevious(t *testing.T) {
	m := model{
		git: gitinfo.Info{Owner: "tjstebbing", Repo: "navehnet", Folder: "navehnet"},
		gh: github.Repo{
			FullName: "tjstebbing/ttybus",
			Name:     "ttybus",
			Stars:    12,
		},
		ghKey:  "tjstebbing/navehnet",
		ghLoad: true,
	}
	next, _ := m.Update(ghMsg{key: "tjstebbing/navehnet", err: github.APIError{Status: 404}})
	got := next.(model)
	if got.gh.Name != "" {
		t.Fatalf("kept previous github card: %+v", got.gh)
	}
	if got.ghErr != "private" {
		t.Fatalf("ghErr %q", got.ghErr)
	}
}

func TestStaleGitHubMsgIgnored(t *testing.T) {
	m := model{
		git: gitinfo.Info{Owner: "tjstebbing", Repo: "navehnet"},
		gh:  github.Repo{},
	}
	next, _ := m.Update(ghMsg{
		key:  "tjstebbing/ttybus",
		repo: github.Repo{Name: "ttybus", FullName: "tjstebbing/ttybus"},
	})
	got := next.(model)
	if got.gh.Name == "ttybus" {
		t.Fatal("accepted github card for the previous repo")
	}
}

func TestSameRepoInspectDoesNotDropCard(t *testing.T) {
	m := model{
		path: "/tmp/ttybus/README.md",
		git:  gitinfo.Info{Owner: "tjstebbing", Repo: "ttybus", Folder: "ttybus"},
		gh: github.Repo{
			FullName: "tjstebbing/ttybus",
			Name:     "ttybus",
			Stars:    12,
		},
		ghKey: "tjstebbing/ttybus",
	}
	next, cmd := m.Update(inspectMsg{
		path: "/tmp/ttybus/README.md",
		info: gitinfo.Info{Owner: "tjstebbing", Repo: "ttybus", Folder: "ttybus", Branch: "main"},
	})
	got := next.(model)
	if got.gh.Stars != 12 {
		t.Fatalf("dropped live card: %+v", got.gh)
	}
	if cmd != nil {
		t.Fatal("refetched same repo")
	}
}

func TestGhLive(t *testing.T) {
	m := model{
		git: gitinfo.Info{Owner: "tjstebbing", Repo: "navehnet"},
		gh:  github.Repo{FullName: "tjstebbing/ttybus", Name: "ttybus"},
	}
	if m.ghLive() {
		t.Fatal("ttybus card should not be live for navehnet")
	}
	m.gh = github.Repo{FullName: "tjstebbing/navehnet", Name: "navehnet"}
	if !m.ghLive() {
		t.Fatal("expected live")
	}
}

var _ tea.Model = model{}
