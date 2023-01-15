package cmd

import (
	"strings"
)

func FormatTable(table [][]string, sepColumn, sepRow, prefixRow string) string {
	var lengths []int
	for _, row := range table {
		for i, column := range row {
			if l := len(lengths); l <= i {
				lengths = append(lengths, make([]int, i-l+1)...)
			}
			if l := lengths[i]; l < len(column) {
				lengths[i] = len(column)
			}
		}
	}
	var rows []string
	for _, row := range table {
		var formatted []string
		last := len(row) - 1
		for i, column := range row {
			if i == last {
				// No need to pad if no other column.
				formatted = append(formatted, column)
			} else {
				formatted = append(formatted, padRight(column, lengths[i]))
			}
		}
		rows = append(rows, prefixRow+strings.Join(formatted, sepColumn))
	}
	return strings.Join(rows, sepRow)
}

func padRight(s string, n int) string {
	return s + strings.Repeat(" ", n-len(s))
}
