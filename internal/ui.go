package internal

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the UI state
type Model struct {
	hosts         []Host
	Selected      *Host
	verbose       bool
	detailed      bool
	sshOpts       string
	cursor        int
	viewIndex     int // Current view index (0 = flat, 1+ = groups)
	groups        []string
	filteredHosts []Host
	filterText    string
	width         int
	height        int
}

// InitStyles initializes the UI styling
func InitStyles(configPath string) {
	// Apply color configuration from theme
	ApplyColorConfig(configPath)
}

// SetupUI creates a new UI model
func SetupUI(hosts []Host, verbose bool, detailed bool, sshOpts string) *Model {
	m := &Model{
		hosts:         hosts,
		verbose:       verbose,
		detailed:      detailed,
		sshOpts:       sshOpts,
		cursor:        0,
		viewIndex:     0,
		filteredHosts: hosts,
	}

	// Get all groups
	m.groups = GetAllGroups(hosts)

	// Update filtered hosts for current view
	m.updateFilteredHosts()

	return m
}

// RunUI runs the interactive menu
func RunUI(m *Model) error {
	// Clear screen
	clearScreen()

	// Create and run the Bubble Tea program
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}

	// Check if a host was selected
	if fm, ok := finalModel.(*Model); ok {
		m.Selected = fm.Selected
	}

	return nil
}

// Init initializes the model
func (m *Model) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update handles messages and updates the model
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		return m.handleWindowSize(msg), nil
	case tea.KeyMsg:
		return m.handleKeyPress(msg)
	}
	return m, nil
}

// handleWindowSize handles window resize messages
func (m *Model) handleWindowSize(msg tea.WindowSizeMsg) tea.Model {
	m.width = msg.Width
	m.height = msg.Height
	return m
}

// handleKeyPress handles keyboard input
func (m *Model) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEscape:
		return m, tea.Quit
	case tea.KeyEnter:
		return m.handleEnterKey()
	case tea.KeyUp:
		m.moveCursor(-1)
	case tea.KeyDown:
		m.moveCursor(1)
	case tea.KeyLeft:
		m.navigateView(-1)
	case tea.KeyRight:
		m.navigateView(1)
	case tea.KeyBackspace:
		m.handleBackspace()
	case tea.KeyRunes:
		m.handleTypedCharacter(msg.String())
	case tea.KeyTab:
		m.navigateView(1)
	}
	return m, nil
}

// handleEnterKey handles the enter key press
func (m *Model) handleEnterKey() (tea.Model, tea.Cmd) {
	if len(m.filteredHosts) == 1 {
		m.Selected = &m.filteredHosts[0]
		return m, tea.Quit
	}
	if len(m.filteredHosts) > 0 && m.cursor < len(m.filteredHosts) {
		m.Selected = &m.filteredHosts[m.cursor]
		return m, tea.Quit
	}
	return m, nil
}

// handleBackspace handles the backspace key
func (m *Model) handleBackspace() {
	if len(m.filterText) > 0 {
		m.filterText = m.filterText[:len(m.filterText)-1]
		m.updateFilteredHosts()
		m.cursor = 0
	}
}

// handleTypedCharacter handles typed characters for filtering
func (m *Model) handleTypedCharacter(char string) {
	m.filterText += char
	m.updateFilteredHosts()
	m.cursor = 0
}

// View renders the UI
func (m *Model) View() string {
	colors := GetCurrentColors()

	var s strings.Builder

	// Calculate help text width for positioning
	helpText := "↑/↓ Navigate • ←/→ Switch View • Type to Filter • Enter Select • Esc Quit"
	helpWidth := lipgloss.Width(helpText)

	// Title and help on the same line
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color(colors.Accent))

	title := "SSH Menu"
	titleWidth := lipgloss.Width(title)

	// Calculate spacing for right-aligned help
	spacing := ""
	if m.width > 0 && m.width > titleWidth+helpWidth+2 {
		spacing = strings.Repeat(" ", m.width-titleWidth-helpWidth)
	}

	// Help style
	helpStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Dimmed))

	// Render title and help on same line
	s.WriteString(titleStyle.Render(title))
	s.WriteString(spacing)
	s.WriteString(helpStyle.Render(helpText))
	s.WriteString("\n")

	// View selector
	if len(m.groups) > 0 {
		s.WriteString(m.renderViewSelector())
		s.WriteString("\n")
	}
	s.WriteString("\n")

	// Filter indicator
	if m.filterText != "" {
		filterStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Accent)).
			Bold(true)
		s.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s", m.filterText)))
		s.WriteString("\n\n")
	}

	// Host list
	if len(m.filteredHosts) == 0 {
		dimStyle := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Dimmed))
		s.WriteString(dimStyle.Render("No hosts match your filter"))
	} else {
		for i, host := range m.filteredHosts {
			cursor := " "
			if m.cursor == i {
				cursor = "▸"
			}

			hostLine := fmt.Sprintf("%s %2d) %-20s %s@%s:%s",
				cursor, host.MenuNumber, host.ShortName, host.User, host.LongName, host.Port)

			if m.detailed {
				hostLine += fmt.Sprintf(" - %s", m.getHostDescription(host))
			}

			if m.cursor == i {
				selectedStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(colors.Selected)).
					Bold(true)
				s.WriteString(selectedStyle.Render(hostLine))
			} else {
				normalStyle := lipgloss.NewStyle().
					Foreground(lipgloss.Color(colors.Foreground))
				s.WriteString(normalStyle.Render(hostLine))
			}
			s.WriteString("\n")
		}
	}

	return s.String()
}

