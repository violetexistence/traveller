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

func Open(app Application) tea.Cmd {
	return func() tea.Msg {
		return ChooseApp{
			App: app,
		}
	}
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
	lip  lipgloss.Style
}

func New(lip lipgloss.Style, height int, width int) tea.Model {
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

	m := Model{
		list: list,
		lip:  lip,
	}
	m.Resize(height, width)

	return m
}

func (m *Model) Resize(height int, width int) {
	h, w := m.lip.GetFrameSize()
	m.list.SetSize(width-w, height-h)
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, Open(ExitMenu)
		case "enter":
			app := m.list.SelectedItem().(Item).app
			cmd := func() tea.Msg { return ChooseApp{App: app} }
			return m, cmd
		}
	case tea.WindowSizeMsg:
		h, w := m.lip.GetFrameSize()
		m.list.SetSize(msg.Width-w, msg.Height-h)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.lip.Render(m.list.View())
}
