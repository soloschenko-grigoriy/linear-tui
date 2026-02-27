package components

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var PriorityColor = "#fab387"

func RenderPriorityName(priority int) string {
	switch priority {
	case 1:
		return "Urgent"
	case 2:
		return "High"
	case 3:
		return "Medium"
	case 4:
		return "Low"
	default:
		return "No Priority"
	}
}

func RenderPriorityStyle(priority int) string {
	s := RenderPriorityName(priority)

	style := lipgloss.NewStyle().Foreground(lipgloss.Color(PriorityColor))

	return style.Render(fmt.Sprintf("[%s]", s))
}
