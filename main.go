package main

import (
	"fmt"
	"linear-tui/client"
	"os"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type model struct {
	spinner  spinner.Model
	issues   []client.Issue
	loading  bool
	errorMsg errorMsg
	cursor   int
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

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil

}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("Loading issues...\n%s", m.spinner.View())
	}

	if m.errorMsg != (errorMsg{}) {
		return fmt.Sprintf("Error: %s\n", m.errorMsg.message)
	}

	s := "You have issues! \n"
	for i, issue := range m.issues {
		if i == m.cursor {
			s += "> "
		} else {
			s += "  "
		}

		s += fmt.Sprintf("[%s]: %s\n", issue.State.Name, issue.Title)
	}

	s += "\nPress q to quit\n"

	return s
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
