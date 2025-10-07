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
	"net/url"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	reset     = "\033[0m"
	red       = "\033[31m"
	green     = "\033[32m"
	yellow    = "\033[33m"
	blue      = "\033[34m"
	cyan      = "\033[36m"
	copyright = "Copyright (C) 2025  EPOS ERIC"
	logo      = `
                                                 *************                              
&&&&&&&&&&&&&&&&&& *&&&&&&&%&&&%               *****************               &&&&&&/      
&&&&&&&&&&&&&&&&&& *&&&&&&&&&&&&&&&&&       **  **********  *******       &&&&&&&&&&&&&&&&& 
&&&&&&&&&&&%&&&&&& *&&&&&&&%    &&&&&&&   ,************     *********    &&%&&&&&&&&&&&&&   
&&&&&&             *&&&&&&        &&&&&( ************   **   ********** &&&&&&#             
&&&&&&             *&&&&&&(       &&&&& ****** * *****  **  *********** &&&&&&&&#           
&&&&&&&&&&&&&&&&.  *&&&&&&&&&&&&&&&&&&& *******   *   , *    *********** &&&&&&&&&&&&&&&&   
&&&%&&&&&&&%&&&&.  *&&&&&&&%&&&&&&&%&   *******                 ,*******    &&&&&&&%&&&&&&& 
&&&&&&             *&&&&&&               *                   , ********              &&&&&&.
&&&&&&             *&&&&&&               .    ******  *,    ******* **    &&         &&&&&& 
&&&&&&&&&&&&&&&&&& *&&&&&&                 ************** *         *   &&&&&&&&&&&&&&&&&&& 
&&&&&&&&&&&%&&&&&& *&&&&&&                   ************* ,*******     &&&%&&&&&&&&&&&&    
                                               ****************          &&&&&&&&&&&&       
`
)

// printStdout formats and prints a message with color, icon and label to standard out
func printStdout(color, label, format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[%s]  %s%s\n", color, label, reset, message)
}

func Error(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Fprintf(os.Stderr, "%s[%s]  %s%s\n", red, "ERROR", message, reset)
}

func Warn(format string, a ...any) {
	printStdout(yellow, "WARN", format, a...)
}

func Info(format string, a ...any) {
	printStdout(blue, "INFO", format, a...)
}

func Step(format string, a ...any) {
	printStdout(cyan, "STEP", format, a...)
}

func Done(format string, a ...any) {
	printStdout(green, "DONE", format, a...)
}

// Urls prints the URLs for the data portal, API gateway, and backoffice for a specific environment
func Urls(portalURL, gatewayURL, backofficeURL, title string) {
	if newGatewayURL, err := url.JoinPath(gatewayURL, "ui"); err != nil {
		Warn("Could not construct gateway URL: %v", err)
	} else {
		gatewayURL = newGatewayURL
	}

	if newBackofficeURL, err := url.JoinPath(backofficeURL, "home"); err != nil {
		Warn("Could not construct backoffice URL: %v", err)
	} else {
		backofficeURL = newBackofficeURL
	}

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

	// Merge config for logo and footer
	merge := table.RowConfig{AutoMerge: true, AutoMergeAlign: text.AlignLeft}

	// Add content
	t.AppendRow(table.Row{logo, logo}, merge)
	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS Platform Interface", portalURL})
	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS API Gateway", gatewayURL})
	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS Backoffice", backofficeURL})
	t.AppendFooter(table.Row{copyright, copyright}, merge)

	// Highlight first row (logo)
	rowIndex := -1
	t.SetRowPainter(func(row table.Row) text.Colors {
		rowIndex++
		if rowIndex == 0 {
			return text.Colors{text.FgGreen, text.Bold}
		}
		return nil
	})

	fmt.Println(t.Render())
}
