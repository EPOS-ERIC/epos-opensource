package common

import (
	"fmt"
	"slices"

	"github.com/epos-eu/epos-opensource/db"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func PrintEnvironmentList(envs []db.Environment, title string) {
	if len(envs) == 0 {
		PrintInfo("No installed environments found")
		return
	}

	t := table.NewWriter()
	t.SetTitle(title)
	t.SetStyle(table.StyleRounded)
	t.Style().Title.Align = text.AlignCenter
	t.Style().Title.Colors = text.Colors{text.FgYellow, text.Bold}
	t.Style().Color.Border = text.Colors{text.FgGreen}
	t.Style().Color.Footer = text.Colors{text.FgGreen}
	t.Style().Color.Separator = text.Colors{text.FgGreen}
	t.SetColumnConfigs([]table.ColumnConfig{
		{Number: 1, Colors: text.Colors{text.FgYellow, text.Bold}, AlignHeader: text.AlignCenter},
		{Number: 2, Colors: text.Colors{text.FgHiCyan}, AlignHeader: text.AlignCenter},
		{Number: 3, Colors: text.Colors{text.FgHiCyan}, AlignHeader: text.AlignCenter},
	})

	t.AppendHeader(table.Row{"Platform", "Name", "Directory"})

	// Sort environments by platform, then name
	sortedEnvs := make([]db.Environment, len(envs))
	copy(sortedEnvs, envs)
	slices.SortFunc(sortedEnvs, func(a, b db.Environment) int {
		if a.Platform == b.Platform {
			if a.Name < b.Name {
				return -1
			} else if a.Name > b.Name {
				return 1
			}
			return 0
		}
		if a.Platform < b.Platform {
			return -1
		}
		return 1
	})

	for _, env := range sortedEnvs {
		t.AppendRow(table.Row{env.Platform, env.Name, env.Directory})
	}

	rowMerge := table.RowConfig{
		AutoMerge:      true,
		AutoMergeAlign: text.AlignLeft,
	}
	t.AppendFooter(table.Row{
		copyright,
		copyright,
		copyright,
	}, rowMerge)

	fmt.Println(t.Render())
}
