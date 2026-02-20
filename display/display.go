// Package display provides simple, colorful console output functions for CLI applications.
//
// Example usage:
//
//	display.Error("Failed to connect: %v", err)
//	display.Info("Starting process...")
//	display.Step("Processing file %s", filename)
//	display.Done("All tasks completed successfully")
package display

import (
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

var (
	Stdout      io.Writer = os.Stdout
	Stderr      io.Writer = os.Stderr
	EnableDebug bool      = false
)

// ImageUpdateInfo holds information about an image update.
type ImageUpdateInfo struct {
	Name       string
	LastUpdate time.Time
}

const (
	resetSeq     = "\033[0m"
	boldSeq      = "\033[1m"
	underlineSeq = "\033[4m"
	redSeq       = "\033[31m"
	greenSeq     = "\033[32m"
	yellowSeq    = "\033[33m"
	blueSeq      = "\033[34m"
	purpleSeq    = "\033[35m"
	cyanSeq      = "\033[36m"
	logoHollow   = `⣠⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣠⣤⣶⣶⣶⣶⣶⣤⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣤⣤⣶⣶⣦⣤⣤⣀⡀⠀⠀⠀
⣿⣿⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⠛⣿⣿⠛⠛⠛⠛⠛⠛⠛⠛⠿⠿⣿⣿⣷⣄⠀⠀⠀⣀⣴⣿⣿⡿⠿⠛⠛⠛⠛⠛⠻⢿⣿⣿⣦⣄⠀⠀⠀⣠⣾⣿⣿⠿⠟⠛⠛⠛⠿⠿⣿⣿⣷⣤⡀
⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⢻⣿⣷⣀⣾⣿⡿⠛⠁⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⢿⣿⣷⣄⣾⣿⡟⠋⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⣻⣿⣷
⣿⣿⠀⠀⠀⠀⢠⣤⣤⣤⣤⣤⣤⣤⣤⣿⣿⠀⠀⠀⠀⢀⣴⣶⣷⣶⣄⠀⠀⠀⢻⣿⣿⣿⠏⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⣿⣿⣿⡟⠀⠀⠀⠀⣠⣤⣤⣤⣄⡀⣠⣾⣿⡿⠁
⣿⣿⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⢼⣿⣟⢙⣿⣿⠀⠀⠀⠘⣿⣿⠃⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠘⣿⣿⡇⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⡿⠋⠀⠀
⣿⣿⠀⠀⠀⠀⠈⠉⠉⠉⠉⠉⠉⢹⣿⣿⣿⠀⠀⠀⠀⠘⢿⣿⣿⣿⠟⠀⠀⠀⢸⣿⡏⠀⠀⠀⠀⠀⠀⠀⣠⣶⣶⣶⣤⠀⠀⠀⠀⠀⠀⠀⢸⣿⣇⠀⠀⠀⠀⠈⠉⠙⠛⠛⠿⢿⣿⣷⣄⠀
⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣰⣿⣿⡇⠀⠀⠀⠀⠀⠀⠰⣿⣿⣋⣿⣿⡇⠀⠀⠀⠀⠀⠀⢸⣿⣿⣦⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⢿⣿⣆
⣿⣿⠀⠀⠀⠀⢰⣶⣶⣶⣶⣶⣶⣾⣿⣿⣿⠀⠀⠀⠀⣀⣀⣀⣀⣠⣤⣴⣿⣿⣿⣿⣇⠀⠀⠀⠀⠀⠀⠀⠻⣿⣿⣿⠟⠀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣷⣶⣶⣦⣤⡀⠀⠀⠀⢸⣿⣿
⣿⣿⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⠀⣿⣿⡿⠿⠿⠿⠟⠛⠁⠸⣿⣿⡄⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⡀⠀⠀⠀⠀⠀⠀⢀⣿⣿⣿⣿⢿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⢠⣿⣿
⣿⣿⠀⠀⠀⠀⠈⠉⠉⠉⠉⠉⠉⠉⠉⣿⣿⠀⠀⠀⠀⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠹⣿⣿⣄⠀⠀⠀⠀⠀⢸⣿⣿⣿⣇⠀⠀⠀⠀⠀⢠⣾⣿⣿⠟⠁⠀⠀⠉⠉⠉⠉⠁⠀⠀⠀⠀⣼⣿⡟
⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⠀⠀⠀⠀⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠙⢿⣿⣷⣄⠀⠀⢀⣿⣿⠟⣿⣿⡄⠀⠀⣠⣴⣿⣿⣿⣷⣤⣀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣠⣾⣿⡟⠁
⣿⣿⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣶⣿⣿⣶⣶⣶⣶⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠙⢿⣿⣿⣶⣾⣿⡿⠀⢻⣿⣷⣶⣿⣿⡿⠋⠁⠉⠻⢿⣿⣿⣷⣶⣶⣶⣶⣾⣿⣿⣿⠟⠋⠀⠀
⠈⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠙⠛⠛⠋⠀⠀⠀⠙⠛⠛⠋⠁⠀⠀⠀⠀⠀⠀⠀⠉⠉⠛⠛⠛⠛⠋⠉⠁⠀⠀⠀⠀⠀`
	logoFull = ` ⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⣤⠀⠀⣤⣤⣤⣤⣤⣤⣤⣤⣀⣀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢀⣀⣤⣤⣤⣤⣤⣄⡀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣀⣠⣤⣤⣤⣀⣀⠀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⡄⠀⠀⠀⠀⠀⢀⣤⣾⣿⣿⣿⣿⣿⣿⣿⣿⣿⣷⣦⡀⠀⠀⠀⠀⠀⢠⣴⣿⣿⣿⣿⣿⣿⣿⣿⣿⣶⠄
⣿⣿⣿⣿⡟⠛⠛⠛⠛⠛⠛⠛⠛⠀⠀⣿⣿⣿⣿⡿⠋⠉⠈⠉⠻⣿⣿⣿⡄⠀⠀⠀⣰⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⠀⠀⠀⢠⣿⣿⣿⣿⠟⠛⠛⠛⠻⢿⠟⠁⠀
⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⡃⠀⠀⠀⠀⠀⣿⣿⣿⣧⠀⠀⣼⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣧⠀⠀⢸⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀
⣿⣿⣿⣿⣷⣶⣶⣶⣶⣶⣶⡆⠀⠀⠀⣿⣿⣿⣿⣧⡀⠀⠀⠀⣠⣿⣿⣿⡇⠀⢰⣿⣿⣿⣿⣿⣿⣿⠟⠉⠉⠉⠛⣿⣿⣿⣿⣿⣿⣿⡇⠀⠸⣿⣿⣿⣿⣷⣶⣦⣤⣤⣀⡀⠀⠀
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠏⠀⠀⢸⣿⣿⣿⣿⣿⣿⣏⠀⠀⠀⠀⠀⢸⣿⣿⣿⣿⣿⣿⡇⠀⠀⠙⠿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣦⡀
⣿⣿⣿⣿⡏⠉⠉⠉⠉⠉⠉⠁⠀⠀⠀⣿⣿⣿⣿⠿⠿⠿⠿⠟⠛⠋⠀⠀⠀⠀⠸⣿⣿⣿⣿⣿⣿⣿⣄⠀⠀⠀⣠⣿⣿⣿⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠈⠉⠉⠙⠛⢿⣿⣿⣿⡇
⣿⣿⣿⣿⡇⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⢻⣿⣿⣿⣿⣿⣿⣿⠀⠀⠀⢿⣿⣿⣿⣿⣿⣿⡿⠀⠀⠀⠀⡀⠀⠀⠀⠀⠀⠀⢸⣿⣿⣿⡟
⣿⣿⣿⣿⣷⣶⣶⣶⣶⣶⣶⣶⣶⠀⠀⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠻⣿⣿⣿⣿⣿⡇⠀⠀⠀⠸⣿⣿⣿⣿⣿⡟⠁⠀⠀⣠⣾⣿⣿⣶⣶⣶⣶⣾⣿⣿⣿⣿⠃
⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⣿⠀⠀⣿⣿⣿⣿⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠻⣿⣿⡿⠀⠀⠀⠀⠀⢻⣿⣿⠟⠋⠀⠀⠀⠈⠛⠿⣿⣿⣿⣿⣿⣿⣿⣿⡿⠟⠁⠀
⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠉⠀⠀⠉⠉⠉⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠉⠁⠀⠀⠀⠀⠀⠈⠉⠀⠀⠀⠀⠀⠀⠀⠀⠀⠀⠈⠉⠉⠉⠉⠁⠀⠀⠀⠀⠀ `
)

// printStdout formats and prints a message with color, icon and label to standard out
func printStdout(bold bool, color, label, format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	modifiers := color
	if bold {
		modifiers += boldSeq
	}
	_, _ = fmt.Fprintf(Stdout, "%s[%s]%s  %s\n", modifiers, label, resetSeq, message)
}

func Error(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	log.Printf(format, a...)
	_, _ = fmt.Fprintf(Stderr, "%s%s[%s] %s  %s\n", redSeq, boldSeq, "ERROR", message, resetSeq)
}

func Warn(format string, a ...any) {
	printStdout(true, yellowSeq, "WARN", format, a...)
}

func Info(format string, a ...any) {
	printStdout(false, blueSeq, "INFO", format, a...)
}

func Step(format string, a ...any) {
	printStdout(false, cyanSeq, "STEP", format, a...)
}

func Done(format string, a ...any) {
	printStdout(false, greenSeq, "DONE", format, a...)
}

func Debug(format string, a ...any) {
	if EnableDebug {
		printStdout(true, purpleSeq, "DEBUG", format, a...)
	}
}

func Copyright() string {
	return fmt.Sprintf("Copyright (C) %d  EPOS ERIC", time.Now().Year())
}

// URLs prints the URLs for the data portal, API gateway, and backoffice for a specific environment
func URLs(portalURL, gatewayURL, title string, backofficeURL *string) {
	t := table.NewWriter()
	t.SetTitle(title)
	t.SetStyle(table.StyleRounded)

	// Style configuration
	t.Style().Title.Align = text.AlignCenter
	t.Style().Title.Colors = text.Colors{text.FgYellow, text.Bold}
	t.Style().Color.Border = text.Colors{text.FgGreen}
	t.Style().Color.Footer = text.Colors{text.FgGreen}
	t.Style().Color.Separator = text.Colors{text.FgGreen}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Colors: text.Colors{text.FgYellow, text.Bold}},
	})

	// Add content
	t.AppendRow(table.Row{logoHollow, logoHollow}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignCenter})
	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS Platform UI", portalURL})
	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS API Gateway", gatewayURL})
	if backofficeURL != nil {
		t.AppendSeparator()
		t.AppendRow(table.Row{"EPOS Backoffice UI", *backofficeURL})
	}
	copyrightText := Copyright()
	t.AppendFooter(table.Row{copyrightText, copyrightText}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft})

	// Highlight first row (logo)
	rowIndex := -1
	t.SetRowPainter(func(row table.Row) text.Colors {
		rowIndex++
		if rowIndex == 0 {
			return text.Colors{text.FgGreen, text.Bold}
		}
		return nil
	})

	_, _ = fmt.Fprintf(Stdout, "%s\n", t.Render())
}

