package components

import (
	"fmt"
	"linear-tui/client"

	"github.com/charmbracelet/lipgloss"
)

var StateColors = map[string]string{
	"In Review":   "#89b4fa", // Blue — active work
	"Pending":     "#7f849c", // Mauve — waiting for review
	"In Progress": "#f9e2af", // Yellow — on hold
	"Todo":        "#a6adc8", // Subtext 0 — not started, muted
	"Done":        "#a6e3a1", // Green — completed
	"Canceled":    "#6c7086", // Overlay 0 — dimmed/inactive
}

func RenderState(state client.State) string {
	return lipgloss.NewStyle().
		Foreground(lipgloss.Color(StateColors[state.Name])).
		Render(fmt.Sprintf("[%s]", state.Name))

}
