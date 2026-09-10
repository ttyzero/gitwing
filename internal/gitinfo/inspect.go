package gitinfo

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const gitTimeout = 2 * time.Second

// Info is a snapshot of a worktree and (optionally) one file inside it.
type Info struct {
	Root      string
	Branch    string
	SHA       string
	Detached  bool
	Dirty     bool
	Ahead     int
	Behind    int
	Staged    int
	Unstaged  int
	Untracked int
	RemoteURL string
	Owner     string
	Repo      string
	Folder    string
	File      File
}

// File is the selected path relative to the repo root.
type File struct {
	Abs       string
	Rel       string
	Name      string
	Status    string // porcelain XY, trimmed; "?" untracked; "" clean/unknown
	Added     int
	Deleted   int
	Untracked bool
	InRepo    bool
	IsDir     bool
}

// Inspect walks from path (file, dir, or cwd) to a git worktree.
func Inspect(path string) (Info, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		wd, err := os.Getwd()
		if err != nil {
			return Info{}, err
		}
		path = wd
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return Info{}, err
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		abs = resolved
	}

	start := abs
	fi, statErr := os.Stat(abs)
	if statErr == nil && !fi.IsDir() {
		start = filepath.Dir(abs)
	}

	ctx, cancel := context.WithTimeout(context.Background(), gitTimeout)
	defer cancel()

	root, err := git(ctx, start, "rev-parse", "--show-toplevel")
	if err != nil {
		return Info{}, fmt.Errorf("not a git repo")
	}
	root = filepath.Clean(root)
	if resolved, err := filepath.EvalSymlinks(root); err == nil {
		root = resolved
	}

	info := Info{
		Root:   root,
		Folder: filepath.Base(root),
	}

	if sha, err := git(ctx, root, "rev-parse", "--short", "HEAD"); err == nil {
		info.SHA = sha
	}
	if br, err := git(ctx, root, "rev-parse", "--abbrev-ref", "HEAD"); err == nil {
		info.Branch = br
		if br == "HEAD" {
			info.Detached = true
		}
	}
	if url, err := git(ctx, root, "remote", "get-url", "origin"); err == nil {
		info.RemoteURL = url
		info.Owner, info.Repo = GitHubSlug(url)
	}

	status, err := git(ctx, root, "status", "--porcelain=v1", "-b", "--untracked-files=all")
	if err == nil {
		parseStatus(&info, status)
	}

	info.File = File{Abs: abs}
	if statErr == nil {
		info.File.IsDir = fi.IsDir()
	}
	rel, relErr := filepath.Rel(root, abs)
	if relErr == nil && rel != "." && !strings.HasPrefix(rel, "..") {
		info.File.InRepo = true
		info.File.Rel = rel
		info.File.Name = filepath.Base(rel)
		applyFileStatus(&info, status)
		if !info.File.IsDir && !info.File.Untracked {
			if ns, err := git(ctx, root, "diff", "--numstat", "HEAD", "--", rel); err == nil {
				info.File.Added, info.File.Deleted = parseNumstat(ns)
			}
		}
	} else if relErr == nil && rel == "." {
		info.File.InRepo = true
		info.File.IsDir = true
		info.File.Name = filepath.Base(root)
		info.File.Rel = "."
	}

	return info, nil
}

var (
	reAhead  = regexp.MustCompile(`ahead (\d+)`)
	reBehind = regexp.MustCompile(`behind (\d+)`)
)

func parseStatus(info *Info, out string) {
	for _, line := range strings.Split(out, "\n") {
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "## ") {
			rest := strings.TrimPrefix(line, "## ")
			if strings.HasPrefix(rest, "HEAD (no branch)") {
				info.Detached = true
				if info.Branch == "" {
					info.Branch = "HEAD"
				}
			}
			if m := reAhead.FindStringSubmatch(rest); len(m) == 2 {
				info.Ahead, _ = strconv.Atoi(m[1])
			}
			if m := reBehind.FindStringSubmatch(rest); len(m) == 2 {
				info.Behind, _ = strconv.Atoi(m[1])
			}
			continue
		}
		if len(line) < 3 {
			continue
		}
		xy := line[:2]
		switch {
		case xy == "??":
			info.Untracked++
		case xy == "!!":
			// ignored
		default:
			if xy[0] != ' ' && xy[0] != '?' {
				info.Staged++
			}
			if xy[1] != ' ' && xy[1] != '?' {
				info.Unstaged++
			}
		}
	}
	info.Dirty = info.Staged+info.Unstaged+info.Untracked > 0
}

func applyFileStatus(info *Info, out string) {
	rel := info.File.Rel
	want := filepath.ToSlash(rel)
	for _, line := range strings.Split(out, "\n") {
		if len(line) < 4 || strings.HasPrefix(line, "## ") {
			continue
		}
		xy := line[:2]
		name := strings.TrimSpace(line[3:])
		if i := strings.Index(name, " -> "); i >= 0 {
			name = name[i+4:]
		}
		name = strings.Trim(name, `"`)
		if filepath.ToSlash(name) != want {
			continue
		}
		info.File.Status = strings.TrimSpace(xy)
		if xy == "??" {
			info.File.Untracked = true
			info.File.Status = "?"
		}
		return
	}
	if info.File.InRepo && !info.File.IsDir {
		info.File.Status = ""
	}
}

func parseNumstat(out string) (added, deleted int) {
	out = strings.TrimSpace(out)
	if out == "" {
		return 0, 0
	}
	// First line: added \t deleted \t path  ("-" for binary)
	f := strings.Split(out, "\t")
	if len(f) < 2 {
		return 0, 0
	}
	if f[0] != "-" {
		added, _ = strconv.Atoi(f[0])
	}
	if f[1] != "-" {
		deleted, _ = strconv.Atoi(f[1])
	}
	return added, deleted
}

func git(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "GIT_OPTIONAL_LOCKS=0")
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", args[0], msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}
