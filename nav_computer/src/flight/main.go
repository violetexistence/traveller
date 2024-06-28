package flight

import (
	"log"
	"reflect"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type viewState uint

const (
	listView viewState = iota
	createView
)

type model struct {
	state     viewState
	viewModel tea.Model
	lip       lipgloss.Style
	height    int
	width     int
}

func New(lip lipgloss.Style, height int, width int) tea.Model {
	return model{
		state:     listView,
		viewModel: NewListModel(lip, height, width),
		lip:       lip,
		height:    height,
		width:     width,
	}
}

func (m model) Init() tea.Cmd {
	return func() tea.Msg { return ListAllMsg{} }
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	log.Printf("Flight received msg: %v", reflect.TypeOf(msg))

	var cmds []tea.Cmd

	switch msg.(type) {
	case ListAllMsg:
		m.state = listView
		m.viewModel = NewListModel(m.lip, m.height, m.width)
		return m, nil
	case CreatePlanMsg:
		log.Println("Received create plan msg")
		m.state = createView
		m.viewModel = NewCreatePlan(m.lip, m.height, m.width)
		return m, nil
	}

	var cmd tea.Cmd
	m.viewModel, cmd = m.viewModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	return m.viewModel.View()
}
