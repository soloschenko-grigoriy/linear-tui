package main

import (
	"fmt"
	"linear-tui/client"
	"os"
	"os/exec"
	"slices"

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
	includeDone        bool
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		headerHeight: 3,
		spinner:      s,
		loading:      true,
		includeDone:  false,
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
	return tea.Batch(fetchIssuesCmd(m.includeDone), m.spinner.Tick)
}

func fetchIssuesCmd(includeDone bool) tea.Cmd {
	return func() tea.Msg {
		issues, err := client.FetchIssues(includeDone)

		if err != nil {
			return issuesLoadedMsg{nil, errorMsg{err.Error()}}
		}

		var stateOrder = map[string]int{
			"In Progress": 0,
			"In Review":   1,
			"Pending":     2,
			"Todo":        3,
			"Done":        4,
			"Canceled":    5,
		}
		slices.SortFunc(issues, func(a, b client.Issue) int {
			orderA := stateOrder[a.State.Name]
			orderB := stateOrder[b.State.Name]

			return orderA - orderB
		})

		return issuesLoadedMsg{issues, errorMsg{}}
	}
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
		case "e":
			m.cursor = 0
			m.loading = true
			m.offset = 0
			m.includeDone = !m.includeDone // Toogle include done
			return m, fetchIssuesCmd(m.includeDone)
		case "o":
			if m.cursor >= 0 || m.cursor < len(m.issues) {
				url := m.issues[m.cursor].URL
				return m, func() tea.Msg {
					exec.Command("open", url).Start()
					return nil
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

	var stateColors = map[string]string{
		"In Review":   "#89b4fa", // Blue — active work
		"Pending":     "#7f849c", // Mauve — waiting for review
		"In Progress": "#f9e2af", // Yellow — on hold
		"Todo":        "#a6adc8", // Subtext 0 — not started, muted
		"Done":        "#a6e3a1", // Green — completed
		"Canceled":    "#6c7086", // Overlay 0 — dimmed/inactive
	}

	style := lipgloss.NewStyle().Foreground(lipgloss.Color("205")).Width(width)
	if len(issues) == 0 {
		return style.Render("No issues found")
	}

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

		color := stateColors[issue.State.Name]
		stateStyle := lipgloss.NewStyle().Foreground(lipgloss.Color(color))

		stateStr := fmt.Sprintf("[%s]", issue.State.Name)
		s += fmt.Sprintf("%s: %s\n", stateStyle.Render(stateStr), title)
	}
	// s += fmt.Sprintf("\nh=%d visible=%d offset=%d", m.height, m.visibleIssuesCount, m.offset)

	return style.Render(s)
}

func RenderPreview(m model) string {
	if len(m.issues) == 0 {
		return ""
	}

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
