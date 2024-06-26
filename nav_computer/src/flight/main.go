package flight

import (
	"fmt"
	"nav_computer/menu"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var styles = lipgloss.NewStyle().Margin(1, 2)

type listKeyMap struct {
	newItem    key.Binding
	deleteItem key.Binding
	open       key.Binding
}

func newListKeyMap() *listKeyMap {
	return &listKeyMap{
		newItem: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "new flight plan"),
		),
		deleteItem: key.NewBinding(
			key.WithKeys("x", "delete"),
			key.WithHelp("del/x", "delete"),
		),
		open: key.NewBinding(
			key.WithKeys("enter", "O"),
			key.WithHelp("enter/O", "open"),
		),
	}
}

type FlightPlan struct {
	Origin  string
	Dest    string
	EstTime float64
}

func (p FlightPlan) Title() string { return fmt.Sprintf("%s to %s", p.Origin, p.Dest) }
func (p FlightPlan) Description() string {
	return fmt.Sprintf("Estimted travel time: %2.f hours", p.EstTime)
}
func (p FlightPlan) FilterValue() string { return fmt.Sprintf("%s %s", p.Origin, p.Dest) }

type Model struct {
	Flights []FlightPlan
	list    list.Model
	lip     lipgloss.Style
	keys    *listKeyMap
}

func New(lip lipgloss.Style, height int, width int) Model {
	m := Model{}
	items := []list.Item{
		FlightPlan{
			Origin:  "Trindel",
			Dest:    "Archipelago",
			EstTime: 123.4,
		},
		FlightPlan{
			Origin:  "Marina",
			Dest:    "Trindel",
			EstTime: 39,
		},
		FlightPlan{
			Origin:  "Paladin",
			Dest:    "Marina",
			EstTime: 66.54,
		},
		FlightPlan{
			Origin:  "Haro",
			Dest:    "Paladin",
			EstTime: 381.4,
		},
	}

	m.keys = newListKeyMap()

	m.list = list.New(items, list.NewDefaultDelegate(), 0, 0)
	m.list.DisableQuitKeybindings()
	m.list.Title = "Flight Plans"
	m.list.AdditionalFullHelpKeys = func() []key.Binding {
		return []key.Binding{
			m.keys.open,
			m.keys.deleteItem,
			m.keys.newItem,
		}
	}
	m.lip = lip

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
	case tea.WindowSizeMsg:
		m.Resize(msg.Height, msg.Width)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, menu.Open(menu.MainMenu)
		}

		switch {
		case key.Matches(msg, m.keys.deleteItem):
			selected := m.list.SelectedItem()
			m.list.RemoveItem(m.list.Index())
			return m, m.list.NewStatusMessage(fmt.Sprintf("Deleted %s", selected.(FlightPlan).Title()))
		case key.Matches(msg, m.keys.newItem):

		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return styles.Render(m.list.View())
}
