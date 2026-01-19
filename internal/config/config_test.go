package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPrefersEnvOverridesAndExpandsPaths(t *testing.T) {
	tmp := t.TempDir()
	homeCfgDir := filepath.Join(tmp, ".github-fork-manager")
	if err := os.MkdirAll(homeCfgDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	cfgPath := filepath.Join(homeCfgDir, "config.json")
	err := os.WriteFile(cfgPath, []byte(`{
		"token": "filetoken",
		"api_base": "https://example.com/api",
		"log_path": "~/.github-fork-manager/log.txt"
	}`), 0o644)
	if err != nil {
		t.Fatalf("write config: %v", err)
	}

	t.Setenv("HOME", tmp)
	t.Setenv("GITHUB_TOKEN", "envtoken")
	t.Setenv("GITHUB_API_BASE", "https://env.example.com")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Token != "envtoken" {
		t.Fatalf("expected env token override, got %q", cfg.Token)
	}
	if cfg.APIBase != "https://env.example.com" {
		t.Fatalf("expected env api base override, got %q", cfg.APIBase)
	}
	wantLog := filepath.Join(tmp, ".github-fork-manager", "log.txt")
	if cfg.LogPath != wantLog {
		t.Fatalf("expected expanded log path %q, got %q", wantLog, cfg.LogPath)
	}
}

func TestLoadDefaultsWhenNoFile(t *testing.T) {
	tmp := t.TempDir()
	t.Setenv("HOME", tmp)
	t.Setenv("GITHUB_TOKEN", "")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.APIBase == "" || cfg.APIBase != "https://api.github.com" {
		t.Fatalf("unexpected api base: %q", cfg.APIBase)
	}
	defLog := filepath.Join(tmp, ".github-fork-manager", "actions.log")
	if cfg.LogPath != defLog {
		t.Fatalf("expected default log path %q, got %q", defLog, cfg.LogPath)
	}
}
