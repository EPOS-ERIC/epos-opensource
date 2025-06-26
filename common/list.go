package common

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/db"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func PrintEnvironmentList(envs []db.Environment, title string) {
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

	for _, env := range envs {
		// TODO make this look better
		t.AppendSeparator()
		t.AppendRow(table.Row{env.Name, env.Directory, env.Platform})
	}

	rowIndex := -1
	highlight := text.Colors{text.FgGreen, text.Bold}
	t.SetRowPainter(func(row table.Row) text.Colors {
		rowIndex++
		if rowIndex == 0 {
			return highlight
		}
		return nil
	})

	t.AppendFooter(table.Row{"Copyright (C) 2023  EPOS ERIC", "Copyright (C) 2023  EPOS ERIC"}, rowMerge)

	fmt.Println(t.Render())
}
