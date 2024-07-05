package flight

import (
	"errors"
	"log"
	"math"
	"nav_computer/travellermap"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type CreatePlanModel struct {
	lip                     lipgloss.Style
	height                  int
	width                   int
	steps                   map[stepId]StepCreator
	savedSteps              map[stepId]tea.Model
	currentStepId           stepId
	currentStep             tea.Model
	originQuery             string
	originQueryResults      SearchResults
	originWorld             WorldItem
	destinationQuery        string
	destinationQueryResults SearchResults
	destinationWorld        WorldItem
	ship                    ShipDetail
	finishing               bool
}

func NewCreatePlan(lip lipgloss.Style, height int, width int) tea.Model {
	m := CreatePlanModel{
		lip:           lip,
		height:        height,
		width:         width,
		steps:         createSteps(),
		savedSteps:    make(map[stepId]tea.Model),
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
	case error:
		log.Println(msg.Error())
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
		return m.Finish()
	}

	var cmd tea.Cmd
	m.currentStep, cmd = m.currentStep.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m CreatePlanModel) View() string {
	log.Println("render create")
	return m.currentStep.View()
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

		plan = CreateFlightPlan(plan)

		return CreatePlanFinishedMsg{
			result: PlanCreated,
			plan:   plan,
		}
	}
}

func computeJump(world WorldItem, ship ShipDetail) (*travellermap.JumpParams, error) {
	if originDetail, err := travellermap.FetchWorldDetail(world.sector, world.hex); err == nil {
		stellerClass := travellermap.ComputeSpectralClass(*originDetail)
		worldDiameter := travellermap.ComputeWorldDiameter(*originDetail)
		jumpParams := travellermap.ComputeJump(stellerClass, worldDiameter, ship.mRating)

		return &jumpParams, nil
	} else {
		return nil, errors.Join(errors.New("Error requesting world detail from travellermap"), err)
	}
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
	if nextStep, ok := m.savedSteps[nextStepId]; ok {
		m.savedSteps[m.currentStepId] = m.currentStep
		m.currentStepId = nextStepId
		m.currentStep = nextStep
	} else if nextStepCreator, ok := m.steps[nextStepId]; ok {
		m.savedSteps[m.currentStepId] = m.currentStep
		m.currentStepId = nextStepId
		nextStep := nextStepCreator(m)
		m.currentStep = nextStep
	}

	return m, nil
}

func (m CreatePlanModel) Prev() (tea.Model, tea.Cmd) {
	prevStepId := m.currentStepId - 1
	if prevStep, ok := m.savedSteps[prevStepId]; ok {
		m.savedSteps[m.currentStepId] = m.currentStep
		m.currentStepId = prevStepId
		m.currentStep = prevStep
	} else if stepCreator, ok := m.steps[prevStepId]; ok {
		m.savedSteps[m.currentStepId] = m.currentStep
		m.currentStepId = prevStepId
		m.currentStep = stepCreator(m)
	} else {
		return m.Abort()
	}

	return m, nil
}

type InsertPlanMsg struct {
	FlightPlan FlightPlan
}

type FlightPlan struct {
	Id            int
	Origin        WorldItem
	Destination   WorldItem
	Outjump       travellermap.JumpParams
	Breakout      travellermap.JumpParams
	EstTravelTime int
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
		shipDetailStep:        NewShipDetail,
	}
}

func createNewWorldSeach(title string) StepCreator {
	return func(model CreatePlanModel) tea.Model {
		return NewWorldSearch(model, title)
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
