package github

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const ghAuthTimeout = 2 * time.Second

// ResolveToken finds a GitHub token for private repos, first match wins:
//
//  1. $GITHUB_TOKEN
//  2. $GH_TOKEN
//  3. `gh auth token` (gh auth login / macOS keyring). Looks on PATH, then
//     /opt/homebrew/bin/gh and /usr/local/bin/gh so a slim tmux PATH still works.
//  4. oauth_token in $GH_CONFIG_DIR/hosts.yml or ~/.config/gh/hosts.yml
//     (older gh that stored the token in plaintext)
func ResolveToken() string {
	if v := strings.TrimSpace(os.Getenv("GITHUB_TOKEN")); v != "" {
		return v
	}
	if v := strings.TrimSpace(os.Getenv("GH_TOKEN")); v != "" {
		return v
	}
	if v := ghCLIToken(); v != "" {
		return v
	}
	if v := hostsFileToken(); v != "" {
		return v
	}
	return ""
}

func ghCLIToken() string {
	bin := lookGH()
	if bin == "" {
		return ""
	}
	ctx, cancel := context.WithTimeout(context.Background(), ghAuthTimeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, bin, "auth", "token")
	cmd.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1")
	cmd.Stdin = nil
	var out, errBuf bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errBuf
	if err := cmd.Run(); err != nil {
		return ""
	}
	return strings.TrimSpace(out.String())
}

func lookGH() string {
	if p, err := exec.LookPath("gh"); err == nil {
		return p
	}
	var extras []string
	if home, err := os.UserHomeDir(); err == nil {
		extras = append(extras, filepath.Join(home, ".local", "bin", "gh"))
	}
	extras = append(extras,
		"/opt/homebrew/bin/gh",
		"/usr/local/bin/gh",
	)
	for _, p := range extras {
		if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
			return p
		}
	}
	return ""
}

func hostsFileToken() string {
	path := ghHostsPath()
	if path == "" {
		return ""
	}
	b, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return tokenFromHostsYML(b)
}

func ghHostsPath() string {
	if d := os.Getenv("GH_CONFIG_DIR"); d != "" {
		return filepath.Join(d, "hosts.yml")
	}
	if d := os.Getenv("XDG_CONFIG_HOME"); d != "" {
		return filepath.Join(d, "gh", "hosts.yml")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".config", "gh", "hosts.yml")
}

// tokenFromHostsYML pulls the first non-empty oauth_token under github.com.
func tokenFromHostsYML(b []byte) string {
	inGitHub := false
	for _, line := range strings.Split(string(b), "\n") {
		trim := strings.TrimSpace(line)
		if trim == "" || strings.HasPrefix(trim, "#") {
			continue
		}
		indent := len(line) - len(strings.TrimLeft(line, " \t"))
		if indent == 0 && strings.HasSuffix(trim, ":") {
			host := strings.TrimSuffix(trim, ":")
			inGitHub = host == "github.com"
			continue
		}
		if !inGitHub {
			continue
		}
		key, val, ok := strings.Cut(trim, ":")
		if !ok || strings.TrimSpace(key) != "oauth_token" {
			continue
		}
		val = strings.Trim(strings.TrimSpace(val), `"'`)
		if val != "" && val != "null" {
			return val
		}
	}
	return ""
}
