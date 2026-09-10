package gitinfo

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestGitHubSlug(t *testing.T) {
	cases := []struct {
		in, owner, repo string
	}{
		{"git@github.com:ttyzero/ttybus.git", "ttyzero", "ttybus"},
		{"https://github.com/ttyzero/ttybus.git", "ttyzero", "ttybus"},
		{"https://github.com/ttyzero/ttybus", "ttyzero", "ttybus"},
		{"ssh://git@github.com/ttyzero/ttybus.git", "ttyzero", "ttybus"},
		{"https://gitlab.com/foo/bar.git", "", ""},
		{"git@gitlab.com:foo/bar.git", "", ""},
		{"", "", ""},
	}
	for _, c := range cases {
		o, r := GitHubSlug(c.in)
		if o != c.owner || r != c.repo {
			t.Errorf("GitHubSlug(%q) = %s/%s, want %s/%s", c.in, o, r, c.owner, c.repo)
		}
	}
}

func TestInspectRepo(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=gitwing",
			"GIT_AUTHOR_EMAIL=gitwing@example.com",
			"GIT_COMMITTER_NAME=gitwing",
			"GIT_COMMITTER_EMAIL=gitwing@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "main")
	run("config", "user.email", "gitwing@example.com")
	run("config", "user.name", "gitwing")
	run("remote", "add", "origin", "git@github.com:tjstebbing/ttybus.git")

	readme := filepath.Join(dir, "README.md")
	if err := os.WriteFile(readme, []byte("hello\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "README.md")
	run("commit", "-m", "init")

	if err := os.WriteFile(readme, []byte("hello\nworld\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	info, err := Inspect(readme)
	if err != nil {
		t.Fatal(err)
	}
	wantRoot, err := filepath.EvalSymlinks(dir)
	if err != nil {
		wantRoot = dir
	}
	if info.Root != wantRoot {
		t.Fatalf("root %q want %q", info.Root, wantRoot)
	}
	if info.Branch != "main" {
		t.Fatalf("branch %q", info.Branch)
	}
	if info.Owner != "tjstebbing" || info.Repo != "ttybus" {
		t.Fatalf("slug %s/%s", info.Owner, info.Repo)
	}
	if !info.Dirty {
		t.Fatal("expected dirty")
	}
	if info.File.Name != "README.md" || !info.File.InRepo {
		t.Fatalf("file %+v", info.File)
	}
	if info.File.Added < 1 {
		t.Fatalf("expected +lines, got %+v", info.File)
	}
}

func TestInspectNotRepo(t *testing.T) {
	dir := t.TempDir()
	_, err := Inspect(dir)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestCommitHeat(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not installed")
	}
	dir := t.TempDir()
	cmd := func(args ...string) {
		t.Helper()
		c := exec.Command("git", args...)
		c.Dir = dir
		c.Env = append(os.Environ(),
			"GIT_AUTHOR_NAME=gitwing",
			"GIT_AUTHOR_EMAIL=gitwing@example.com",
			"GIT_COMMITTER_NAME=gitwing",
			"GIT_COMMITTER_EMAIL=gitwing@example.com",
		)
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	cmd("init", "-b", "main")
	cmd("config", "user.email", "gitwing@example.com")
	cmd("config", "user.name", "gitwing")
	p := filepath.Join(dir, "a.txt")
	if err := os.WriteFile(p, []byte("a\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	cmd("add", "a.txt")
	cmd("commit", "-m", "today")
	heat := CommitHeat(dir, 14)
	if len(heat) != 14 {
		t.Fatalf("len %d", len(heat))
	}
	if heat[0] < 1 {
		t.Fatalf("expected a commit today, got %v", heat[:7])
	}
}

func TestParseNumstat(t *testing.T) {
	a, d := parseNumstat("4\t1\tcmd/main.go")
	if a != 4 || d != 1 {
		t.Fatalf("%d %d", a, d)
	}
	a, d = parseNumstat("-\t-\tbin.dat")
	if a != 0 || d != 0 {
		t.Fatalf("binary %d %d", a, d)
	}
}