// UpdateAvailable prints a notification when a newer version of the CLI is available
func UpdateAvailable(currentVersion, latestVersion string) {
	t := table.NewWriter()
	t.SetTitle("Update Available")
	t.SetStyle(table.StyleRounded)

	// Style configuration
	t.Style().Title.Align = text.AlignCenter
	t.Style().Title.Colors = text.Colors{text.FgYellow, text.Bold}
	t.Style().Color.Border = text.Colors{text.FgYellow}
	t.Style().Color.Separator = text.Colors{text.FgYellow}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Colors: text.Colors{text.FgCyan, text.Bold}},
		{Number: 2, Colors: text.Colors{text.FgWhite}},
	})

	// Add content
	t.AppendRow(table.Row{"Current Version", currentVersion})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Latest Version", latestVersion})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Update Command", "Run 'epos-opensource update' to upgrade"})

	_, _ = fmt.Fprintf(Stdout, "%s\n", t.Render())
}

// UpdateStarting prints a table indicating the start of an update with version details
func UpdateStarting(oldVersion, newVersion string) {
	t := table.NewWriter()
	t.SetTitle("Starting Update")
	t.SetStyle(table.StyleRounded)

	t.Style().Title.Align = text.AlignCenter
	t.Style().Title.Colors = text.Colors{text.FgGreen, text.Bold}
	t.Style().Color.Border = text.Colors{text.FgGreen}
	t.Style().Color.Separator = text.Colors{text.FgGreen}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Colors: text.Colors{text.FgCyan, text.Bold}},
		{Number: 2, Colors: text.Colors{text.FgWhite}},
	})

	t.AppendRow(table.Row{"Current Version", oldVersion})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Target Version", newVersion})
	t.AppendSeparator()
	t.AppendRow(table.Row{"Release Notes", fmt.Sprintf("https://github.com/EPOS-ERIC/epos-opensource/releases/tag/%s", newVersion)})

	_, _ = fmt.Fprintf(Stdout, "%s\n", t.Render())
}

