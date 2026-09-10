package gitinfo

import (
	"net/url"
	"strings"
)

// GitHubSlug extracts owner/repo from a git remote URL. Non-GitHub
// remotes return empty strings.
func GitHubSlug(remote string) (owner, repo string) {
	remote = strings.TrimSpace(remote)
	remote = strings.TrimSuffix(remote, ".git")
	remote = strings.TrimSuffix(remote, "/")

	if strings.HasPrefix(remote, "git@") {
		// git@github.com:owner/repo
		_, rest, ok := strings.Cut(remote, "@")
		if !ok {
			return "", ""
		}
		host, path, ok := strings.Cut(rest, ":")
		if !ok {
			return "", ""
		}
		if !isGitHubHost(host) {
			return "", ""
		}
		return splitOwnerRepo(path)
	}

	if strings.Contains(remote, "://") {
		u, err := url.Parse(remote)
		if err != nil {
			return "", ""
		}
		if !isGitHubHost(u.Hostname()) {
			return "", ""
		}
		return splitOwnerRepo(strings.TrimPrefix(u.Path, "/"))
	}

	// scp-like without git@ — rare
	if host, path, ok := strings.Cut(remote, ":"); ok && strings.Contains(host, "github") {
		if !isGitHubHost(host) {
			return "", ""
		}
		return splitOwnerRepo(path)
	}
	return "", ""
}

func isGitHubHost(host string) bool {
	host = strings.ToLower(host)
	return host == "github.com" || strings.HasSuffix(host, ".github.com")
}

func splitOwnerRepo(path string) (owner, repo string) {
	path = strings.TrimPrefix(path, "/")
	path = strings.TrimSuffix(path, ".git")
	path = strings.TrimSuffix(path, "/")
	parts := strings.Split(path, "/")
	if len(parts) < 2 {
		return "", ""
	}
	return parts[0], parts[1]
}
