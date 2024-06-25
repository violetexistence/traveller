package menu

import (
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Application uint

const (
	MainMenu Application = iota
	FlightPlan
	Comms
	LibraryData
	ExitMenu
)

type ChooseApp struct {
	App Application
}

type Item struct {
	title, desc string
	app         Application
}

func (i Item) Title() string       { return i.title }
func (i Item) Description() string { return i.desc }
func (i Item) FilterValue() string { return i.title }

type Model struct {
	list list.Model
}

var menuStyle = lipgloss.NewStyle().Margin(1, 2)

func New() tea.Model {
	items := []list.Item{
		Item{
			title: "Flight Plans",
			desc:  "Compute jump points, estimate travel time",
			app:   FlightPlan,
		},
		Item{
			title: "Communication",
			desc:  "Contact other ships or stations in system",
			app:   Comms,
		},
		Item{
			title: "Exit",
			desc:  "Close connection",
			app:   ExitMenu,
		},
	}

	list := list.New(items, list.NewDefaultDelegate(), 0, 0)
	list.Title = "Main Menu"

	return Model{
		list: list,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "enter":
			app := m.list.SelectedItem().(Item).app
			cmd := func() tea.Msg { return ChooseApp{App: app} }
			return m, cmd
		}
	case tea.WindowSizeMsg:
		h, w := menuStyle.GetFrameSize()
		m.list.SetSize(msg.Width-w, msg.Height-h)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return menuStyle.Render(m.list.View())
}
