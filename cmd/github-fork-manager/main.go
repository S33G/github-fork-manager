package main

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/seeg/github-fork-manager/internal/config"
	"github.com/seeg/github-fork-manager/internal/gh"
)

var version = "dev"

type mode int

const (
	modeNormal mode = iota
	modeFiltering
)

type model struct {
	cfg           config.Config
	client        gh.Client
	repos         []gh.Repo
	filtered      []gh.Repo
	cursor        int
	listOffset    int
	listHeight    int
	selected      map[string]bool
	status        string
	err           error
	loading       bool
	deleting      bool
	deleteQueue   []gh.Repo
	deleteResults map[string]string
	filterInput   textinput.Model
	mode          mode
}

func newModel(cfg config.Config) model {
	ti := textinput.New()
	ti.Placeholder = "type to filter (owner/name, language); enter to apply, esc to clear"
	ti.CharLimit = 64
	ti.Prompt = "/ "

	return model{
		cfg:           cfg,
		client:        gh.New(cfg.APIBase, cfg.Token),
		selected:      make(map[string]bool),
		deleteResults: make(map[string]string),
		filterInput:   ti,
		loading:       true,
		status:        "Loading forks…",
		mode:          modeNormal,
		listHeight:    15,
	}
}

func (m model) Init() tea.Cmd {
	return loadReposCmd(m.client)
}

type reposLoadedMsg struct {
	repos []gh.Repo
	err   error
}

type deleteResultMsg struct {
	repo gh.Repo
	err  error
}

func loadReposCmd(client gh.Client) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		repos, err := client.FetchForks(ctx)
		return reposLoadedMsg{repos: repos, err: err}
	}
}

func deleteNextCmd(client gh.Client, repo gh.Repo) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		defer cancel()
		err := client.DeleteRepo(ctx, repo.FullName)
		return deleteResultMsg{repo: repo, err: err}
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.listHeight = msg.Height - 8 // leave room for header/footer lines
		if m.listHeight < 5 {
			m.listHeight = 5
		}
		m.ensureVisible()
		return m, nil
	case reposLoadedMsg:
		m.loading = false
		m.err = msg.err
		if msg.err == nil {
			m.repos = sortRepos(msg.repos)
			m.filtered = m.applyFilter(m.filterInput.Value())
			m.status = fmt.Sprintf("Loaded %d forks", len(m.repos))
			m.ensureVisible()
		} else {
			m.status = "Failed to load forks"
		}
		return m, nil
	case deleteResultMsg:
		m.deleting = len(m.deleteQueue) > 1
		m.deleteQueue = popQueue(m.deleteQueue)
		if msg.err != nil {
			m.deleteResults[msg.repo.FullName] = "error: " + msg.err.Error()
			m.status = fmt.Sprintf("Failed to delete %s", msg.repo.FullName)
		} else {
			m.deleteResults[msg.repo.FullName] = "deleted"
			m.status = fmt.Sprintf("Deleted %s", msg.repo.FullName)
			m.removeRepo(msg.repo.FullName)
			delete(m.selected, msg.repo.FullName)
		}
		logLine(m.cfg.LogPath, fmt.Sprintf("delete %s -> %s", msg.repo.FullName, m.deleteResults[msg.repo.FullName]))

		if len(m.deleteQueue) > 0 {
			return m, deleteNextCmd(m.client, m.deleteQueue[0])
		}
		m.deleting = false
		return m, nil
	}

	// Key handling.
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.mode == modeFiltering {
			var cmd tea.Cmd
			m.filterInput, cmd = m.filterInput.Update(msg)
			switch msg.Type {
			case tea.KeyEnter:
				m.mode = modeNormal
				m.cursor = 0
				m.filtered = m.applyFilter(m.filterInput.Value())
				m.status = fmt.Sprintf("Filter applied: %d shown", len(m.filtered))
				m.ensureVisible()
			case tea.KeyEsc:
				m.mode = modeNormal
				m.filterInput.SetValue("")
				m.filtered = m.applyFilter("")
				m.status = "Filter cleared"
				m.ensureVisible()
			}
			return m, cmd
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "j", "down":
			if m.cursor < len(m.filtered)-1 {
				m.cursor++
				m.ensureVisible()
			}
		case "k", "up":
			if m.cursor > 0 {
				m.cursor--
				m.ensureVisible()
			}
		case "r":
			m.loading = true
			m.status = "Refreshing…"
			return m, loadReposCmd(m.client)
		case "/":
			m.mode = modeFiltering
			m.filterInput.Focus()
			return m, nil
		case " ":
			if len(m.filtered) == 0 {
				return m, nil
			}
			m.toggleSelection(m.filtered[m.cursor].FullName)
		case "a":
			m.toggleSelectAll()
		case "d":
			if m.deleting {
				m.status = "Delete already in progress"
				return m, nil
			}
			queue := m.selectedRepos()
			if len(queue) == 0 {
				m.status = "Nothing selected"
				return m, nil
			}
			m.deleteQueue = queue
			m.deleting = true
			m.status = fmt.Sprintf("Deleting %d repos…", len(queue))
			if len(queue) > 0 {
				return m, deleteNextCmd(m.client, queue[0])
			}
		case "?":
			m.status = "Keys: j/k move · space select · a select all · / filter · d delete · r refresh · q quit"
		}
	}

	return m, nil
}

