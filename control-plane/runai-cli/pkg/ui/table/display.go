package table

import (
	"fmt"

	"golang.org/x/term"

	"github.com/run-ai/runai/control-plane/runai-cli/pkg/ui"

	"modernc.org/mathutil"

	"github.com/charmbracelet/bubbles/table"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	ErrorTableEmpty       = "%s table data is empty"
	ErrorNoSelection      = "no %s was selected"
	ErrorNoDataIDFound    = "didn't find data ID: %s"
	tableDataMargin       = 1
	tableDefaultHeight    = 5
	tableHeightMargin     = 3
	borderForegroundColor = "240"
	helpTextColor         = "240"
)

var baseStyle = lipgloss.NewStyle().
	BorderStyle(lipgloss.NormalBorder()).
	BorderForeground(lipgloss.Color(borderForegroundColor))

var helpStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(helpTextColor))

type Constructor interface {
	Subject() string
	Columns() []table.Column
	Data() [][]string
}

type Display struct {
	table       table.Model
	selectedRow table.Row
	constructor Constructor
}

func NewTableDisplay(constructor Constructor) *Display {
	return &Display{constructor: constructor}
}

func (td *Display) generateTable(interactive bool) table.Model {
	var rows []table.Row
	dataTable := td.constructor.Data()
	columns := td.constructor.Columns()
	for _, singleRow := range dataTable {
		var row []string
		for i := range singleRow {
			columns[i].Width = mathutil.Max(columns[i].Width, len(singleRow[i]))
			row = append(row, singleRow[i])
		}

		rows = append(rows, row)
	}

	var tableHeight int
	if interactive {
		tableHeight = calculateTableSize(len(dataTable))
	} else {
		tableHeight = len(dataTable)
	}

	t := table.New(
		table.WithColumns(columns),
		table.WithRows(rows),
		table.WithFocused(true),
		table.WithHeight(tableHeight),
	)

	style := table.DefaultStyles()
	style.Header = style.Header.
		BorderStyle(lipgloss.NormalBorder()).
		BorderBottom(true).
		Bold(false)

	if interactive {
		style.Selected = style.Selected.
			Bold(true)
	} else {
		style.Selected = style.Selected.
			Bold(false).
			UnsetForeground()
	}

	t.SetStyles(style)

	return t
}

func (td *Display) Generate() (string, error) {
	generateTable := td.generateTable(false)
	return generateTable.View(), nil
}

func (td *Display) Init() tea.Cmd {
	return nil
}

func (td *Display) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) { //nolint:gocritic // this is by design
	case tea.KeyMsg:
		switch msg.String() {
		case ui.KeyEsc:
			if td.table.Focused() {
				td.table.Blur()
			} else {
				td.table.Focus()
			}
		case ui.KeyQ, ui.CtrlPlusC:
			return td, tea.Quit
		case ui.KeyEnter:
			td.selectedRow = td.table.SelectedRow()
			return td, tea.Quit
		}
	}
	td.table, cmd = td.table.Update(msg)
	return td, cmd
}

func (td *Display) View() string {
	return ui.ForegroundGreenStyle.Render(fmt.Sprintf(ui.SelectClusterTableMsg, td.constructor.Subject())) +
		"\n" + baseStyle.Render(td.table.View()) + td.helpView()
}

func (td *Display) helpView() string {
	return helpStyle.Render("\n  ↑/↓: Navigate • q: Quit • ↵: Select\n")
}

func Selection(constructor Constructor) (*map[string]string, error) {
	data := constructor.Data()
	subject := constructor.Subject()
	columns := constructor.Columns()

	if len(data) == 0 {
		return nil, fmt.Errorf(ErrorTableEmpty, subject)
	}

	if len(columns) != len(data[0]) {
		return nil, fmt.Errorf(ErrorTableEmpty, subject)
	}

	bubble := NewTableDisplay(constructor)
	bubble.table = bubble.generateTable(true)
	if _, err := tea.NewProgram(bubble, tea.WithAltScreen()).Run(); err != nil {
		return nil, err
	}
	if bubble.selectedRow == nil {
		return nil, fmt.Errorf(ErrorNoSelection, subject)
	}
	for i := range data {
		if data[i][0] == bubble.selectedRow[0] {
			foundRow := make(map[string]string)
			for value := range data[i] {
				foundRow[columns[value].Title] = data[i][value]
			}
			return &foundRow, nil
		}
	}
	return nil, fmt.Errorf(ErrorNoDataIDFound, bubble.selectedRow[0])
}

func calculateTableSize(dataTableLength int) int {
	if !term.IsTerminal(0) {
		return tableDefaultHeight
	}

	_, termHeight, err := term.GetSize(0)
	if err != nil {
		return tableDefaultHeight
	}

	if termHeight <= tableDefaultHeight {
		return termHeight
	}

	if dataTableLength <= 0 {
		return tableDefaultHeight
	}

	if dataTableLength < termHeight-tableHeightMargin {
		return dataTableLength + tableDataMargin
	}

	return termHeight - tableHeightMargin
}
