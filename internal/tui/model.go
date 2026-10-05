package tui

import (
	"fmt"
	"path/filepath"
	"sort"

	"gitvac/internal/scanner"

	tea "charm.land/bubbletea/v2"
)

type scanFinished struct {
	candidates []scanner.Candidate
	errors     []error
}

type deleteFinished struct {
	candidates []scanner.Candidate
	errors     []error
	deleted    int
	failed     int
}

type Model struct {
	root       string
	targets    []string
	excludes   []string
	candidates []scanner.Candidate
	selected   map[string]bool
	cursor     int
	offset     int
	width      int
	height     int
	scanning   bool
	deleting   bool
	confirming bool
	status     string
	warnings   []string
}

func New(root string, targets, excludes []string) Model {
	return Model{
		root:     root,
		targets:  append([]string(nil), targets...),
		excludes: append([]string(nil), excludes...),
		selected: make(map[string]bool),
		width:    100,
		height:   24,
		scanning: true,
	}
}

func (m Model) Init() tea.Cmd {
	return m.scanCmd()
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.keepCursorVisible()
	case scanFinished:
		m.scanning = false
		m.candidates = msg.candidates
		m.warnings = errorStrings(msg.errors)
		m.pruneSelection()
		if len(msg.errors) > 0 {
			m.status = fmt.Sprintf("Scan finished with %d warning(s)", len(msg.errors))
		} else {
			m.status = "Scan complete"
		}
	case deleteFinished:
		m.deleting = false
		m.candidates = msg.candidates
		m.warnings = errorStrings(msg.errors)
		m.pruneSelection()
		m.status = fmt.Sprintf("Removed %d selected directories; %d failed", msg.deleted, msg.failed)
	case tea.KeyPressMsg:
		if m.scanning || m.deleting {
			if msg.String() == "ctrl+c" || msg.String() == "q" {
				return m, tea.Quit
			}
			return m, nil
		}
		if m.confirming {
			return m, m.updateConfirmation(msg)
		}
		return m, m.updateKeys(msg)
	}
	return m, nil
}

func (m Model) scanCmd() tea.Cmd {
	root := m.root
	targets := append([]string(nil), m.targets...)
	excludes := append([]string(nil), m.excludes...)
	return func() tea.Msg {
		candidates, errs := scanner.Scan(root, targets, excludes)
		return scanFinished{candidates: candidates, errors: errs}
	}
}

func (m *Model) updateKeys(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "ctrl+c", "q":
		return tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(m.candidates)-1 {
			m.cursor++
		}
	case "space":
		if len(m.candidates) > 0 {
			path := m.candidates[m.cursor].Path
			m.selected[path] = !m.selected[path]
			m.status = ""
		}
	case "a":
		for _, candidate := range m.candidates {
			m.selected[candidate.Path] = true
		}
		m.status = "All candidates selected"
	case "n":
		m.selected = make(map[string]bool)
		m.status = "Selection cleared"
	case "d":
		if m.selectedCount() > 0 {
			m.confirming = true
			m.status = ""
		}
	}
	m.keepCursorVisible()
	return nil
}

func (m *Model) updateConfirmation(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc", "n":
		m.confirming = false
		m.status = "Deletion cancelled"
	case "q", "ctrl+c":
		return tea.Quit
	case "y":
		m.confirming = false
		m.deleting = true
		m.status = "Deleting selected directories…"
		selected := m.selectedCandidates()
		root := m.root
		targets := append([]string(nil), m.targets...)
		excludes := append([]string(nil), m.excludes...)
		return func() tea.Msg {
			deleted, failed, errs := deleteCandidates(root, selected)
			candidates, scanErrors := scanner.Scan(root, targets, excludes)
			errs = append(errs, scanErrors...)
			return deleteFinished{candidates: candidates, errors: errs, deleted: deleted, failed: failed}
		}
	}
	return nil
}

func deleteCandidates(root string, candidates []scanner.Candidate) (int, int, []error) {
	selected := append([]scanner.Candidate(nil), candidates...)
	sort.Slice(selected, func(i, j int) bool {
		return len(selected[i].Path) < len(selected[j].Path)
	})
	deleted, failed := 0, 0
	var errs []error
	removed := make(map[string]bool, len(selected))
	for _, candidate := range selected {
		covered := false
		for parent := filepath.Dir(candidate.Path); parent != "." && parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
			if removed[parent] {
				covered = true
				break
			}
		}
		if covered {
			deleted++
			continue
		}
		if err := scanner.Delete(root, candidate.Path); err != nil {
			failed++
			errs = append(errs, fmt.Errorf("%s: %w", candidate.Rel, err))
			continue
		}
		deleted++
		removed[candidate.Path] = true
	}
	return deleted, failed, errs
}

func (m *Model) selectedCandidates() []scanner.Candidate {
	var selected []scanner.Candidate
	for _, candidate := range m.candidates {
		if m.selected[candidate.Path] {
			selected = append(selected, candidate)
		}
	}
	return selected
}

func (m *Model) selectedCount() int {
	count := 0
	for _, candidate := range m.candidates {
		if m.selected[candidate.Path] {
			count++
		}
	}
	return count
}

func (m *Model) pruneSelection() {
	found := make(map[string]bool, len(m.candidates))
	for _, candidate := range m.candidates {
		found[candidate.Path] = true
	}
	for path := range m.selected {
		if !found[path] {
			delete(m.selected, path)
		}
	}
	if m.cursor >= len(m.candidates) {
		m.cursor = len(m.candidates) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.keepCursorVisible()
}

func (m *Model) keepCursorVisible() {
	visible := m.visibleRows()
	if visible < 1 {
		visible = 1
	}
	if m.cursor < m.offset {
		m.offset = m.cursor
	}
	if m.cursor >= m.offset+visible {
		m.offset = m.cursor - visible + 1
	}
	if m.offset < 0 {
		m.offset = 0
	}
}

func (m Model) visibleRows() int {
	return m.height - 13
}

func (m Model) selectedSize() int64 {
	selected := make(map[string]bool)
	for _, candidate := range m.selectedCandidates() {
		selected[filepath.Clean(candidate.Path)] = true
	}
	return uniqueSize(m.selectedCandidates(), selected)
}

func (m Model) reclaimableSize() int64 {
	all := make(map[string]bool, len(m.candidates))
	for _, candidate := range m.candidates {
		all[filepath.Clean(candidate.Path)] = true
	}
	return uniqueSize(m.candidates, all)
}

func uniqueSize(candidates []scanner.Candidate, included map[string]bool) int64 {
	var total int64
	for _, candidate := range candidates {
		path := filepath.Clean(candidate.Path)
		covered := false
		for parent := filepath.Dir(path); parent != "." && parent != filepath.Dir(parent); parent = filepath.Dir(parent) {
			if included[filepath.Clean(parent)] {
				covered = true
				break
			}
		}
		if !covered {
			total += candidate.Size
		}
	}
	return total
}

func errorStrings(errs []error) []string {
	out := make([]string, 0, len(errs))
	for _, err := range errs {
		out = append(out, err.Error())
	}
	return out
}