func (m *model) toggleSelection(fullName string) {
	if m.selected[fullName] {
		delete(m.selected, fullName)
		m.status = fmt.Sprintf("Unselected %s", fullName)
		return
	}
	m.selected[fullName] = true
	m.status = fmt.Sprintf("Selected %s", fullName)
}

func (m *model) toggleSelectAll() {
	if len(m.filtered) == 0 {
		return
	}
	visibleSelected := 0
	for _, repo := range m.filtered {
		if m.selected[repo.FullName] {
			visibleSelected++
		}
	}
	if visibleSelected == len(m.filtered) {
		for _, repo := range m.filtered {
			delete(m.selected, repo.FullName)
		}
		m.status = "Cleared visible selections"
		return
	}
	for _, repo := range m.filtered {
		m.selected[repo.FullName] = true
	}
	m.status = fmt.Sprintf("Selected %d visible repos", len(m.filtered))
}

func (m *model) ensureVisible() {
	if len(m.filtered) == 0 {
		m.cursor = 0
		m.listOffset = 0
		return
	}
	if m.cursor >= len(m.filtered) {
		m.cursor = len(m.filtered) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}

	height := m.listHeight
	if height <= 0 || height > len(m.filtered) {
		height = len(m.filtered)
	}

	maxOffset := len(m.filtered) - height
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.listOffset > maxOffset {
		m.listOffset = maxOffset
	}
	if m.cursor < m.listOffset {
		m.listOffset = m.cursor
	} else if m.cursor >= m.listOffset+height {
		m.listOffset = m.cursor - height + 1
	}
}

func (m model) selectedRepos() []gh.Repo {
	if len(m.selected) == 0 {
		return nil
	}
	lookup := make(map[string]gh.Repo, len(m.repos))
	for _, r := range m.repos {
		lookup[r.FullName] = r
	}
	var out []gh.Repo
	for name := range m.selected {
		if repo, ok := lookup[name]; ok {
			out = append(out, repo)
		}
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].FullName < out[j].FullName
	})
	return out
}

func (m *model) removeRepo(fullName string) {
	filtered := make([]gh.Repo, 0, len(m.repos))
	for _, r := range m.repos {
		if r.FullName == fullName {
			continue
		}
		filtered = append(filtered, r)
	}
	m.repos = filtered
	m.filtered = m.applyFilter(m.filterInput.Value())
	if m.cursor >= len(m.filtered) && m.cursor > 0 {
		m.cursor = len(m.filtered) - 1
	}
	m.ensureVisible()
}

func (m model) applyFilter(filter string) []gh.Repo {
	if filter == "" {
		return append([]gh.Repo{}, m.repos...)
	}
	filter = strings.ToLower(filter)
	var out []gh.Repo
	for _, repo := range m.repos {
		if strings.Contains(strings.ToLower(repo.FullName), filter) ||
			strings.Contains(strings.ToLower(repo.Language), filter) ||
			strings.Contains(strings.ToLower(repo.Owner), filter) {
			out = append(out, repo)
		}
	}
	return out
}

