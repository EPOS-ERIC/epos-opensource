package common

import (
	"fmt"
	"net/url"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
	copyright   = "Copyright (C) 2023  EPOS ERIC"
	logo        = `
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

func PrintError(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[ERROR]\t%s%s\n", colorRed, message, colorReset)
}

func PrintWarn(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[WARNING]\t%s%s\n", colorYellow, message, colorReset)
}

func PrintInfo(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[INFO]\t%s%s\n", colorBlue, message, colorReset)
}

func PrintStep(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[STEP]\t%s%s\n", colorCyan, message, colorReset)
}

func PrintDone(format string, a ...any) {
	message := fmt.Sprintf(format, a...)
	fmt.Printf("%s[DONE]\t%s%s\n", colorGreen, message, colorReset)
}

// PrintUrls prints the urls for the dataportal and the api gateway for a specific environment in the `dir` directory
func PrintUrls(portalURL, gatewayURL, backofficeURL, title string) {
	var err error
	gatewayURL, err = url.JoinPath(gatewayURL, "ui")
	if err != nil {
		panic("error building gateway url") // TODO
	}
	backofficeURL, err = url.JoinPath(backofficeURL, "home")
	if err != nil {
		panic("error building backoffice url") // TODO
	}

	t := table.NewWriter()
	t.SetTitle(title)
	t.Style().Title.Align = text.AlignCenter
	t.SetStyle(table.StyleRounded)
	t.Style().Title.Colors = text.Colors{text.FgYellow, text.Bold}
	t.Style().Color.Border = text.Colors{text.FgGreen}
	t.Style().Color.Footer = text.Colors{text.FgGreen}
	t.Style().Color.Separator = text.Colors{text.FgGreen}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Colors: text.Colors{text.FgYellow, text.Bold}},
		{Number: 2, Colors: text.Colors{text.FgHiCyan}},
	})

	rowMerge := table.RowConfig{
		AutoMerge:      true,
		AutoMergeAlign: text.AlignLeft,
	}
	// HACK: using a row with two logo columns so that they can be merged into one column
	t.AppendRow(table.Row{logo, logo}, rowMerge)

	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS Data Portal", portalURL})
	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS API Gateway", gatewayURL})
	t.AppendSeparator()
	t.AppendRow(table.Row{"EPOS Backoffice", backofficeURL})
	rowIndex := -1
	highlight := text.Colors{text.FgGreen, text.Bold}
	t.SetRowPainter(func(row table.Row) text.Colors {
		rowIndex++
		if rowIndex == 0 {
			return highlight
		}
		return nil
	})

	t.AppendFooter(table.Row{copyright, copyright}, rowMerge)

	fmt.Println(t.Render())
}
