package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Model represents the UI model
type Model struct {
	list     list.Model
	width    int
	height   int
	ready    bool
	quitting bool
}

// NewModel creates a new UI model
func NewModel() Model {
	return Model{
		list:     list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0),
		width:    80,
		height:   24,
		ready:    false,
		quitting: false,
	}
}

// Init initializes the UI model
func (m Model) Init() tea.Cmd {
	return nil
}

// Update updates the UI model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			m.quitting = true
			return m, tea.Quit
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

// View renders the UI model
func (m Model) View() string {
	if m.quitting {
		return "Goodbye!\n"
	}

	if !m.ready {
		return "Initializing..."
	}

	style := lipgloss.NewStyle().
		Foreground(lipgloss.Color("#FAFAFA")).
		Background(lipgloss.Color("#7D56F4")).
		Padding(1, 2)

	header := style.Render("OSS Manager")

	return fmt.Sprintf("%s\n\n%s", header, m.list.View())
}

// Run starts the UI
func Run() error {
	p := tea.NewProgram(NewModel(), tea.WithAltScreen())
	_, err := p.Run()
	return err
}
