package main

import (
	"fmt"
	"github.com/violetexistence/traveller/generator/sector"
	"os"

	tea "github.com/charmbracelet/bubbletea"
)

type sessionState int

const (
	mainMenu sessionState = iota
	sectorGenerator
)

type choice struct {
	shortcut string
	label    string
}

type model struct {
	choices   []choice
	cursor    int
	state     sessionState
	generator tea.Model
}

func initialModel() model {
	return model{
		choices: []choice{
			{shortcut: "s", label: "Sector"},
			{shortcut: "w", label: "World"},
			{shortcut: "p", label: "Ship"},
		},
		state: mainMenu,
	}
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.state == mainMenu {
				return m, tea.Quit
			} else {
				m.state = mainMenu
				m.generator = nil
				return m, nil
			}
		}
	}

	switch m.state {
	case mainMenu:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}

			case "s":
				m.state = sectorGenerator
				m.generator = sector.New()
				cmds = append(cmds, m.generator.Init())
			case "enter", " ":
				choice := m.choices[m.cursor]
				return m.Update(tea.KeyMsg{
					Type:  tea.KeyRunes,
					Runes: []rune(choice.shortcut),
				})
			}
		}
	case sectorGenerator:
		generator, cmd := m.generator.Update(msg)
		m.generator = generator
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	switch m.state {
	case sectorGenerator:
		return m.generator.View()
	default:
		s := "What should we generate?\n\n"

		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			s += fmt.Sprintf("%s [%s] %s\n", cursor, choice.shortcut, choice.label)
		}

		s += "\nPress q to quit.\n"

		return s
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
