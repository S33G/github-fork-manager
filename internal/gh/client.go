package gh

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Repo represents a GitHub repository.
type Repo struct {
	ID            int64
	Name          string
	FullName      string
	Owner         string
	Private       bool
	Archived      bool
	Fork          bool
	Size          int
	Language      string
	DefaultBranch string
	Parent        string
	PushedAt      time.Time
	HTMLURL       string
	SSHURL        string
}

// Client is a minimal GitHub client.
type Client struct {
	BaseURL    string
	Token      string
	HTTPClient *http.Client
}

// New returns a Client with defaults applied.
func New(baseURL, token string) Client {
	return Client{
		BaseURL: strings.TrimSuffix(baseURL, "/"),
		Token:   token,
		HTTPClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// FetchForks retrieves all forks owned by the authenticated user.
func (c Client) FetchForks(ctx context.Context) ([]Repo, error) {
	if c.Token == "" {
		return nil, errors.New("GITHUB_TOKEN not set")
	}

	var repos []Repo
	page := 1

	for {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/user/repos?per_page=100&page=%d&affiliation=owner", c.BaseURL, page), nil)
		if err != nil {
			return nil, err
		}
		c.applyHeaders(req)

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			return nil, err
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("list repos: %s: %s", resp.Status, strings.TrimSpace(string(body)))
		}

		var payload []apiRepo
		if err := json.Unmarshal(body, &payload); err != nil {
			return nil, err
		}

		if len(payload) == 0 {
			break
		}

		for _, r := range payload {
			if !r.Fork {
				continue
			}
			repos = append(repos, mapRepo(r))
		}

		page++
	}

	return repos, nil
}

// DeleteRepo deletes a repository by full name.
func (c Client) DeleteRepo(ctx context.Context, fullName string) error {
	if c.Token == "" {
		return errors.New("GITHUB_TOKEN not set")
	}
	url := fmt.Sprintf("%s/repos/%s", c.BaseURL, fullName)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return err
	}
	c.applyHeaders(req)

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return err
	}
	body, err := io.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return err
	}

	if resp.StatusCode == http.StatusNoContent {
		return nil
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found: %s", fullName)
	}
	if resp.StatusCode == http.StatusForbidden {
		return fmt.Errorf("forbidden: %s", strings.TrimSpace(string(body)))
	}

	return fmt.Errorf("delete %s: %s: %s", fullName, resp.Status, strings.TrimSpace(string(body)))
}

func (c Client) applyHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	if c.Token != "" {
		req.Header.Set("Authorization", "Bearer "+c.Token)
	}
	req.Header.Set("User-Agent", "github-fork-manager")
}

type apiRepo struct {
	ID            int64     `json:"id"`
	Name          string    `json:"name"`
	FullName      string    `json:"full_name"`
	Private       bool      `json:"private"`
	Archived      bool      `json:"archived"`
	Fork          bool      `json:"fork"`
	Size          int       `json:"size"`
	Language      string    `json:"language"`
	DefaultBranch string    `json:"default_branch"`
	PushedAt      time.Time `json:"pushed_at"`
	Owner         struct {
		Login string `json:"login"`
	} `json:"owner"`
	Parent *struct {
		FullName string `json:"full_name"`
	} `json:"parent"`
	HTMLURL string `json:"html_url"`
	SSHURL  string `json:"ssh_url"`
}

func mapRepo(r apiRepo) Repo {
	parent := ""
	if r.Parent != nil {
		parent = r.Parent.FullName
	}

	return Repo{
		ID:            r.ID,
		Name:          r.Name,
		FullName:      r.FullName,
		Owner:         r.Owner.Login,
		Private:       r.Private,
		Archived:      r.Archived,
		Fork:          r.Fork,
		Size:          r.Size,
		Language:      r.Language,
		DefaultBranch: r.DefaultBranch,
		Parent:        parent,
		PushedAt:      r.PushedAt,
		HTMLURL:       r.HTMLURL,
		SSHURL:        r.SSHURL,
	}
}
