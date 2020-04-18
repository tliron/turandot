package common

import (
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/tliron/puccini/common/terminal"
	sshterminal "golang.org/x/crypto/ssh/terminal"
)

type Table struct {
	Headings []string
	Rows     [][][]string
	Style    TableStyle
	MaxWidth int

	HeadingSeparator       string
	RowSeparator           string
	TopDividerSeparator    string
	TopDivider             string
	MiddleDividerSeparator string
	MiddleDivider          string
	BottomDividerSeparator string
	BottomDivider          string
}

type TableStyle uint8

const (
	TopSeparatorTableStyle = 1 << iota
	ColumnSeparatorTableStyle
	RowSeparatorTableStyle
	BottomSeparatorTableStyle
)

func NewTable(headings ...string) *Table {
	var width int
	var err error
	if width, _, err = sshterminal.GetSize(int(os.Stdout.Fd())); err != nil {
		width = -1
	}

	return &Table{
		Headings: headings,
		Style:    TopSeparatorTableStyle | ColumnSeparatorTableStyle | RowSeparatorTableStyle | BottomSeparatorTableStyle,
		MaxWidth: width,

		// https://en.wikipedia.org/wiki/Box-drawing_character#DOS
		HeadingSeparator:       " ",
		RowSeparator:           "│",
		TopDividerSeparator:    "╤",
		TopDivider:             "═",
		MiddleDividerSeparator: "┼",
		MiddleDivider:          "─",
		BottomDividerSeparator: "┴",
		BottomDivider:          "─",
	}
}

func (self *Table) ColumnWidths() (int, []int) {
	columns := len(self.Headings)
	columnWidths := make([]int, columns)

	for index, heading := range self.Headings {
		width := len(heading)
		if width > columnWidths[index] {
			columnWidths[index] = width
		}
	}

	for _, row := range self.Rows {
		for index, cell := range row {
			width := cellWidth(cell)
			if width > columnWidths[index] {
				columnWidths[index] = width
			}
		}
	}

	return columns, columnWidths
}

func (self *Table) Add(cells ...string) {
	row := make([][]string, len(cells))
	for index, cell := range cells {
		row[index] = strings.Split(strings.Trim(cell, "\n"), "\n")
	}
	self.Rows = append(self.Rows, row)
}

func (self *Table) Rebuild() {
	// TODO: make sure it fits in maxwidth
}

func (self *Table) Write(writer io.Writer) {
	self.Rebuild()

	columns, columnWidths := self.ColumnWidths()
	rows := len(self.Rows)

	separator := func(divider string, dividerSeparator string) string {
		r := ""
		for column := 0; column < columns; column++ {
			r += strings.Repeat(divider, columnWidths[column])
			if column < columns-1 {
				r += dividerSeparator
			}
		}
		return r
	}

	for column, heading := range self.Headings {
		fmt.Fprint(writer, terminal.ColorTypeName(pad(heading, columnWidths[column])))
		if column < columns-1 {
			fmt.Fprint(writer, self.HeadingSeparator)
		}
	}
	fmt.Fprint(writer, "\n")

	if self.Style&TopSeparatorTableStyle != 0 {
		if self.Style&ColumnSeparatorTableStyle != 0 {
			fmt.Fprintln(writer, separator(self.TopDivider, self.TopDividerSeparator))
		} else {
			fmt.Fprintln(writer, separator(self.TopDivider, self.TopDivider))
		}
	}
	rowSeparator := separator(self.MiddleDivider, self.MiddleDividerSeparator)
	bottomSeparator := separator(self.BottomDivider, self.BottomDividerSeparator)

	for r, row := range self.Rows {
		height := rowHeight(row)
		for line := 0; line < height; line++ {
			for column, cell := range row {
				if line < len(cell) {
					fmt.Fprint(writer, terminal.ColorName(pad(cell[line], columnWidths[column])))
				} else {
					fmt.Fprint(writer, strings.Repeat(" ", columnWidths[column]))
				}
				if column < columns-1 {
					if self.Style&ColumnSeparatorTableStyle != 0 {
						fmt.Fprint(writer, self.RowSeparator)
					} else {
						fmt.Fprint(writer, " ")
					}
				}
			}
			fmt.Fprint(writer, "\n")
		}

		if r < rows-1 {
			if self.Style&RowSeparatorTableStyle != 0 {
				fmt.Fprintln(writer, rowSeparator)
			}
		} else if self.Style&BottomSeparatorTableStyle != 0 {
			fmt.Fprintln(writer, bottomSeparator)
		}
	}
}

func (self *Table) Print() {
	self.Write(terminal.Stdout)
}

func cellWidth(cell []string) int {
	width := 0
	for _, line := range cell {
		w := len(line)
		if w > width {
			width = w
		}
	}
	return width
}

func rowHeight(row [][]string) int {
	height := 0
	for _, cell := range row {
		h := len(cell)
		if h > height {
			height = h
		}
	}
	return height
}

func pad(s string, width int) string {
	return s + strings.Repeat(" ", width-len(s))
}
