package main

import (
	"fmt"
	"nav_computer/flight"
	"nav_computer/menu"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type Model struct {
	app      menu.Application
	appModel tea.Model
}

func New() tea.Model {
	return Model{
		app:      menu.MainMenu,
		appModel: menu.New(),
	}
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		}
	case menu.ChooseApp:
		m.app = msg.App

		switch m.app {
		case menu.FlightPlan:
			m.appModel = flight.New()
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
