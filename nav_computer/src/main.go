package main

import (
	"fmt"
	"nav_computer/flight"
	"nav_computer/menu"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	app      menu.Application
	appModel tea.Model
	lip      lipgloss.Style
	height   int
	width    int
}

func New() tea.Model {
	lip := lipgloss.NewStyle().Margin(1, 2)

	return Model{
		app:      menu.MainMenu,
		appModel: menu.New(lip, 20, 20),
		lip:      lip,
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case menu.ChooseApp:
		m.app = msg.App

		switch m.app {
		case menu.MainMenu:
			m.appModel = menu.New(m.lip, m.height, m.width)
		case menu.FlightPlan:
			m.appModel = flight.New(m.lip, m.height, m.width)
		case menu.ExitMenu:
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.appModel, cmd = m.appModel.Update(msg)

	return m, cmd
}

func (m Model) View() string {
	return m.appModel.View()
}

func main() {
	_, err := tea.NewProgram(New(), tea.WithAltScreen()).Run()
	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
