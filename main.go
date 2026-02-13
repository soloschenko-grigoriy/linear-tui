package main

import (
	"fmt"
	"linear-tui/client"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner  spinner.Model
	issues   []client.Issue
	loading  bool
	errorMsg errorMsg
	cursor   int
	width    int
	height   int
	ready    bool
	viewport viewport.Model
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		spinner: s,
		loading: true,
	}
}

type issuesLoadedMsg struct {
	issues   []client.Issue
	errorMsg errorMsg
}

type errorMsg struct {
	message string
}

func (m model) Init() tea.Cmd {
	return tea.Batch(fetchIssuesCmd, m.spinner.Tick)
}

func fetchIssuesCmd() tea.Msg {
	issues, err := client.FetchIssues()

	if err != nil {
		return issuesLoadedMsg{nil, errorMsg{err.Error()}}
	}

	return issuesLoadedMsg{issues, errorMsg{}}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.issues)-1 {
				m.cursor++
			}
		}
	case issuesLoadedMsg:
		m.issues = msg.issues
		m.errorMsg = msg.errorMsg
		m.loading = false

	case tea.WindowSizeMsg:
		if !m.ready {
			m.viewport = viewport.New(msg.Width, msg.Height)
			m.viewport.YPosition = 0
			m.ready = true
		} else {
			m.width = msg.Width
			m.height = msg.Height
		}

	}

	var cmd1 tea.Cmd
	m.spinner, cmd1 = m.spinner.Update(msg)

	var cmd2 tea.Cmd
	m.viewport, cmd2 = m.viewport.Update(msg)

	return m, tea.Batch(cmd1, cmd2)
}

func RenderList(issues []client.Issue, cursor int, viewport viewport.Model) string {
	width := int(float64(viewport.Width) * 0.6)

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Width(width)
	s := "You have issues! \n"
	for i, issue := range issues {
		if i == cursor {
			s += "> "
		} else {
			s += "  "
		}

		s += fmt.Sprintf("[%s]: %s\n", issue.State.Name, issue.Title)
	}

	s += "\nPress q to quit\n"

	return style.Render(s)
}

func RenderPreview(issue client.Issue, viewport viewport.Model) string {
	width := int(float64(viewport.Width) * 0.4)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("215")).Width(width)

	return style.Render(fmt.Sprintf("Issue: %s\n", issue.Title))
}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("Loading issues...\n%s", m.spinner.View())
	}

	if m.errorMsg != (errorMsg{}) {
		return fmt.Sprintf("Error: %s\n", m.errorMsg.message)
	}

	// m.viewport.SetContent()
	list := RenderList(m.issues, m.cursor, m.viewport)
	preview := RenderPreview(m.issues[m.cursor], m.viewport)
	content := lipgloss.JoinHorizontal(lipgloss.Top, list, preview)

	m.viewport.SetContent(content)
	return m.viewport.View()
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
