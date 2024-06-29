package flight

import (
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CreatePlanModel struct {
	lip                lipgloss.Style
	height             int
	width              int
	steps              map[stepId]StepCreator
	currentStepId      stepId
	currentStep        tea.Model
	originQuery        string
	originQueryResults SearchResults
	originWorld        WorldItem
}

func NewCreatePlan(lip lipgloss.Style, height int, width int) tea.Model {
	m := CreatePlanModel{
		lip:           lip,
		height:        height,
		width:         width,
		steps:         createSteps(),
		currentStepId: originSearchStep,
	}

	m.currentStep = m.steps[originSearchStep](m)

	return m
}

func (m CreatePlanModel) Init() tea.Cmd {
	return nil
}

func (m CreatePlanModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			if m.currentStepId == 0 {
				return m, func() tea.Msg { return ListAllMsg{} }
			}
			return m, transition(PreviousMsg)
		}
	case TransitionMsg:
		switch msg {
		case PreviousMsg:
			return m.Prev(msg)
		case NextMsg:
			return m.Next(msg)
		case FinishMsg:
			return m.Finish(msg)
		case AbortMsg:
			return m.Abort(msg)
		}
	case OriginSearchResultsMsg:
		m.originQuery = msg.query
		m.originQueryResults = msg.results
		return m, transition(NextMsg)
	case OriginWorldSelectedMsg:
		m.originWorld = msg.world
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

func (m CreatePlanModel) Finish(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return InsertPlanMsg{}
	}
}

func (m CreatePlanModel) Abort(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return ListAllMsg{}
	}
}

func (m CreatePlanModel) Next(msg tea.Msg) (tea.Model, tea.Cmd) {
	if nextStepCreator, ok := m.steps[m.currentStepId+1]; ok {
		m.currentStepId++
		m.currentStep = nextStepCreator(m)
	} else {
		log.Printf("Next step is not defined: %d", m.currentStepId+1)
	}

	return m, nil
}

func (m CreatePlanModel) Prev(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.currentStepId--
	m.currentStep = m.steps[m.currentStepId](m)
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
	originSearchStep stepId = iota
	selectOriginStep
	destSearchStep
	selectDestStep
	shipDetailStep
	finishStep
)

type StepCreator func(m CreatePlanModel) tea.Model

func createSteps() map[stepId]StepCreator {
	return map[stepId]StepCreator{
		originSearchStep: NewOriginSearch,
		selectOriginStep: NewSelectOrigin,
	}
}

type CreatePlanMsg struct {
}