func (m model) View() string {
	var b strings.Builder

	title := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("213")).Render("GitHub Fork Manager")
	b.WriteString(title)
	b.WriteString("\n")

	if m.cfg.Token == "" {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("GITHUB_TOKEN not set. Export GITHUB_TOKEN or set token in ~/.github-fork-manager/config.json.\n\n"))
	}

	stats := fmt.Sprintf("Total: %d | Filtered: %d | Selected: %d", len(m.repos), len(m.filtered), len(m.selected))
	if m.deleting {
		stats += fmt.Sprintf(" | Deleting %d…", len(m.deleteQueue))
	}
	b.WriteString(stats + "\n")
	b.WriteString("Commands: j/k move · space select · a select all · / filter · d delete · r refresh · q quit\n")
	b.WriteString("Filter: ")
	if m.mode == modeFiltering {
		b.WriteString(m.filterInput.View())
	} else {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render(m.filterInput.View()))
	}
	b.WriteString("\n\n")

	if m.loading {
		b.WriteString("Loading forks…\n")
		return b.String()
	}

	if m.err != nil {
		b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("196")).Render("Error: "+m.err.Error()) + "\n")
	}

	if len(m.filtered) == 0 {
		b.WriteString("No forks found.\n")
	} else {
		listHeight := m.listHeight
		if listHeight <= 0 || listHeight > len(m.filtered) {
			listHeight = len(m.filtered)
		}
		end := m.listOffset + listHeight
		if end > len(m.filtered) {
			end = len(m.filtered)
		}
		for i := m.listOffset; i < end; i++ {
			repo := m.filtered[i]
			cursor := "  "
			if i == m.cursor {
				cursor = lipgloss.NewStyle().Foreground(lipgloss.Color("12")).Render("> ")
			}
			check := "[ ]"
			if m.selected[repo.FullName] {
				check = lipgloss.NewStyle().Foreground(lipgloss.Color("77")).Render("[x]")
			}
			meta := repoMeta(repo)
			name := repo.FullName
			if repo.HTMLURL != "" {
				name = hyperlink(repo.HTMLURL, repo.FullName)
			}
			line := fmt.Sprintf("%s%s %s — %s", cursor, check, name, meta)
			b.WriteString(line + "\n")
		}
		if len(m.filtered) > listHeight {
			b.WriteString(fmt.Sprintf("Showing %d-%d of %d\n", m.listOffset+1, end, len(m.filtered)))
		}
	}

	if m.status != "" {
		b.WriteString("\n" + m.status + "\n")
	}
	if len(m.deleteResults) > 0 {
		b.WriteString("\nRecent results:\n")
		count := 0
		for name, res := range m.deleteResults {
			b.WriteString(fmt.Sprintf("- %s: %s\n", name, res))
			count++
			if count >= 5 {
				break
			}
		}
	}

	return b.String()
}

func repoMeta(repo gh.Repo) string {
	var parts []string
	if repo.Language != "" {
		parts = append(parts, repo.Language)
	}
	if repo.Private {
		parts = append(parts, "private")
	}
	if repo.Archived {
		parts = append(parts, "archived")
	}
	if repo.Parent != "" {
		parts = append(parts, "parent: "+repo.Parent)
	}
	if repo.PushedAt.IsZero() {
		parts = append(parts, "pushed unknown")
	} else {
		parts = append(parts, "pushed "+repo.PushedAt.Format("2006-01-02"))
	}
	return strings.Join(parts, " · ")
}

func popQueue(queue []gh.Repo) []gh.Repo {
	if len(queue) == 0 {
		return queue
	}
	return queue[1:]
}

func sortRepos(repos []gh.Repo) []gh.Repo {
	out := append([]gh.Repo{}, repos...)
	sort.Slice(out, func(i, j int) bool {
		return out[i].PushedAt.After(out[j].PushedAt)
	})
	return out
}

func hyperlink(url, text string) string {
	if url == "" || text == "" {
		return text
	}
	esc := "\x1b"
	return fmt.Sprintf("%s]8;;%s%s\\%s%s]8;;%s\\", esc, url, esc, text, esc, esc)
}

func logLine(path, line string) {
	if path == "" {
		return
	}
	if err := config.EnsureLogDir(path); err != nil {
		return
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	defer f.Close()
	fmt.Fprintf(f, "%s %s\n", time.Now().Format(time.RFC3339), line)
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}
	if err := config.EnsureLogDir(cfg.LogPath); err != nil {
		fmt.Fprintf(os.Stderr, "log dir error: %v\n", err)
	}

	p := tea.NewProgram(newModel(cfg))
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
