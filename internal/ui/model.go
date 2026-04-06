package ui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/evix1101/ssh-menu/internal/config"
	"github.com/evix1101/ssh-menu/internal/host"
	"github.com/evix1101/ssh-menu/internal/theme"
)

const minWidthForTwoPane = 60

// Model is the top-level Bubble Tea model.
type Model struct {
	hosts         []host.Host
	Selected      *host.Host
	PinToggled    bool
	verbose       bool
	sshOpts       string
	cursor        int
	scrollOffset  int
	viewIndex     int
	groups        []string
	filteredHosts []host.Host
	filterText    string
	statusMsg     string
	width         int
	height        int
}

// New creates a new UI model.
func New(hosts []host.Host, verbose bool, sshOpts string) *Model {
	m := &Model{
		hosts:   hosts,
		verbose: verbose,
		sshOpts: sshOpts,
		groups:  host.GetAllGroups(hosts),
	}
	m.updateFilteredHosts()
	return m
}

// Run starts the Bubble Tea program.
func Run(m *Model) error {
	p := tea.NewProgram(m, tea.WithAltScreen())
	finalModel, err := p.Run()
	if err != nil {
		return err
	}
	if fm, ok := finalModel.(*Model); ok {
		m.Selected = fm.Selected
		m.PinToggled = fm.PinToggled
	}
	return nil
}

// Init implements tea.Model.
func (m *Model) Init() tea.Cmd {
	return tea.WindowSize()
}

// Update implements tea.Model.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcScroll()
		return m, nil
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m *Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	action := classifyKey(msg)

	// When filter is active, 'p' is just a character, not pin toggle
	if action == keyTogglePin && m.filterText != "" {
		action = keyRune
	}

	// Clear status message on any key press
	m.statusMsg = ""

	switch action {
	case keyQuit:
		return m, tea.Quit
	case keySelect:
		return m.selectHost()
	case keyUp:
		m.moveCursor(-1)
	case keyDown:
		m.moveCursor(1)
	case keyLeft:
		m.navigateView(-1)
	case keyRight:
		m.navigateView(1)
	case keyTab:
		m.navigateView(1)
	case keyBackspace:
		m.handleBackspace()
	case keyTogglePin:
		m.togglePin()
	case keyRune:
		m.filterText += msg.String()
		m.updateFilteredHosts()
		m.cursor = 0
		m.scrollOffset = 0
	case keyNoop:
		// Ignore unhandled keys
	}
	return m, nil
}