// renderViewSelector renders the colored view selector
func (m *Model) renderViewSelector() string {
	colors := GetCurrentColors()
	totalViews := 1 + len(m.groups)
	selectors := make([]string, totalViews)

	// Styles for selected and unselected views
	selectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Background)).
		Background(lipgloss.Color(colors.Selected)).
		Bold(true).
		Padding(0, 1)

	unselectedStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Foreground)).
		Padding(0, 1)

	// Add "All" view
	if m.viewIndex == 0 {
		selectors[0] = selectedStyle.Render("All")
	} else {
		selectors[0] = unselectedStyle.Render("All")
	}

	// Add group views
	for i, group := range m.groups {
		displayName := group
		if len(group) > 12 {
			displayName = group[:12] + "…"
		}

		if m.viewIndex == i+1 {
			selectors[i+1] = selectedStyle.Render(displayName)
		} else {
			selectors[i+1] = unselectedStyle.Render(displayName)
		}
	}

	// Join with a subtle separator
	separatorStyle := lipgloss.NewStyle().
		Foreground(lipgloss.Color(colors.Dimmed))

	return strings.Join(selectors, separatorStyle.Render(" • "))
}

// moveCursor moves the cursor up or down
func (m *Model) moveCursor(delta int) {
	m.cursor += delta

	if m.cursor < 0 {
		m.cursor = 0
	} else if m.cursor >= len(m.filteredHosts) {
		m.cursor = len(m.filteredHosts) - 1
	}
}

// navigateView cycles through views using left/right arrows
func (m *Model) navigateView(delta int) {
	// Total views = 1 (flat) + number of groups
	totalViews := 1 + len(m.groups)

	m.viewIndex += delta

	// Wrap around
	if m.viewIndex < 0 {
		m.viewIndex = totalViews - 1
	} else if m.viewIndex >= totalViews {
		m.viewIndex = 0
	}

	// Reset filter and cursor when changing views
	m.filterText = ""
	m.cursor = 0
	m.updateFilteredHosts()
}

// updateFilteredHosts updates the filtered host list based on current view and filter
func (m *Model) updateFilteredHosts() {
	// Get hosts for current view
	var viewHosts []Host

	if m.viewIndex == 0 {
		// Flat view - all hosts
		viewHosts = m.hosts
	} else {
		// Group view
		groupIndex := m.viewIndex - 1
		if groupIndex < len(m.groups) {
			groupName := m.groups[groupIndex]
			viewHosts = m.getHostsForGroup(groupName)
		}
	}

	// Apply filter
	if m.filterText == "" {
		m.filteredHosts = viewHosts
	} else {
		m.filteredHosts = []Host{}
		filterLower := strings.ToLower(m.filterText)

		for _, host := range viewHosts {
			// Check if filter matches menu number (as string prefix)
			menuNumStr := fmt.Sprintf("%d", host.MenuNumber)
			if strings.HasPrefix(menuNumStr, m.filterText) {
				m.filteredHosts = append(m.filteredHosts, host)
				continue
			}

			// Check if filter matches hostname (case insensitive)
			if strings.HasPrefix(strings.ToLower(host.ShortName), filterLower) {
				m.filteredHosts = append(m.filteredHosts, host)
				continue
			}

			// Also check long name
			if strings.HasPrefix(strings.ToLower(host.LongName), filterLower) {
				m.filteredHosts = append(m.filteredHosts, host)
			}
		}
	}
}

// getHostsForGroup returns hosts for a specific group
func (m *Model) getHostsForGroup(groupName string) []Host {
	var hosts []Host

	if groupName == "Ungrouped" {
		for _, h := range m.hosts {
			if len(h.Groups) == 0 {
				hosts = append(hosts, h)
			}
		}
	} else {
		for _, h := range m.hosts {
			for _, g := range h.Groups {
				if g == groupName {
					hosts = append(hosts, h)
					break
				}
			}
		}
	}

	return hosts
}

// getHostDescription returns a formatted description for a host
func (m *Model) getHostDescription(h Host) string {
	desc := h.DescText
	if h.IP != "" {
		desc += fmt.Sprintf(" (%s)", h.IP)
	}
	// Don't show groups in group view
	if m.viewIndex == 0 && len(h.Groups) > 0 {
		desc += fmt.Sprintf(" [%s]", strings.Join(h.Groups, ", "))
	}
	return desc
}

// ShowHelp displays keyboard shortcuts (deprecated - help is now inline)
func ShowHelp() {
	// This function is kept for compatibility but is no longer used
}

// clearScreen clears the terminal screen
func clearScreen() {
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd", "/c", "cls")
	} else {
		cmd = exec.Command("clear")
	}
	cmd.Stdout = os.Stdout
	cmd.Run()
}
