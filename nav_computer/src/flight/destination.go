package flight

import (
	"fmt"
	"nav_computer/travellermap"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screenMode uint

const (
	idle screenMode = iota
	searching
	listMode
)

type DestinationScreen struct {
	lip            lipgloss.Style
	startingSector string
	startingHex    string
	jump           int
	worldsInRange  []travellermap.WorldDetail
	destination    *travellermap.WorldDetail
	spinner        spinner.Model
	list           list.Model
	mode           screenMode
}

func (m DestinationScreen) Init() tea.Cmd {
	return startScreen()
}

func (m DestinationScreen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case startMsg:
		m.spinner = createSpinner()
		m.mode = searching
		cmds = append(cmds, m.spinner.Tick, fetchWorldsInRange(m))
	case worldsInRangeMsg:
		m.mode = listMode
		m.worldsInRange = msg.worlds
		m.list = createList(m)
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, transition(PreviousMsg)
		case tea.KeyEnter:
			if m.mode == listMode {
				m.list, _ = m.list.Update(msg)
				selection := m.list.SelectedItem().(DestinationWorldItem)
				m.destination = &selection.world
				cmds = append(cmds, func() tea.Msg {
					return WorldSelectedMsg{
						world: *m.destination,
					}
				})
			}
		}
	}

	var cmd tea.Cmd
	switch m.mode {
	case searching:
		m.spinner, cmd = m.spinner.Update(msg)
	case listMode:
		m.list, cmd = m.list.Update(msg)
	}
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m DestinationScreen) View() string {
	switch m.mode {
	case searching:
		return m.spinner.View() + " Searching..."
	case listMode:
		return m.list.View()
	default:
		return ""
	}
}

func NewDestinationScreen(plan CreatePlanModel) tea.Model {
	return DestinationScreen{
		lip:            plan.lip,
		startingSector: plan.originWorld.Sector,
		startingHex:    plan.originWorld.Hex,
		jump:           plan.ship.jdrive,
	}
}

func startScreen() tea.Cmd {
	return func() tea.Msg {
		return startMsg{}
	}
}

func fetchWorldsInRange(m DestinationScreen) tea.Cmd {
	return func() tea.Msg {
		worlds, err := travellermap.FetchNearbyWorlds(m.startingSector, m.startingHex, m.jump)
		if err == nil {
			return worldsInRangeMsg{
				worlds: worlds,
			}
		} else {
			return err
		}
	}
}

type startMsg struct{}
type worldsInRangeMsg struct {
	worlds []travellermap.WorldDetail
}

func createSpinner() spinner.Model {
	result := spinner.New()
	result.Spinner = spinner.Dot
	result.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
	return result
}

func createList(m DestinationScreen) list.Model {
	all := m.worldsInRange
	var items []list.Item

	for i := 0; i < len(all); i++ {
		world := all[i]
		items = append(items, DestinationWorldItem{
			world: world,
		})
	}

	list := list.New(items, list.NewDefaultDelegate(), 40, 40)
	list.Title = fmt.Sprintf("Worlds within jump %d", m.jump)

	return list
}

type DestinationWorldItem struct {
	world travellermap.WorldDetail
}

func (item DestinationWorldItem) Title() string {
	return fmt.Sprintf("%s %s", item.world.Name, item.world.Uwp)
}

func (item DestinationWorldItem) Description() string {
	return fmt.Sprintf("%s %s", item.world.SectorAbbreviation, item.world.Hex)
}

func (item DestinationWorldItem) FilterValue() string {
	return item.world.Name
}