func (m *Model) selectHost() (tea.Model, tea.Cmd) {
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

func (m *Model) moveCursor(delta int) {
	m.cursor += delta
	if m.cursor < 0 {
		m.cursor = 0
	}
	if m.cursor >= len(m.filteredHosts) {
		m.cursor = len(m.filteredHosts) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
	m.recalcScroll()
}

func (m *Model) navigateView(delta int) {
	totalViews := 1 + len(m.groups)
	m.viewIndex += delta
	if m.viewIndex < 0 {
		m.viewIndex = totalViews - 1
	} else if m.viewIndex >= totalViews {
		m.viewIndex = 0
	}
	m.filterText = ""
	m.cursor = 0
	m.scrollOffset = 0
	m.updateFilteredHosts()
}

func (m *Model) handleBackspace() {
	if len(m.filterText) > 0 {
		m.filterText = m.filterText[:len(m.filterText)-1]
		m.updateFilteredHosts()
		m.cursor = 0
		m.scrollOffset = 0
	}
}

func (m *Model) togglePin() {
	if len(m.filteredHosts) == 0 || m.cursor >= len(m.filteredHosts) {
		return
	}
	selected := &m.filteredHosts[m.cursor]
	newPinState := !selected.Pinned

	if selected.SourceFile != "" {
		if err := config.TogglePin(selected.SourceFile, selected.ShortName, newPinState); err != nil {
			m.statusMsg = fmt.Sprintf("Pin failed: %v", err)
			return
		}
	}

	for i := range m.hosts {
		if m.hosts[i].ShortName == selected.ShortName && m.hosts[i].SourceFile == selected.SourceFile {
			m.hosts[i].Pinned = newPinState
		}
	}

	m.PinToggled = true
	m.updateFilteredHosts()
}

func (m *Model) updateFilteredHosts() {
	var viewHosts []host.Host
	if m.viewIndex == 0 {
		viewHosts = m.hosts
	} else {
		groupIndex := m.viewIndex - 1
		if groupIndex < len(m.groups) {
			viewHosts = host.HostsForGroup(m.hosts, m.groups[groupIndex])
		}
	}

	filtered := host.FilterHosts(m.filterText, viewHosts)
	m.filteredHosts = host.SortWithPins(filtered)
}

func (m *Model) contentHeight() int {
	headerLines := 3
	if m.filterText != "" {
		headerLines += 2
	}
	if m.statusMsg != "" {
		headerLines++
	}
	h := m.height - headerLines
	if h < 1 {
		h = 20
	}
	return h
}

func (m *Model) recalcScroll() {
	m.scrollOffset = calculateScrollOffset(m.cursor, m.scrollOffset, m.contentHeight(), len(m.filteredHosts))
}

// View implements tea.Model.
func (m *Model) View() string {
	colors := theme.Current()
	var s strings.Builder

	helpText := "↑/↓ Navigate • ←/→ View • p Pin • Enter Select • Esc Quit"
	helpWidth := lipgloss.Width(helpText)
	title := "SSH Menu"
	titleWidth := lipgloss.Width(title)

	spacing := ""
	if m.width > 0 && m.width > titleWidth+helpWidth+2 {
		spacing = strings.Repeat(" ", m.width-titleWidth-helpWidth)
	}

	s.WriteString(theme.TitleStyle().Render(title))
	s.WriteString(spacing)
	s.WriteString(theme.DimStyle().Render(helpText))
	s.WriteString("\n")

	if len(m.groups) > 0 {
		s.WriteString(renderViewBar(m.groups, m.viewIndex))
		s.WriteString("\n")
	}
	s.WriteString("\n")

	if m.filterText != "" {
		filterStyle := lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color(colors.Accent))
		s.WriteString(filterStyle.Render(fmt.Sprintf("Filter: %s", m.filterText)))
		s.WriteString("\n\n")
	}

	if m.statusMsg != "" {
		s.WriteString(theme.WarningStyle().Render(m.statusMsg))
		s.WriteString("\n")
	}

	ch := m.contentHeight()

	if m.width >= minWidthForTwoPane {
		leftWidth := m.width * 55 / 100
		rightWidth := m.width - leftWidth - 1

		leftPane := renderHostList(m.filteredHosts, m.cursor, m.scrollOffset, leftWidth, ch)

		rightPane := ""
		if len(m.filteredHosts) > 0 && m.cursor < len(m.filteredHosts) {
			rightPane = renderDetail(m.filteredHosts[m.cursor], rightWidth, ch)
		}

		separator := lipgloss.NewStyle().
			Foreground(lipgloss.Color(colors.Border)).
			Render("│")

		leftLines := strings.Split(leftPane, "\n")
		rightLines := strings.Split(rightPane, "\n")

		maxLines := ch
		for len(leftLines) < maxLines {
			leftLines = append(leftLines, "")
		}
		for len(rightLines) < maxLines {
			rightLines = append(rightLines, "")
		}

		leftStyle := lipgloss.NewStyle().Width(leftWidth)
		for i := 0; i < maxLines; i++ {
			left := leftStyle.Render(safeGet(leftLines, i))
			right := safeGet(rightLines, i)
			s.WriteString(left)
			s.WriteString(separator)
			s.WriteString(right)
			s.WriteString("\n")
		}
	} else {
		s.WriteString(renderHostList(m.filteredHosts, m.cursor, m.scrollOffset, m.width, ch))
	}

	return s.String()
}

func safeGet(lines []string, i int) string {
	if i < len(lines) {
		return lines[i]
	}
	return ""
}
