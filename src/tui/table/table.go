package table

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/aliaksandrZh/worklog/src/internal/format"
	"github.com/aliaksandrZh/worklog/src/internal/model"
	"github.com/aliaksandrZh/worklog/src/tui"
)

// Columns defines the column order.
var Columns = []string{"date", "type", "number", "name", "timeSpent", "comments"}

// Config controls table rendering.
type Config struct {
	Width       int
	SortBy      string
	SortDir     string
	SelectedRow int // -1 = no selection
	SelectedCol int
	EditingCell bool   // true when a cell is being edited inline
	EditView    string // rendered textinput View() for inline editing
}

func colWidths(width int) (nameW, commentsW int) {
	numGaps := 5
	gap := len(format.Gap)
	fixedW := format.DateWidth + format.TypeWidth + format.NumberWidth + format.TimeSpentWidth + numGaps*gap
	remaining := width - fixedW
	if remaining < format.MinName+format.MinComments {
		remaining = format.MinName + format.MinComments
	}
	nameW = remaining * 65 / 100
	commentsW = remaining - nameW
	return
}

// ColWidth returns the width of a column given the total terminal width.
func ColWidth(col string, totalWidth int) int {
	if totalWidth <= 0 {
		totalWidth = 80
	}
	nameW, commentsW := colWidths(totalWidth)
	return getColWidth(col, nameW, commentsW)
}

func getColWidth(col string, nameW, commentsW int) int {
	switch col {
	case "date":
		return format.DateWidth
	case "type":
		return format.TypeWidth
	case "number":
		return format.NumberWidth
	case "name":
		return nameW
	case "timeSpent":
		return format.TimeSpentWidth
	case "comments":
		return commentsW
	}
	return 10
}

func getField(t model.Task, col string) string {
	switch col {
	case "date":
		return t.Date
	case "type":
		return t.Type
	case "number":
		return t.Number
	case "name":
		return t.Name
	case "timeSpent":
		return t.TimeSpent
	case "comments":
		return t.Comments
	}
	return ""
}

// Render produces a styled table string.
func Render(tasks []model.IndexedTask, cfg Config) string {
	if len(tasks) == 0 {
		return lipgloss.NewStyle().Faint(true).Render("No tasks to display.")
	}

	w := cfg.Width
	if w <= 0 {
		w = 80
	}
	nameW, commentsW := colWidths(w)

	indicator := "▲"
	if cfg.SortDir == "desc" {
		indicator = "▼"
	}

	// Build header
	var headerParts []string
	for _, col := range Columns {
		cw := getColWidth(col, nameW, commentsW)
		label := HeaderLabel(col)
		if cfg.SortBy == col {
			label += indicator
		}
		headerParts = append(headerParts, format.Pad(label, cw))
	}
	header := tui.TableHeaderStyle.Render(strings.Join(headerParts, format.Gap))

	// Build rows
	var rows []string
	for i, t := range tasks {
		isSelected := i == cfg.SelectedRow
		var cellParts []string

		for ci, col := range Columns {
			cw := getColWidth(col, nameW, commentsW)
			val := getField(t.Task, col)

			// Inline editing: render textinput in a fixed-width box to prevent column shift
			if isSelected && cfg.EditingCell && ci == cfg.SelectedCol {
				fixed := tui.EditingCellStyle.Width(cw).MaxWidth(cw).Render(cfg.EditView)
				cellParts = append(cellParts, fixed)
				continue
			}

			padded := format.Pad(val, cw)

			// Apply styling
			style := lipgloss.NewStyle()
			if col == "type" {
				style = style.Foreground(tui.TypeColor(val))
			}
			if isSelected && ci == cfg.SelectedCol {
				style = style.Reverse(true).Bold(true)
			}

			cellParts = append(cellParts, style.Render(padded))
		}

		rows = append(rows, strings.Join(cellParts, format.Gap))
	}

	return header + "\n" + strings.Join(rows, "\n")
}

// HeaderLabel returns the display label for a column.
func HeaderLabel(col string) string {
	switch col {
	case "date":
		return "Date"
	case "type":
		return "Type"
	case "number":
		return "Number"
	case "name":
		return "Name"
	case "timeSpent":
		return "Time"
	case "comments":
		return "Comments"
	}
	return col
}
