package gh

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestFetchReposFiltersByForkFlag(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user/repos" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page") == "2" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(`[]`))
			return
		}
		repos := []apiRepo{
			{FullName: "me/forked", Fork: true},
			{FullName: "me/owned", Fork: false},
		}
		_ = json.NewEncoder(w).Encode(repos)
	}))
	defer ts.Close()

	client := New(ts.URL, "token")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	forks, err := client.FetchRepos(ctx, true)
	if err != nil {
		t.Fatalf("fetch forks: %v", err)
	}
	if len(forks) != 1 || forks[0].FullName != "me/forked" {
		t.Fatalf("expected only forked repo, got %#v", forks)
	}

	owned, err := client.FetchRepos(ctx, false)
	if err != nil {
		t.Fatalf("fetch owned: %v", err)
	}
	if len(owned) != 1 || owned[0].FullName != "me/owned" {
		t.Fatalf("expected only owned repo, got %#v", owned)
	}
}

func TestDeleteRepoHandlesErrors(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/repos/me/ok":
			w.WriteHeader(http.StatusNoContent)
		case "/repos/me/forbidden":
			http.Error(w, "nope", http.StatusForbidden)
		default:
			http.NotFound(w, r)
		}
	}))
	defer ts.Close()

	client := New(ts.URL, "token")
	ctx := context.Background()

	if err := client.DeleteRepo(ctx, "me/ok"); err != nil {
		t.Fatalf("delete ok: %v", err)
	}
	if err := client.DeleteRepo(ctx, "me/forbidden"); err == nil {
		t.Fatalf("expected forbidden error")
	}
	if err := client.DeleteRepo(ctx, "me/missing"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestCurrentUser(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"login":"octocat"}`))
	}))
	defer ts.Close()

	client := New(ts.URL, "token")
	ctx := context.Background()
	login, err := client.CurrentUser(ctx)
	if err != nil {
		t.Fatalf("whoami: %v", err)
	}
	if login != "octocat" {
		t.Fatalf("expected login octocat, got %s", login)
	}
}
