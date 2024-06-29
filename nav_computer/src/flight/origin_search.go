package flight

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"
)

type OriginSearchModel struct {
	parent    CreatePlanModel
	query     string
	form      huh.Form
	searching bool
	spinner   spinner.Model
	results   SearchResults
	err       error
}

func (m OriginSearchModel) Init() tea.Cmd {
	return nil
}

func (m OriginSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg.(type) {
	case SearchResults:
		m.searching = false
		m.results = msg.(SearchResults)
		return m, gotResults(m)
	case error:
		m.searching = false
		m.err = msg.(error)
	}

	if m.searching {
		log.Println("dispatching to spinner")
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	var cmds []tea.Cmd

	form, cmd := m.form.Update(msg)
	if f, ok := form.(*huh.Form); ok {
		m.form = *f
		cmds = append(cmds, cmd)
	}

	if m.form.State == huh.StateCompleted {
		m.query = m.form.GetString("origin")
		m.searching = true
		cmd := searchForWorld(m.query)
		cmds = append(cmds, m.spinner.Tick, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m OriginSearchModel) View() string {
	header := ""
	if m.err != nil {
		header += m.err.Error() + "\n\n"
	}

	if m.searching {
		return m.parent.lip.Render(m.spinner.View() + " Searching...")
	} else {
		return m.parent.lip.Render(header + m.form.View())
	}
}

func NewOriginSearch(m CreatePlanModel) tea.Model {
	log.Println("New origin search")
	model := OriginSearchModel{
		parent:  m,
		query:   m.originQuery,
		spinner: spinner.New(),
	}

	model.spinner.Spinner = spinner.Dot
	model.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	model.form = *huh.NewForm(
		huh.NewGroup(
			huh.NewInput().Key("origin").Title("Origin").Description("Enter main world name"),
		),
	)
	model.form.Init()

	return model
}

type OriginSearchResultsMsg struct {
	query   string
	results SearchResults
}

func gotResults(m OriginSearchModel) tea.Cmd {
	return func() tea.Msg {
		return OriginSearchResultsMsg{
			query:   m.query,
			results: m.results,
		}
	}
}

func searchForWorld(query string) tea.Cmd {
	return func() tea.Msg {
		log.Println("Executing world search on travellermap.com for " + query)

		resp, err := http.Get(fmt.Sprintf("https://travellermap.com/api/search?q=%s", query))
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != 200 {
			return errors.New(fmt.Sprintf("%d response from travellermap", resp.StatusCode))
		}
		body, _ := io.ReadAll(io.Reader(resp.Body))
		var results SearchResults
		if err = json.Unmarshal(body, &results); err != nil {
			return err
		}
		log.Printf("Response from travellermap.com with %d results", results.Results.Count)
		return results
	}
}

type SearchResults struct {
	Results struct {
		Count int `json:"Count"`
		Items []struct {
			World *struct {
				HexX       int    `json:"HexX"`
				HexY       int    `json:"HexY"`
				Sector     string `json:"Sector"`
				Uwp        string `json:"Uwp"`
				SectorX    int    `json:"SectorX"`
				SectorY    int    `json:"SectorY"`
				Name       string `json:"Name"`
				SectorTags string `json:"SectorTags"`
			} `json:"World,omitempty"`
			Label *struct {
				HexX       int    `json:"HexX"`
				HexY       int    `json:"HexY"`
				Scale      int    `json:"Scale"`
				SectorX    int    `json:"SectorX"`
				SectorY    int    `json:"SectorY"`
				Name       string `json:"Name"`
				SectorTags string `json:"SectorTags"`
			} `json:"Label,omitempty"`
		} `json:"Items"`
	} `json:"Results"`
}
