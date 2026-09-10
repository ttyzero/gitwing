package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"
)

const cacheTTL = 5 * time.Minute

// Repo is the slice of GitHub metadata gitwing shows.
type Repo struct {
	FullName    string `json:"full_name"`
	Name        string `json:"name"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	Language    string `json:"language"`
	Stars       int    `json:"stargazers_count"`
	Forks       int    `json:"forks_count"`
	Watchers    int    `json:"subscribers_count"`
	OpenIssues  int    `json:"open_issues_count"`
	Private     bool   `json:"private"`
	Archived    bool   `json:"archived"`
	Homepage    string `json:"homepage"`
}

type cacheEntry struct {
	repo Repo
	err  error
	at   time.Time
}

// APIError is a non-200 from api.github.com.
type APIError struct {
	Status int
	Body   string
}

func (e APIError) Error() string {
	if e.Body == "" {
		return fmt.Sprintf("github %d", e.Status)
	}
	return fmt.Sprintf("github %d: %s", e.Status, e.Body)
}

// Short is a compact UI label for a fetch error.
func Short(err error) string {
	if err == nil {
		return ""
	}
	var e APIError
	if errors.As(err, &e) {
		switch e.Status {
		case http.StatusUnauthorized, http.StatusForbidden, http.StatusNotFound:
			return "private"
		}
	}
	return "offline"
}

// Client fetches repo metadata with a small in-memory cache.
type Client struct {
	HTTP    *http.Client
	BaseURL string
	Token   string // explicit override; empty means ResolveToken on first use
	UA      string

	mu            sync.Mutex
	cache         map[string]cacheEntry
	resolvedToken string
	tokenResolved bool
}

// New uses api.github.com. The token is resolved lazily (env, then gh).
func New() *Client {
	return &Client{
		HTTP:    &http.Client{Timeout: 4 * time.Second},
		BaseURL: "https://api.github.com",
		UA:      "gitwing/0.1",
		cache:   map[string]cacheEntry{},
	}
}

func (c *Client) token() string {
	if c.Token != "" {
		return c.Token
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.tokenResolved {
		c.resolvedToken = ResolveToken()
		c.tokenResolved = true
	}
	return c.resolvedToken
}

// Forget drops a cached slug so the next Fetch hits the network.
func (c *Client) Forget(owner, repo string) {
	c.mu.Lock()
	delete(c.cache, owner+"/"+repo)
	c.mu.Unlock()
}

// Fetch owner/repo. Cached for five minutes, including errors (for 20s)
// so a missing network does not hammer the API.
func (c *Client) Fetch(owner, repo string) (Repo, error) {
	if owner == "" || repo == "" {
		return Repo{}, fmt.Errorf("empty slug")
	}
	key := owner + "/" + repo
	c.mu.Lock()
	if e, ok := c.cache[key]; ok {
		ttl := cacheTTL
		if e.err != nil {
			ttl = 20 * time.Second
		}
		if time.Since(e.at) < ttl {
			c.mu.Unlock()
			return e.repo, e.err
		}
	}
	c.mu.Unlock()

	r, err := c.get(owner, repo)
	c.mu.Lock()
	c.cache[key] = cacheEntry{repo: r, err: err, at: time.Now()}
	c.mu.Unlock()
	return r, err
}

func (c *Client) get(owner, repo string) (Repo, error) {
	url := c.BaseURL + "/repos/" + owner + "/" + repo
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return Repo{}, err
	}
	req.Header.Set("User-Agent", c.UA)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if tok := c.token(); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	res, err := c.HTTP.Do(req)
	if err != nil {
		return Repo{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(res.Body, 1<<20))
	if res.StatusCode != http.StatusOK {
		return Repo{}, APIError{Status: res.StatusCode, Body: brief(body)}
	}
	var r Repo
	if err := json.Unmarshal(body, &r); err != nil {
		return Repo{}, err
	}
	return r, nil
}

func brief(b []byte) string {
	s := string(b)
	if len(s) > 120 {
		return s[:120]
	}
	return s
}
