package main

import (
	"strings"
	"testing"
	"time"

	"github.com/seeg/github-fork-manager/internal/gh"
)

func TestApprovalPhrase(t *testing.T) {
	if got := approvalPhrase("alice"); got != "alice approves" {
		t.Fatalf("expected approval phrase to include login, got %q", got)
	}
	if got := approvalPhrase(""); got != "your-github-username approves" {
		t.Fatalf("expected fallback approval phrase, got %q", got)
	}
}

func TestHyperlinkWrapsText(t *testing.T) {
	out := hyperlink("https://example.com/repo", "repo")
	if !strings.Contains(out, "https://example.com/repo") || !strings.Contains(out, "repo") {
		t.Fatalf("hyperlink missing content: %q", out)
	}
}

func TestEnsureVisibleClampsOffsets(t *testing.T) {
	m := model{
		filtered:   []gh.Repo{{}, {}, {}, {}, {}},
		listHeight: 3,
		cursor:     4,
	}
	m.ensureVisible()
	if m.listOffset != 2 {
		t.Fatalf("expected listOffset 2, got %d", m.listOffset)
	}
	if m.cursor != 4 {
		t.Fatalf("cursor changed unexpectedly: %d", m.cursor)
	}
}

func TestApplyFilterMatchesNameAndLanguage(t *testing.T) {
	repos := []gh.Repo{
		{FullName: "alice/foo", Language: "Go", PushedAt: time.Now()},
		{FullName: "bob/bar", Language: "Python", PushedAt: time.Now()},
	}
	m := model{repos: repos}
	got := m.applyFilter("go")
	if len(got) != 1 || got[0].FullName != "alice/foo" {
		t.Fatalf("expected only Go repo, got %#v", got)
	}
}
