package flight

import (
	"fmt"
	"math"
	"nav_computer/travellermap"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CreatePlanModel struct {
	lip                     lipgloss.Style
	height                  int
	width                   int
	savedSteps              map[stepId]tea.Model
	currentStepId           stepId
	currentStep             tea.Model
	originQuery             string
	originQueryResults      travellermap.SearchResults
	originWorld             travellermap.WorldDetail
	destinationQuery        string
	destinationQueryResults travellermap.SearchResults
	destinationWorld        travellermap.WorldDetail
	ship                    ShipDetail
	finishing               bool
}

func NewCreatePlan(lip lipgloss.Style, height int, width int) tea.Model {
	m := CreatePlanModel{
		lip:        lip,
		height:     height,
		width:      width,
		savedSteps: make(map[stepId]tea.Model),
	}

	m.currentStep = stepDefinitions[0](m)

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
	case ShipDetail:
		m.ship = msg
		return m, transition(NextMsg)
	}

	var cmd tea.Cmd
	m.currentStep, cmd = m.currentStep.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m CreatePlanModel) View() string {
	column := lipgloss.NewStyle().Width(40)
	left := m.currentStep.View()
	right := m.SummaryView()

	return lipgloss.JoinHorizontal(lipgloss.Top, column.Render(left), "   ", column.Render(right))
}

func (m CreatePlanModel) SummaryView() string {
	frameStyle := lipgloss.NewStyle().
		BorderStyle(lipgloss.NormalBorder()).
		BorderForeground(lipgloss.Color("63")).
		Padding(0, 1).
		MarginTop(1).
		Width(30)

	activeStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("164"))
	normalStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("250"))
	var sb strings.Builder

	for i := range summaries {
		step := summaries[i]
		label := summaryMap[step].label
		value := summaryMap[step].value(m)
		line := fmt.Sprintf("%s: %s", label, value)

		if m.currentStepId == step {
			sb.WriteString(activeStyle.Render(" -> " + line))
		} else {
			sb.WriteString(normalStyle.Render("    " + line))
		}
		sb.WriteString("\n")
	}

	return frameStyle.Render(sb.String())
}

func (m CreatePlanModel) Finish() (tea.Model, tea.Cmd) {
	m.finishing = true

	return m, func() tea.Msg {
		plan := FlightPlan{
			Origin:      m.originWorld,
			Destination: m.destinationWorld,
		}

		if outjump, err := computeJump(m.originWorld, m.ship); err == nil {
			plan.Outjump = *outjump
		} else {
			return err
		}

		if breakout, err := computeJump(m.destinationWorld, m.ship); err == nil {
			plan.Breakout = *breakout
		} else {
			return err
		}

		travelTime := plan.Outjump.TravelTime + 168 + plan.Breakout.TravelTime
		plan.EstTravelTime = int(math.Round(travelTime))
		plan.CreatedDate = time.Now()

		plan = CreateFlightPlan(plan)

		return CreatePlanFinishedMsg{
			result: PlanCreated,
			plan:   plan,
		}
	}
}

func computeJump(world travellermap.WorldDetail, ship ShipDetail) (*travellermap.JumpParams, error) {
	stellerClass := travellermap.ComputeSpectralClass(world)
	worldDiameter := travellermap.ComputeWorldDiameter(world)
	jumpParams := travellermap.ComputeJump(stellerClass, worldDiameter, ship.mRating)

	return &jumpParams, nil
}

func (m CreatePlanModel) Abort() (tea.Model, tea.Cmd) {
	return m, func() tea.Msg {
		return CreatePlanFinishedMsg{
			result: PlanCanceled,
		}
	}
}

func (m CreatePlanModel) Next() (tea.Model, tea.Cmd) {
	nextStepId := m.currentStepId + 1
	var cmd tea.Cmd

	if nextStep, ok := m.savedSteps[nextStepId]; ok {
		m.savedSteps[m.currentStepId] = m.currentStep
		m.currentStepId = nextStepId
		m.currentStep = nextStep
	} else if nextStepCreator, ok := stepDefinitions[nextStepId]; ok {
		m.savedSteps[m.currentStepId] = m.currentStep
		m.currentStepId = nextStepId
		nextStep := nextStepCreator(m)
		m.currentStep = nextStep
		cmd = m.currentStep.Init()
	} else {
		return m.Finish()
	}

	return m, cmd
}

func (m CreatePlanModel) Prev() (tea.Model, tea.Cmd) {
	prevStepId := m.currentStepId - 1
	if prevStep, ok := m.savedSteps[prevStepId]; ok {
		m.savedSteps[m.currentStepId] = m.currentStep
		m.currentStepId = prevStepId
		m.currentStep = prevStep
	} else {
		return m.Abort()
	}

	return m, func() tea.Msg {
		return ReturnToStepMsg{}
	}
}

type ReturnToStepMsg struct {
}

type InsertPlanMsg struct {
	FlightPlan FlightPlan
}

type FlightPlan struct {
	Id            int
	Origin        travellermap.WorldDetail
	Destination   travellermap.WorldDetail
	Outjump       travellermap.JumpParams
	Breakout      travellermap.JumpParams
	EstTravelTime int
	CreatedDate   time.Time
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
	shipDetailStep stepId = iota
	chooseOriginStep
	chooseDestinationStep
	finishStep
)

type StepCreator func(m CreatePlanModel) tea.Model

var stepDefinitions = map[stepId]StepCreator{
	chooseOriginStep:      createNewWorldSeach("Origin"),
	chooseDestinationStep: NewDestinationScreen,
	shipDetailStep:        NewShipDetail,
}

var summaries = [3]stepId{shipDetailStep, chooseOriginStep, chooseDestinationStep}
var summaryMap = map[stepId]struct {
	label string
	value func(m CreatePlanModel) string
}{
	chooseOriginStep: {
		label: "Choose Origin",
		value: func(m CreatePlanModel) string {
			return m.originWorld.Name
		},
	},
	chooseDestinationStep: {
		label: "Choose Destination",
		value: func(m CreatePlanModel) string {
			return m.destinationWorld.Name
		},
	},
	shipDetailStep: {
		label: "Ship Detail",
		value: func(m CreatePlanModel) string {
			return fmt.Sprintf("%vG", m.ship.mRating)
		},
	},
}

func createNewWorldSeach(title string) StepCreator {
	return func(model CreatePlanModel) tea.Model {
		return NewWorldSearch(model, title)
	}
}

func createNewDestinationScreen() StepCreator {
	return func(model CreatePlanModel) tea.Model {
		return NewDestinationScreen(model)
	}
}

type CreatePlanMsg struct {
}

type CreatePlanResult uint

const (
	PlanCreated CreatePlanResult = iota
	PlanCanceled
)

type CreatePlanFinishedMsg struct {
	result CreatePlanResult
	plan   FlightPlan
}
