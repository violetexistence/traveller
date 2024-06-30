package flight

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type WorldSearchModel struct {
	state   WorldSearchState
	title   string
	parent  CreatePlanModel
	query   string
	results SearchResults
	err     error
	input   textinput.Model
	list    list.Model
	spinner spinner.Model
}

type WorldSearchState uint

const (
	SearchEntryState WorldSearchState = iota
	WaitingState
	SelectState
)

func (m WorldSearchModel) Init() tea.Cmd {
	return nil
}

func (m WorldSearchModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case SearchResults:
		m.state = SelectState
		m.results = msg
		m.list = buildList(m)
	case error:
		m.state = SearchEntryState
		m.err = msg.(error)
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc, tea.KeyCtrlC:
			switch m.state {
			case SearchEntryState:
				return m, transition(PreviousMsg)
			case WaitingState, SelectState:
				m.state = SearchEntryState
				m.input.Focus()
			}
		}
	}

	var cmds []tea.Cmd

	switch m.state {
	case SearchEntryState:
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		cmds = append(cmds, cmd)

		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				m.query = m.input.Value()
				m.state = WaitingState
				cmd = searchForWorld(m.query)
				cmds = append(cmds, cmd)
			}
		}
	case WaitingState:
		log.Println("dispatching to spinner")
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	case SelectState:
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.Type {
			case tea.KeyEnter:
				selectedItem, isVisible := m.getVisibleSelection()
				if isVisible {
					cmds = append(cmds, selectWorld(selectedItem))
				}
			}
		}

		var cmd tea.Cmd
		m.list, cmd = m.list.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m WorldSearchModel) View() string {
	header := ""
	if m.err != nil {
		header += m.err.Error() + "\n\n"
	}

	switch m.state {
	case SearchEntryState:
		return m.parent.lip.Render(m.inputView())
	case WaitingState:
		return m.parent.lip.Render(m.spinner.View() + " Searching...")
	default:
		return m.parent.lip.Render(m.list.View())
	}
}

func (m WorldSearchModel) inputView() string {
	var sb strings.Builder

	sb.WriteString(lipgloss.NewStyle().Bold(true).Foreground(Indigo).Render(m.title))
	sb.WriteString("\n")
	sb.WriteString("Main world name")
	sb.WriteString("\n\n")
	sb.WriteString(m.input.View())
	sb.WriteString("\n\n")
	sb.WriteString(lipgloss.NewStyle().Foreground(Subdued).Render("enter - search, esc - go back"))

	return sb.String()
}

func (m WorldSearchModel) getVisibleSelection() (WorldItem, bool) {
	selectedItem := m.list.SelectedItem().(list.Item)
	isVisible := false

	for _, visibleItem := range m.list.VisibleItems() {
		if visibleItem == selectedItem {
			isVisible = true
		}
	}

	return selectedItem.(WorldItem), isVisible
}

func NewWorldSearch(m CreatePlanModel, title string) tea.Model {
	model := WorldSearchModel{
		parent:  m,
		title:   title,
		spinner: spinner.New(),
	}

	model.spinner.Spinner = spinner.Dot
	model.spinner.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	model.input = textinput.New()
	model.input.PromptStyle = lipgloss.NewStyle().Foreground(Red)
	model.input.Focus()
	model.input.CharLimit = 156
	model.input.Width = m.width
	model.input.Cursor.Style = lipgloss.NewStyle().Foreground(Green)

	return model
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

func buildList(m WorldSearchModel) list.Model {
	all := m.results.Results
	var items []list.Item

	for i := 0; i < all.Count; i++ {
		r := all.Items[i]
		if r.World != nil {
			items = append(items, WorldItem{
				name:   r.World.Name,
				sector: r.World.Sector,
				hex:    fmt.Sprintf("%02d", r.World.HexX) + fmt.Sprintf("%02d", r.World.HexY),
				uwp:    r.World.Uwp,
			})
		}
	}

	h, v := m.parent.lip.GetFrameSize()
	list := list.New(items, list.NewDefaultDelegate(), m.parent.width-h, m.parent.height-v)
	list.Title = fmt.Sprintf("Matches for %s: \"%s\"", m.title, m.query)

	return list
}

type WorldItem struct {
	name   string
	hex    string
	uwp    string
	sector string
}

func (w WorldItem) Title() string { return w.name }
func (w WorldItem) Description() string {
	return fmt.Sprintf("%s Sector, %s, UWP: %s", w.sector, w.hex, w.uwp)
}
func (w WorldItem) FilterValue() string {
	return w.Title() + w.Description()
}

type WorldSelectedMsg struct {
	world WorldItem
}

func selectWorld(world WorldItem) tea.Cmd {
	return func() tea.Msg {
		return WorldSelectedMsg{
			world: world,
		}
	}
}
