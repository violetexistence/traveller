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
	jdrive  int
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
	case ReturnToStepMsg:
		return m.Restart()
	}

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = *f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		m.ship.mRating = m.form.Get("mrating").(float64)
		m.ship.jdrive = m.form.Get("jdrive").(int)
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

func (m ShipDetailModel) Restart() (tea.Model, tea.Cmd) {
	m.form = createForm(m)
	m.form.Init()

	return m, nil
}

func NewShipDetail(model CreatePlanModel) tea.Model {
	m := ShipDetailModel{
		parent: model,
	}
	m.form = createForm(m)
	m.form.Init()

	return m
}

func createForm(model ShipDetailModel) huh.Form {
	return *huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[float64]().Title("M-Drive Rating").Description("Acceleration in G").Key("mrating").Options(
				huh.NewOption("0.5G", 0.5).Selected(model.ship.mRating == 0.5),
				huh.NewOption("1G", 1.0).Selected(model.ship.mRating == 1.0),
				huh.NewOption("2G", 2.0).Selected(model.ship.mRating == 2.0),
			),
			huh.NewSelect[int]().Title("J-Drive Rating").Description("Jump distance in parsecs").Key("jdrive").Options(
				huh.NewOption("1", 1),
				huh.NewOption("2", 2),
				huh.NewOption("3", 3),
				huh.NewOption("4", 4),
				huh.NewOption("5", 5),
				huh.NewOption("6", 6),
			),
		),
	)
}
