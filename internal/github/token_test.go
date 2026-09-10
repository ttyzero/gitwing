package github

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveTokenEnv(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "from-github-token")
	t.Setenv("GH_TOKEN", "from-gh-token")
	t.Setenv("GH_CONFIG_DIR", t.TempDir())
	if got := ResolveToken(); got != "from-github-token" {
		t.Fatalf("got %q", got)
	}
	t.Setenv("GITHUB_TOKEN", "")
	if got := ResolveToken(); got != "from-gh-token" {
		t.Fatalf("GH_TOKEN: %q", got)
	}
}

func TestTokenFromHostsYML(t *testing.T) {
	plain := []byte(`github.com:
    user: tjstebbing
    oauth_token: gho_plain
    git_protocol: ssh
`)
	if got := tokenFromHostsYML(plain); got != "gho_plain" {
		t.Fatalf("%q", got)
	}

	nested := []byte(`enterprise.example:
    oauth_token: gho_wrong
github.com:
    users:
        tjstebbing:
            oauth_token: gho_nested
    user: tjstebbing
`)
	if got := tokenFromHostsYML(nested); got != "gho_nested" {
		t.Fatalf("nested %q", got)
	}

	keyring := []byte(`github.com:
    git_protocol: ssh
    user: tjstebbing
`)
	if got := tokenFromHostsYML(keyring); got != "" {
		t.Fatal("keyring-only config should not invent a token")
	}
}

func TestResolveTokenHostsFile(t *testing.T) {
	t.Setenv("GITHUB_TOKEN", "")
	t.Setenv("GH_TOKEN", "")
	dir := t.TempDir()
	t.Setenv("GH_CONFIG_DIR", dir)
	path := filepath.Join(dir, "hosts.yml")
	if err := os.WriteFile(path, []byte("github.com:\n    oauth_token: gho_file\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	// gh CLI may still win if installed; hosts.yml is last resort.
	got := tokenFromHostsYML([]byte("github.com:\n    oauth_token: gho_file\n"))
	if got != "gho_file" {
		t.Fatalf("%q", got)
	}
}

func TestFetchSendsToken(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		w.Write([]byte(`{"name":"navehnet","full_name":"tjstebbing/navehnet","private":true}`))
	}))
	defer srv.Close()
	c := New()
	c.HTTP = srv.Client()
	c.BaseURL = srv.URL
	c.Token = "gho_test"
	c.cache = map[string]cacheEntry{}
	r, err := c.Fetch("tjstebbing", "navehnet")
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "navehnet" || !r.Private {
		t.Fatalf("%+v", r)
	}
	if gotAuth != "Bearer gho_test" {
		t.Fatalf("auth %q", gotAuth)
	}
}

func TestShortPrivate(t *testing.T) {
	if Short(APIError{Status: 404}) != "private" {
		t.Fatal(Short(APIError{Status: 404}))
	}
	if Short(APIError{Status: 401}) != "private" {
		t.Fatal(Short(APIError{Status: 401}))
	}
	if Short(APIError{Status: 500}) != "offline" {
		t.Fatal(Short(APIError{Status: 500}))
	}
}

func TestFetch404(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"message":"Not Found"}`, 404)
	}))
	defer srv.Close()
	c := New()
	c.HTTP = srv.Client()
	c.BaseURL = srv.URL
	c.Token = ""
	c.tokenResolved = true // skip gh
	c.cache = map[string]cacheEntry{}
	_, err := c.Fetch("tjstebbing", "navehnet")
	if Short(err) != "private" {
		t.Fatalf("%v", err)
	}
}
