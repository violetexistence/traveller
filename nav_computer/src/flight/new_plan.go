package flight

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CreatePlanModel struct {
	lip                     lipgloss.Style
	height                  int
	width                   int
	steps                   map[stepId]StepCreator
	currentStepId           stepId
	currentStep             tea.Model
	originQuery             string
	originQueryResults      SearchResults
	originWorld             WorldItem
	destinationQuery        string
	destinationQueryResults SearchResults
	destinationWorld        WorldItem
}

func NewCreatePlan(lip lipgloss.Style, height int, width int) tea.Model {
	m := CreatePlanModel{
		lip:           lip,
		height:        height,
		width:         width,
		steps:         createSteps(),
		currentStepId: chooseOriginStep,
	}

	m.currentStep = m.steps[chooseOriginStep](m)

	return m
}

func (m CreatePlanModel) Init() tea.Cmd {
	return nil
}

func (m CreatePlanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case TransitionMsg:
		switch msg {
		case PreviousMsg:
			return m.Prev()
		case NextMsg:
			return m.Next()
		case FinishMsg:
			return m.Finish()
		case AbortMsg:
			return m.Abort()
		}
	case WorldSelectedMsg:
		if m.currentStepId == chooseOriginStep {
			m.originWorld = msg.world
		}
		if m.currentStepId == chooseDestinationStep {
			m.destinationWorld = msg.world
		}
		return m, transition(NextMsg)
	}

	var cmd tea.Cmd
	m.currentStep, cmd = m.currentStep.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m CreatePlanModel) View() string {
	return m.currentStep.View()
}

func (m CreatePlanModel) Finish() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return InsertPlanMsg{}
	}
}

func (m CreatePlanModel) Abort() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return ListAllMsg{}
	}
}

func (m CreatePlanModel) Next() (tea.Model, tea.Cmd) {
	if nextStepCreator, ok := m.steps[m.currentStepId+1]; ok {
		m.currentStepId++
		m.currentStep = nextStepCreator(m)
	}

	return m, nil
}

func (m CreatePlanModel) Prev() (tea.Model, tea.Cmd) {
	if stepCreator, ok := m.steps[m.currentStepId-1]; ok {
		m.currentStepId--
		m.currentStep = stepCreator(m)
	} else {
		return m.Abort()
	}

	return m, nil
}

type InsertPlanMsg struct {
}

type TransitionMsg uint

const (
	PreviousMsg TransitionMsg = iota
	NextMsg
	FinishMsg
	AbortMsg
)

func transition(msg TransitionMsg) tea.Cmd {
	return func() tea.Msg {
		return msg
	}
}

type stepId uint

const (
	chooseOriginStep stepId = iota
	chooseDestinationStep
	shipDetailStep
	finishStep
)

type StepCreator func(m CreatePlanModel) tea.Model

func createSteps() map[stepId]StepCreator {
	return map[stepId]StepCreator{
		chooseOriginStep:      createNewWorldSeach("Origin"),
		chooseDestinationStep: createNewWorldSeach("Destination"),
	}
}

func createNewWorldSeach(title string) StepCreator {
	return func(model CreatePlanModel) tea.Model {
		return NewWorldSearch(model, title)
	}
}

type CreatePlanMsg struct {
}
