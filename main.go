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
	spinner            spinner.Model
	issues             []client.Issue
	loading            bool
	errorMsg           errorMsg
	cursor             int
	width              int
	height             int
	visibleIssuesCount int
	offset             int
	headerHeight       int
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		headerHeight: 3,
		spinner:      s,
		loading:      true,
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

				if m.cursor < m.offset {
					m.offset--
				}

			}
		case "down", "j":
			if m.cursor < len(m.issues)-1 {
				m.cursor++

				if m.cursor >= m.visibleIssuesCount+m.offset {
					m.offset++
				}
			}
		}
	case issuesLoadedMsg:
		m.issues = msg.issues
		m.errorMsg = msg.errorMsg
		m.loading = false
		m.visibleIssuesCount = m.height - m.headerHeight

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.visibleIssuesCount = m.height - m.headerHeight
	}

	var cmd1 tea.Cmd
	m.spinner, cmd1 = m.spinner.Update(msg)

	return m, tea.Batch(cmd1)
}

func RenderList(m model) string {
	issues := m.issues
	cursor := m.cursor
	width := int(float64(m.width) * 0.6)

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Width(width)
	s := "You have issues! \n"

	end := m.offset + m.visibleIssuesCount
	if end > len(issues) {
		end = len(issues)
	}
	for i, issue := range (issues)[m.offset:end] {
		if i+m.offset == cursor {
			s += "> "
		} else {
			s += "  "
		}

		maxLength := width - 20
		title := issue.Title
		if len(title) > maxLength {
			title = title[:maxLength] + "..."
		}
		s += fmt.Sprintf("[%s]: %s\n", issue.State.Name, title)
	}
	// s += fmt.Sprintf("\nh=%d visible=%d offset=%d", m.height, m.visibleIssuesCount, m.offset)

	return style.Render(s)
}

func RenderPreview(m model) string {
	issue := m.issues[m.cursor]
	width := int(float64(m.width) * 0.4)
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("216")).Width(width)

	return style.Render(fmt.Sprintf("Issue: %s\n", issue.Title))
}

func RenderFooter(m model) string {
	style := lipgloss.NewStyle().Foreground(lipgloss.Color("116"))

	return style.Render("Press q to quit")
}

func (m model) View() string {
	if m.loading {
		return fmt.Sprintf("Loading issues...\n%s", m.spinner.View())
	}

	if m.errorMsg != (errorMsg{}) {
		return fmt.Sprintf("Error: %s\n", m.errorMsg.message)
	}

	// m.viewport.SetContent()
	list := RenderList(m)
	preview := RenderPreview(m)
	footer := RenderFooter(m)
	content := lipgloss.JoinHorizontal(lipgloss.Top, list, preview)

	return lipgloss.JoinVertical(lipgloss.Left, content, footer)
}

func main() {
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
