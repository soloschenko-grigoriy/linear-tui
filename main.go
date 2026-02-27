package main

import (
	"fmt"
	"linear-tui/client"
	"linear-tui/components"
	"linear-tui/styles"
	"os"
	"os/exec"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/glamour"
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
	previewOffset      int
}

func initialModel() model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return model{
		offset:        0,
		previewOffset: 0,
		headerHeight:  3,
		spinner:       s,
		loading:       true,
		includeDone:   false,
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

				m.previewOffset = 0

				if m.cursor < m.offset {
					m.offset--
				}
			}
		case "down", "j":
			if m.cursor < len(m.issues)-1 {
				m.cursor++
				m.previewOffset = 0
				if m.cursor >= m.visibleIssuesCount+m.offset {
					m.offset++
				}
			}
		case "ctrl+d":
			m.previewOffset++
		case "ctrl+u":
			if m.previewOffset > 0 {
				m.previewOffset--
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

		stateStr := components.RenderState(issue.State)
		s += fmt.Sprintf("%s: %s\n", stateStr, title)
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

	title := lipgloss.NewStyle().
		Foreground(lipgloss.Color(styles.TitleColor)).
		Bold(true).
		Width(width).
		Render(fmt.Sprintf("%s\n", issue.Title))

	state := components.RenderState(issue.State)
	priorty := components.RenderPriorityStyle(issue.Priority)

	subtitle := lipgloss.NewStyle().
		Width(width).
		BorderBottom(true).
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color(styles.DividerColor)).
		Render(fmt.Sprintf("%s %s\n", state, priorty))

	renderer, _ := glamour.NewTermRenderer(
		glamour.WithAutoStyle(),
		glamour.WithWordWrap(width),
	)

	description, _ := renderer.Render(issue.Description)

	fullContent := lipgloss.JoinVertical(lipgloss.Left, title, subtitle, description)

	lines := strings.Split(fullContent, "\n")

	visibleLines := m.height - m.headerHeight

	if m.previewOffset > len(lines)-visibleLines {
		m.previewOffset = len(lines) - visibleLines
	}
	if m.previewOffset < 0 {
		m.previewOffset = 0
	}

	end := m.previewOffset + visibleLines
	if end > len(lines) {
		end = len(lines)
	}

	visible := strings.Join(lines[m.previewOffset:end], "\n")
	style := lipgloss.NewStyle().Width(width).Height(visibleLines)

	return style.Render(visible)
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
