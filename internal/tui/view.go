package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("6"))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	selectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("2"))
	warningStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
)

func (m Model) View() tea.View {
	var out strings.Builder
	out.WriteString(titleStyle.Render("gitvac"))
	out.WriteString("  ")
	out.WriteString(mutedStyle.Render(truncate(m.root, max(12, m.width-12))))
	out.WriteString("\n")

	if m.confirming {
		fmt.Fprintf(&out, "\nDelete %d selected directories?\n", m.selectedCount())
		fmt.Fprintf(&out, "This cannot be undone. Selected size: %s\n\n", formatSize(m.selectedSize()))
		out.WriteString("y confirm   n / esc cancel")
		return makeView(out.String())
	}

	if m.scanning {
		out.WriteString("\nScanning directories and calculating sizes…\n")
		if m.status != "" {
			out.WriteString(mutedStyle.Render(m.status))
		}
		return makeView(out.String())
	}
	if m.deleting {
		out.WriteString("\nDeleting selected directories…\n")
		return makeView(out.String())
	}

	selectedSize := formatSize(m.selectedSize())
	reclaimable := formatSize(m.reclaimableSize())
	fmt.Fprintf(&out, "\n%d candidates   %d selected · %s   reclaimable · %s\n\n", len(m.candidates), m.selectedCount(), selectedSize, reclaimable)
	if len(m.candidates) == 0 {
		out.WriteString(mutedStyle.Render("No matching directories found."))
		out.WriteString("\n")
	} else {
		pathWidth := max(10, m.width-34)
		fmt.Fprintf(&out, "  %-7s %-15s %s\n", "SIZE", "TYPE", "PATH")
		start := min(m.offset, len(m.candidates))
		end := min(start+max(1, m.visibleRows()), len(m.candidates))
		for i := start; i < end; i++ {
			candidate := m.candidates[i]
			cursor, check := " ", " "
			if i == m.cursor {
				cursor = ">"
			}
			if m.selected[candidate.Path] {
				check = "✓"
			}
			row := fmt.Sprintf("%s %s %-7s %-15s %s", cursor, check, formatSize(candidate.Size), candidate.Name, truncate(candidate.Rel, pathWidth))
			if i == m.cursor {
				out.WriteString(selectedStyle.Render(row))
			} else {
				out.WriteString(row)
			}
			out.WriteByte('\n')
		}
		if end < len(m.candidates) {
			fmt.Fprintf(&out, "  %d more below\n", len(m.candidates)-end)
		}
	}
	if m.status != "" {
		out.WriteString("\n")
		out.WriteString(mutedStyle.Render(m.status))
	}
	if len(m.warnings) > 0 {
		out.WriteString("\n")
		out.WriteString(warningStyle.Render(truncate(m.warnings[0], max(20, m.width-2))))
		if len(m.warnings) > 1 {
			fmt.Fprintf(&out, " (+%d more warning(s))", len(m.warnings)-1)
		}
	}
	out.WriteString("\n\n")
	out.WriteString(mutedStyle.Render("j/k or ↑/↓ move   space select   a all   n none   d delete   q quit"))
	return makeView(out.String())
}

func makeView(content string) tea.View {
	view := tea.NewView(content)
	view.AltScreen = true
	return view
}

func formatSize(bytes int64) string {
	if bytes < 1024 {
		return fmt.Sprintf("%d B", bytes)
	}
	units := []string{"KiB", "MiB", "GiB", "TiB", "PiB"}
	value := float64(bytes)
	for _, unit := range units {
		value /= 1024
		if value < 1024 || unit == units[len(units)-1] {
			return fmt.Sprintf("%.1f %s", value, unit)
		}
	}
	return fmt.Sprintf("%d B", bytes)
}

func truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return "…"
	}
	return "…" + string(runes[len(runes)-width+1:])
}
