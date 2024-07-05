package flight

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

type ShipDetailModel struct {
	parent       CreatePlanModel
	form         huh.Form
	ship         ShipDetail
	errorMessage string
}

type ShipDetail struct {
	mRating float64
}

func (m ShipDetailModel) Init() tea.Cmd {
	return nil
}

func (m ShipDetailModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			return m, transition(PreviousMsg)
		}
	case error:
		m.errorMessage = msg.Error()
		return m, nil
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = *f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		m.ship.mRating = m.form.Get("mrating").(float64)
		cmds = append(cmds, func() tea.Msg { return m.ship })
	}

	return m, tea.Batch(cmds...)
}

func (m ShipDetailModel) View() string {
	if m.errorMessage == "" {
		return m.parent.lip.Render(m.form.View())
	} else {
		return m.parent.lip.Render("Hey")
	}
}

func NewShipDetail(model CreatePlanModel) tea.Model {
	m := ShipDetailModel{
		parent: model,
		form: *huh.NewForm(
			huh.NewGroup(
				huh.NewSelect[float64]().Title("M-Drive Rating").Description("(Acceleration)").Key("mrating").Options(
					huh.NewOption("0.5G", 0.5),
					huh.NewOption("1G", 1.0),
					huh.NewOption("2G", 2.0),
				),
			),
		),
	}

	m.form.Init()

	return m
}
