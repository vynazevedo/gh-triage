package ui

import (
	"strings"

	"github.com/charmbracelet/glamour"
)

var markdownStyle = "dark"

func SetMarkdownStyle(escuro bool) {
	if escuro {
		markdownStyle = "dark"
	} else {
		markdownStyle = "light"
	}
}

func Markdown(text string, width int) string {
	if width < 20 {
		width = 20
	}
	r, err := glamour.NewTermRenderer(
		glamour.WithStandardStyle(markdownStyle),
		glamour.WithWordWrap(width),
	)
	if err != nil {
		return text
	}
	out, err := r.Render(text)
	if err != nil {
		return text
	}
	return strings.Trim(out, "\n")
}
