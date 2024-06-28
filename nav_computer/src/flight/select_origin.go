package flight

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
)

type SelectOriginModel struct {
	parent         CreatePlanModel
	list           list.Model
	selectedOrigin WorldItem
}

func (m SelectOriginModel) Init() tea.Cmd {
	return nil
}

func (m SelectOriginModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	cmds = append(cmds, cmd)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			m.selectedOrigin = m.list.SelectedItem().(WorldItem)
			cmds = append(cmds, selectWorld(m.selectedOrigin))
		}
	}

	return m, tea.Batch(cmds...)
}

func (m SelectOriginModel) View() string {
	return m.parent.lip.Render(m.list.View())
}

func (m SelectOriginModel) Merge(model CreatePlanModel) CreatePlanModel {
	return model
}

func NewSelectOrigin(model CreatePlanModel) tea.Model {
	all := model.originQueryResults.Results
	var items []list.Item

	for i := 0; i < all.Count; i++ {
		r := all.Items[i]
		if r.World != nil {
			items = append(items, WorldItem{
				name:   r.World.Name,
				sector: r.World.Sector,
				hex:    fmt.Sprintf("%02d", r.World.HexX) + fmt.Sprintf("%02d", r.World.HexY),
				uwp:    r.World.Uwp,
			})
		}
	}

	h, v := model.lip.GetFrameSize()
	list := list.New(items, list.NewDefaultDelegate(), model.width-h, model.height-v)
	list.Title = fmt.Sprintf("Matches for Origin: \"%s\"", model.originQuery)

	return SelectOriginModel{
		parent: model,
		list:   list,
	}
}

type WorldItem struct {
	name   string
	hex    string
	uwp    string
	sector string
}

func (w WorldItem) Title() string { return w.name }
func (w WorldItem) Description() string {
	return fmt.Sprintf("%s Sector, %s, UWP: %s", w.sector, w.hex, w.uwp)
}
func (w WorldItem) FilterValue() string {
	return w.Title() + w.Description()
}

type OriginWorldSelectedMsg struct {
	world WorldItem
}

func selectWorld(w WorldItem) tea.Cmd {
	return func() tea.Msg {
		return OriginWorldSelectedMsg{
			world: w,
		}
	}
}
