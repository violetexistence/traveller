package main

import (
	"database/sql"
	"fmt"
	"log"
	"nav_computer/flight"
	"nav_computer/menu"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type Model struct {
	lip      lipgloss.Style
	appModel tea.Model
	app      menu.Application
	height   int
	width    int
	db       *sql.DB
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
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case error:
		log.Fatal(msg)
	case tea.WindowSizeMsg:
		m.height = msg.Height
		m.width = msg.Width
	case menu.ChooseApp:
		m.app = msg.App

		switch m.app {
		case menu.MainMenu:
			m.appModel = menu.New(m.lip, m.height, m.width)
			cmds = append(cmds, m.appModel.Init())
		case menu.FlightPlan:
			m.appModel = flight.New(m.lip, m.height, m.width)
			cmds = append(cmds, m.appModel.Init())
		case menu.ExitMenu:
			return m, tea.Quit
		}
	case sql.DB:
		m.db = &msg
	}

	var cmd tea.Cmd
	m.appModel, cmd = m.appModel.Update(msg)
	cmds = append(cmds, cmd)

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	return m.appModel.View()
}

func main() {
	logfilePath := os.Getenv("BUBBLETEA_LOG")

	if logfilePath == "" {
		logfilePath = "./navcom.log"
	}

	if _, err := tea.LogToFile(logfilePath, "simple"); err != nil {
		log.Fatal(err)
	}

	flight.CreateTables()

	_, err := tea.NewProgram(New(), tea.WithAltScreen()).Run()

	if err != nil {
		fmt.Println("Oh no:", err)
		os.Exit(1)
	}
}
