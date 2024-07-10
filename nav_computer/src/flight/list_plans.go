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

type FlightPlanItem struct {
	Id      int
	Origin  string
	Dest    string
	EstTime float64
}

func (p FlightPlanItem) Title() string { return fmt.Sprintf("%s to %s", p.Origin, p.Dest) }
func (p FlightPlanItem) Description() string {
	return fmt.Sprintf("Estimted travel time: %2.f hours", p.EstTime)
}
func (p FlightPlanItem) FilterValue() string { return fmt.Sprintf("%s %s", p.Origin, p.Dest) }

type ListPlansModel struct {
	list list.Model
	lip  lipgloss.Style
	keys *listKeyMap
}

func NewListModel(lip lipgloss.Style, height int, width int) tea.Model {
	items := []list.Item{}

	m := ListPlansModel{}
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

func (m *ListPlansModel) Resize(height int, width int) {
	h, w := m.lip.GetFrameSize()
	m.list.SetSize(width-w, height-h)
}

func (m ListPlansModel) Init() tea.Cmd {
	return loadFlightPlans()
}

func (m ListPlansModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Resize(msg.Height, msg.Width)
	case []FlightPlan:
		cmds = append(cmds, m.list.SetItems(createItems(msg)))
	case RefreshListMsg:
		cmds = append(cmds, loadFlightPlans())
	case InsertPlanMsg:
		item := FlightPlanItem{
			Origin:  msg.FlightPlan.Origin.Name,
			Dest:    msg.FlightPlan.Destination.Name,
			EstTime: msg.FlightPlan.Outjump.TravelTime + msg.FlightPlan.Breakout.TravelTime + 168,
		}
		m.list.InsertItem(0, item)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, menu.Open(menu.MainMenu)
		}

		switch {
		case key.Matches(msg, m.keys.deleteItem):
			if m.isVisiblySelected() {
				item := m.list.SelectedItem().(FlightPlanItem)
				cmds = append(cmds, deleteFlightPlan(item.Id))
			}
		case key.Matches(msg, m.keys.newItem):
			cmds = append(cmds, func() tea.Msg { return CreatePlanMsg{} })
		}

		list, cmd := m.list.Update(msg)
		m.list = list
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m ListPlansModel) isVisiblySelected() bool {
	visibleItems := m.list.VisibleItems()
	selectedItem := m.list.SelectedItem()
	for i := range visibleItems {
		if visibleItems[i] == selectedItem {
			return true
		}
	}

	return false
}

func (m ListPlansModel) View() string {
	return styles.Render(m.list.View())
}

type ListAllMsg struct {
}

type RefreshListMsg struct {
}

type PlanDeletedMsg struct {
	id int
}

func createItems(plans []FlightPlan) []list.Item {
	items := []list.Item{}
	for i := range plans {
		fp := plans[i]
		items = append(items, FlightPlanItem{
			Id:      fp.Id,
			Origin:  fp.Origin.Name,
			Dest:    fp.Destination.Name,
			EstTime: float64(fp.EstTravelTime),
		})
	}
	return items
}

func loadFlightPlans() tea.Cmd {
	return func() tea.Msg {
		return GetAllFlights()
	}
}

func deleteFlightPlan(id int) tea.Cmd {
	return func() tea.Msg {
		DeleteFlightPlan(id)
		return RefreshListMsg{}
	}
}