// ImageUpdatesAvailable prints a notification when Docker images have updates available
func ImageUpdatesAvailable(updates []ImageUpdateInfo, envName string) {
	if len(updates) == 0 {
		return
	}

	t := table.NewWriter()
	t.SetTitle("Image Updates Available")
	t.SetStyle(table.StyleRounded)

	// Style configuration
	t.Style().Title.Align = text.AlignCenter
	t.Style().Title.Colors = text.Colors{text.FgYellow, text.Bold}
	t.Style().Color.Border = text.Colors{text.FgYellow}
	t.Style().Color.Separator = text.Colors{text.FgYellow}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Colors: text.Colors{text.FgCyan, text.Bold}},
		{Number: 2, Colors: text.Colors{text.FgWhite}},
		{Number: 3, Colors: text.Colors{text.FgWhite}},
	})

	t.AppendHeader(table.Row{"Image", "Status", "Latest Update"})

	// Add content
	for _, update := range updates {
		t.AppendRow(table.Row{update.Name, "Update Available", update.LastUpdate.Format(time.RFC822)})
		t.AppendSeparator()
	}

	// Update instruction row
	prefix := "To update your environment with the new images, run: "
	command := fmt.Sprintf("epos-opensource docker update %s -u", envName)

	coloredPrefix := text.Colors{text.FgWhite}.Sprint(prefix)
	coloredCommand := text.Colors{text.FgGreen, text.Bold}.Sprint(command)

	instructionText := coloredPrefix + coloredCommand
	t.AppendRow(table.Row{instructionText, instructionText, instructionText}, table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft})

	// Style the instruction row
	rowIndex := -1
	t.SetRowPainter(func(row table.Row) text.Colors {
		rowIndex++
		if rowIndex == len(updates) {
			return text.Colors{text.FgGreen}
		}
		return nil
	})

	_, _ = fmt.Fprintf(Stdout, "%s\n", t.Render())
}
