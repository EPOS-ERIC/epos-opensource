package display

import (
	"fmt"
	"strings"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func InfraList(rows [][]any, headers []string, title string) {
	if len(rows) == 0 {
		Info("No installed environments found")
		return
	}

	columnHidden := make([]bool, len(headers))
	for col := range headers {
		emptyInAllRows := true
		for _, row := range rows {
			if col >= len(row) {
				continue
			}

			if row[col] == nil {
				continue
			}

			if strings.TrimSpace(fmt.Sprint(row[col])) != "" {
				emptyInAllRows = false
				break
			}
		}
		columnHidden[col] = emptyInAllRows
	}

	allHidden := true
	for _, hidden := range columnHidden {
		if !hidden {
			allHidden = false
			break
		}
	}
	if allHidden && len(columnHidden) > 0 {
		columnHidden[0] = false
	}

	t := table.NewWriter()
	t.SetTitle(title)
	t.SetStyle(table.StyleRounded)
	t.Style().Title.Align = text.AlignCenter
	t.Style().Title.Colors = text.Colors{text.FgYellow, text.Bold}
	t.Style().Color.Border = text.Colors{text.FgGreen}
	t.Style().Color.Footer = text.Colors{text.FgGreen}
	t.Style().Color.Separator = text.Colors{text.FgGreen}
	t.Style().Color.Header = text.Colors{text.FgCyan}
	colConfigs := make([]table.ColumnConfig, len(headers))
	for i := range headers {
		colConfigs[i] = table.ColumnConfig{Number: i + 1, AlignHeader: text.AlignCenter, Hidden: columnHidden[i]}
	}
	t.SetColumnConfigs(colConfigs)
	headerAny := make([]any, len(headers))
	for i, h := range headers {
		headerAny[i] = h
	}
	t.AppendHeader(table.Row(headerAny))
	for _, row := range rows {
		t.AppendRow(table.Row(row))
	}
	rowMerge := table.RowConfig{
		AutoMerge:      true,
		AutoMergeAlign: text.AlignLeft,
	}
	copyrightText := Copyright()
	footer := make([]any, len(headers))
	for i := range footer {
		footer[i] = copyrightText
	}
	t.AppendFooter(table.Row(footer), rowMerge)
	fmt.Println(t.Render())
}
