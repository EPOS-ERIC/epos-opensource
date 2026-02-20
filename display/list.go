package display

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/EPOS-ERIC/epos-opensource/pkg/docker/db/sqlc"
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

func DockerList(dockers []sqlc.Docker, title string) {
	// TODO: Move Docker row shaping into cmd/docker (like K8s) once a shared list model exists.
	rows := make([][]any, len(dockers))
	for i, d := range dockers {
		gatewayURL, err := url.JoinPath(d.ApiUrl, "ui")
		if err != nil {
			Warn("Could not construct gateway URL: %v", err)
			gatewayURL = d.ApiUrl
		}
		var backofficeURL string
		if d.BackofficeUrl != nil {
			u, err := url.JoinPath(*d.BackofficeUrl, "home")
			if err != nil {
				Warn("Could not construct backoffice URL: %v", err)
				backofficeURL = *d.BackofficeUrl
			} else {
				backofficeURL = u
			}
		}
		rows[i] = []any{d.Name, d.Directory, d.GuiUrl, gatewayURL, backofficeURL}
	}
	headers := []string{"Name", "Directory", "GUI URL", "API URL", "Backoffice URL"}
	InfraList(rows, headers, title)
}
