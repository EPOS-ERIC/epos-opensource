package common

import (
	"fmt"

	"github.com/epos-eu/epos-opensource/db"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func PrintInfraList(rows [][]any, headers []string, title string) {
	if len(rows) == 0 {
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
	colConfigs := make([]table.ColumnConfig, len(headers))
	for i := range headers {
		colConfigs[i] = table.ColumnConfig{Number: i + 1, Colors: text.Colors{text.FgHiCyan}, AlignHeader: text.AlignCenter}
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
	footer := make([]any, len(headers))
	for i := range footer {
		footer[i] = copyright
	}
	t.AppendFooter(table.Row(footer), rowMerge)
	fmt.Println(t.Render())
}

func PrintDockerList(dockers []db.Docker, title string) {
	rows := make([][]any, len(dockers))
	for i, d := range dockers {
		rows[i] = []any{d.Name, d.Directory, d.GuiUrl, d.BackofficeUrl, d.ApiUrl}
	}
	headers := []string{"Name", "Directory", "GUI URL", "Backoffice URL", "API URL"}
	PrintInfraList(rows, headers, title)
}

func PrintKubernetesList(kubes []db.Kubernetes, title string) {
	rows := make([][]any, len(kubes))
	for i, k := range kubes {
		rows[i] = []any{k.Name, k.Directory, k.Context, k.GuiUrl, k.BackofficeUrl, k.ApiUrl}
	}
	headers := []string{"Name", "Directory", "Context", "GUI URL", "Backoffice URL", "API URL"}
	PrintInfraList(rows, headers, title)
}
