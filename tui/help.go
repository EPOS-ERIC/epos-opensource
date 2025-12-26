package tui

import (
	"fmt"
	"strings"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// showHelp displays a modal with keyboard shortcuts for the current context.
func (a *App) showHelp() {
	a.PushFocus()
	prevContext := a.currentContext
	prevPage := a.currentPage
	a.UpdateFooter(GetFooterText(HelpKey), HelpKey)

	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetWordWrap(false).
		SetScrollable(true)
	textView.SetBorderPadding(1, 1, 2, 2)

	var title string
	var groupedHints map[string][]string

	switch prevContext {
	case "docker":
		title = "Docker Environments"
		groupedHints = getHelpHints("docker")
	case "k8s":
		title = "K8s Environments"
		groupedHints = getHelpHints("k8s")
	case "details-docker":
		title = "Environment Details (Docker)"
		groupedHints = getHelpHints("details-docker")
	case "details-k8s":
		title = "Environment Details (K8s)"
		groupedHints = getHelpHints("details-k8s")
	case FilePickerKey:
		title = "File Picker"
		groupedHints = getHelpHints("file-picker")
	default:
		title = "General"
		groupedHints = map[string][]string{"Generic": {"?: show this help", "q: quit application"}}
	}

	var content strings.Builder
	maxWidth := 0
	lineCount := 0

	lineCount += 2

	// Define group order
	groupOrder := []string{"Navigation", "Environment", "Browser", "Generic"}
	for _, group := range groupOrder {
		if hints, exists := groupedHints[group]; exists {
			// Group header
			content.WriteString(DefaultTheme.SecondaryTag("b") + group + "\n")
			lineCount++

			if len(group) > maxWidth {
				maxWidth = len(group)
			}

			// Separator
			separatorLen := 50
			content.WriteString(DefaultTheme.MutedTag("") + strings.Repeat("─", separatorLen) + "\n")
			lineCount++
			if separatorLen > maxWidth {
				maxWidth = separatorLen
			}

			// Hints
			for _, key := range hints {
				parts := strings.SplitN(key, ": ", 2)
				if len(parts) == 2 {
					line := fmt.Sprintf("  %s%-14s%s %s",
						DefaultTheme.PrimaryTag("b"),
						parts[0],
						DefaultTheme.Tag(DefaultTheme.OnSurface, ""),
						parts[1])
					content.WriteString(line + "\n")

					// Calculate plain text width
					plainLine := fmt.Sprintf("  %-14s %s", parts[0], parts[1])
					if len(plainLine) > maxWidth {
						maxWidth = len(plainLine)
					}
					lineCount++
				} else {
					content.WriteString(fmt.Sprintf("  %s\n", key))
					if len(key)+2 > maxWidth {
						maxWidth = len(key) + 2
					}
					lineCount++
				}
			}

			content.WriteString("\n") // Spacing between groups
			lineCount++
		}
	}

	// Now that we know maxWidth, create centered title
	titleLen := len(title)
	if titleLen > maxWidth {
		maxWidth = titleLen
	}
	padding := max((maxWidth-titleLen)/2, 0)
	centeredTitle := strings.Repeat(" ", padding) + DefaultTheme.PrimaryTag("b") + title + "\n\n"

	// Prepend centered title to content
	finalContent := centeredTitle + content.String()
	textView.SetText(finalContent)

	container := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, true)

	container.SetBorder(true).
		SetBorderColor(DefaultTheme.Secondary).
		SetTitle(" [::b]Help ").
		SetTitleColor(DefaultTheme.Secondary)

	container.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		if event.Key() == tcell.KeyEsc || event.Rune() == 'q' {
			a.pages.RemovePage("help")
			a.currentPage = prevPage
			a.PopFocus()
			footerKey := prevContext
			footerText := GetFooterText(footerKey)
			a.UpdateFooter(footerText, footerKey)
			return nil
		}
		return event
	})

	// Calculate dimensions
	// Border takes 2 chars width, internal padding takes 4 (2 left + 2 right)
	width := maxWidth + 6
	// Border takes 2 lines, internal padding takes 2 (1 top + 1 bottom)
	height := lineCount + 4

	// Cap at maximums
	if width > 70 {
		width = 70
	}
	if height > 25 {
		height = 25
	}

	// Ensure minimums
	if width < 55 {
		width = 55
	}
	if height < 12 {
		height = 12
	}

	a.UpdateFooter(GetFooterText(HelpKey), HelpKey)
	a.pages.AddPage("help", CenterPrimitiveFixed(container, width, height), true, true)
	a.currentPage = "help"
	a.tview.SetFocus(container)
}
