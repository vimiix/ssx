package tui

import (
	"io"
	"os"

	"github.com/vimiix/tablewriter"
)

// PrintTable render a format table to stdout
func PrintTable(header []string, rows [][]string) {
	PrintTableTo(os.Stdout, header, rows)
}

func PrintTableTo(wr io.Writer, header []string, rows [][]string) {
	table := tablewriter.NewWriter(wr)
	table.SetHeader(header)
	table.SetAutoWrapText(false)
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.EnableBorder(false)
	table.SetAutoMergeCellsByColumnIndex([]int{0})
	table.AppendBulk(rows)
	table.Render()
}
