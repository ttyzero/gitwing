package github

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetch(t *testing.T) {
	n := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n++
		if r.URL.Path != "/repos/tjstebbing/ttybus" {
			http.NotFound(w, r)
			return
		}
		if r.Header.Get("User-Agent") == "" {
			t.Error("missing User-Agent")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"full_name":         "tjstebbing/ttybus",
			"name":              "ttybus",
			"description":       "A TUI bus",
			"html_url":          "https://github.com/ttyzero/ttybus",
			"language":          "Go",
			"stargazers_count":  12,
			"forks_count":       3,
			"subscribers_count": 4,
			"open_issues_count": 1,
		})
	}))
	defer srv.Close()

	c := New()
	c.HTTP = srv.Client()
	c.BaseURL = srv.URL
	c.cache = map[string]cacheEntry{}

	r, err := c.Fetch("tjstebbing", "ttybus")
	if err != nil {
		t.Fatal(err)
	}
	if r.Name != "ttybus" || r.Stars != 12 || r.Forks != 3 || r.Description != "A TUI bus" {
		t.Fatalf("%+v", r)
	}
	_, _ = c.Fetch("tjstebbing", "ttybus")
	if n != 1 {
		t.Fatalf("cache miss, hits=%d", n)
	}
}
