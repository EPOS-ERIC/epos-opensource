package tui

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/epos-eu/epos-opensource/common"
)

func (m Model) statusBarView() string {
	prefix := fmt.Sprintf("[%s]", m.state)
	helpText := m.helpString()
	suffix := fmt.Sprintf("EPOS Open source [%s]", common.GetVersion())

	suffixStyled := m.styleWithGradient(suffix)

	prefixWidth := lipgloss.Width(prefix)
	suffixWidth := lipgloss.Width(suffix)
	availableWidth := m.width - prefixWidth - suffixWidth - 4

	centeredHelp := lipgloss.PlaceHorizontal(availableWidth, lipgloss.Center, helpText)

	bar := lipgloss.JoinHorizontal(lipgloss.Top, prefix, "  ", centeredHelp, "  ", suffixStyled)

	return lipgloss.NewStyle().
		Background(lipgloss.Color("#2ecc71")).
		Foreground(lipgloss.Color("#000000")).
		Bold(true).
		Width(m.width).
		Render(bar)
}

func (m Model) helpString() string {
	bindings := m.keys.ShortHelp()
	var parts []string
	for _, b := range bindings {
		h := b.Help()
		parts = append(parts, fmt.Sprintf("%s: %s", strings.Split(h.Key, "/")[0], h.Desc))
	}
	return strings.Join(parts, " • ")
}

func (m Model) styleWithGradient(text string) string {
	startColor := "#008800"
	endColor := "#000000"
	var styledParts []string
	textLen := len(text)
	for i, r := range text {
		ratio := 0.0
		if textLen > 1 {
			ratio = float64(i) / float64(textLen-1)
		}
		color := interpolateColor(startColor, endColor, ratio)
		styledParts = append(styledParts, lipgloss.NewStyle().
			Foreground(lipgloss.Color(color)).
			Background(lipgloss.Color("#2ecc71")).
			Bold(true).
			Render(string(r)))
	}
	return strings.Join(styledParts, "")
}

func interpolateColor(start, end string, ratio float64) string {
	r1, g1, b1 := hexToRGB(start)
	r2, g2, b2 := hexToRGB(end)
	r := uint8(float64(r1) + (float64(r2)-float64(r1))*ratio)
	g := uint8(float64(g1) + (float64(g2)-float64(g1))*ratio)
	b := uint8(float64(b1) + (float64(b2)-float64(b1))*ratio)
	return fmt.Sprintf("#%02x%02x%02x", r, g, b)
}

func hexToRGB(hex string) (uint8, uint8, uint8) {
	if len(hex) != 7 || hex[0] != '#' {
		return 0, 0, 0
	}
	r, _ := strconv.ParseUint(hex[1:3], 16, 8)
	g, _ := strconv.ParseUint(hex[3:5], 16, 8)
	b, _ := strconv.ParseUint(hex[5:7], 16, 8)
	return uint8(r), uint8(g), uint8(b)
}
