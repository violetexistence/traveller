package sector

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type sector struct {
	name       string
	subsectors [9]subsector
}

type subsector struct {
	name   string
	worlds []world
}

type world struct {
	mainWorldName string
	location      string
	uwp           string
	bases         string
	travel        string
}

func New() tea.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return model{
		spinner: s,
		waiting: true,
	}
}

type model struct {
	spinner spinner.Model
	waiting bool
	sector  sector
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		generateSector(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case sector:
		m.sector = msg
		m.waiting = false
		m.spinner = spinner.New()
		return m, nil
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m model) View() string {
	if m.waiting {
		return fmt.Sprintf("\n\n%s Generating sector...press esc to cancel", m.spinner.View())
	} else {
		str := fmt.Sprintf("\n\nFinished. Sector: %s generated.", m.sector.name)
		for _, world := range m.sector.subsectors[0].worlds {
			str += fmt.Sprintf("\n%s %-20s %s", world.location, world.mainWorldName, world.uwp)
		}
		return str
	}
}

func generateSector() tea.Cmd {
	return func() tea.Msg {
		coin := newCoin()
		planets := newPlanets()
		var subsectors [9]subsector

		for ss := 0; ss < 8; ss++ { // subsectors across then down
			var worlds []world

			for x := 1; x < 9; x++ { // hexes left to right
				for y := 1; y < 11; y++ { // hexes top to bottom
					locationCode := getLocationCode(x, y)

					if coin.Toss() {
						population := dice(2) - 2
						starport := getStarportQuality(population)
						size := dice(2) - 2
						atmosphere := getAtmosphere(size)
						temperature := getSurfaceTemp(atmosphere)
						hydrographics := getHydrographics(size, atmosphere, temperature)
						government := getGovernment(population)
						law := getLawLevel(population, government)
						tech := getTechLevel(starport, size, atmosphere, hydrographics, population, government)
						uwp := strings.ToUpper(fmt.Sprintf("%s%x%x%x%x%x%x-%x", starport, size, atmosphere, hydrographics, population, government, law, tech))

						world := world{
							mainWorldName: planets.Name(),
							location:      locationCode,
							uwp:           uwp,
						}

						worlds = append(worlds, world)
					}
				}
			}

			subsector := subsector{
				name:   "bar",
				worlds: worlds,
			}
			subsectors[ss] = subsector
		}

		var sector = sector{
			name:       "foo",
			subsectors: subsectors,
		}

		return sector
	}
}

func getLocationCode(x int, y int) string {
	return fmt.Sprintf("%04d", x*100+y)
}

func getStarportQuality(population int) string {
	var dm int
	switch {
	case population == 8, population == 9:
		dm = 1
	case population > 9:
		dm = 2
	case population == 3, population == 4:
		dm = -1
	case population < 3:
		dm = -2
	}

	roll := dice(2) + dm

	switch {
	case roll > 10:
		return "A"
	case roll > 8:
		return "B"
	case roll > 6:
		return "C"
	case roll > 4:
		return "D"
	case roll > 2:
		return "E"
	default:
		return "X"
	}
}

func getAtmosphere(size int) int {
	roll := dice(2) - 7 + size

	return applyMinimum(roll, 0)
}

func getSurfaceTemp(atmosphere int) int {
	var dm int

	switch atmosphere {
	case 2, 3:
		dm = -2
	case 4, 5, 0xE:
		dm = -1
	case 8, 9:
		dm = 1
	case 0xA, 0xD, 0xF:
		dm = 2
	case 0xB, 0xC:
		dm = 6
	}

	habitableZoneLocation := dice(2)

	switch {
	case habitableZoneLocation > 9:
		dm += 4
	case habitableZoneLocation < 5:
		dm -= 4
	}

	return dice(2) + dm
}

func getHydrographics(size int, atmosphere int, temperature int) int {
	if size < 2 {
		return 0
	}

	var dm int

	switch atmosphere {
	case 0, 1, 0xA, 0xB, 0xC, 0xD, 0xE, 0xF:
		dm -= 4
	}

	if atmosphere != 0xD {
		switch {
		case temperature > 11: // Boiling
			dm -= 4
		case temperature > 9: // Hot
			dm -= 2
		}
	}

	roll := dice(2) - 7 + dm

	switch {
	case roll < 0:
		return 0
	default:
		return roll
	}
}

func getGovernment(population int) int {
	var result int

	if population > 0 {
		result = dice(2) - 7 + population
	}

	switch {
	case result < 0:
		return 0
	default:
		return result
	}
}

func getLawLevel(population int, government int) int {
	var result int

	if population > 0 {
		result = dice(2) - 7 + government
	}

	switch {
	case result < 0:
		return 0
	default:
		return result
	}
}

func getTechLevel(starport string, size int, atmosphere int, hydrographics int, population int, government int) int {
	roll := dice(1)

	var dm int

	switch starport {
	case "A":
		dm += 6
	case "B":
		dm += 4
	case "C":
		dm += 2
	case "X":
		dm -= 4
	}

	switch size {
	case 0, 1:
		dm += 2
	case 2, 3, 4:
		dm += 1
	}

	switch atmosphere {
	case 0, 1, 2, 3:
		dm += 1
	case 10, 11, 12, 13, 14, 15:
		dm += 1
	}

	switch hydrographics {
	case 0, 9:
		dm += 1
	case 10:
		dm += 2
	}

	switch population {
	case 1, 2, 3, 4, 5, 8:
		dm += 1
	case 9:
		dm += 2
	case 10:
		dm += 4
	}

	switch government {
	case 0, 5:
		dm += 1
	case 7:
		dm += 2
	case 13, 14:
		dm -= 2
	}

	result := roll + dm

	environmentalLimit := 0
	switch atmosphere {
	case 0, 1:
		environmentalLimit = 8
	case 2, 3:
		environmentalLimit = 5
	case 4, 7, 9:
		environmentalLimit = 3
	case 10:
		environmentalLimit = 8
	case 11:
		environmentalLimit = 9
	case 12:
		environmentalLimit = 10
	case 13, 14:
		environmentalLimit = 5
	case 15:
		environmentalLimit = 8
	}

	result = applyMinimum(result, environmentalLimit)
	result = applyMaximum(result, 15)

	return result
}

func applyMinimum(value int, floor int) int {
	if floor > value {
		return floor
	}
	return value
}

func applyMaximum(value int, ceiling int) int {
	if ceiling < value {
		return ceiling
	}
	return value
}

type cointoss struct {
	src       rand.Source
	cache     int64
	remaining int
}

func newCoin() *cointoss {
	return &cointoss{src: rand.NewSource(time.Now().UnixNano())}
}

func (c *cointoss) Toss() bool {
	if c.remaining == 0 {
		c.cache, c.remaining = c.src.Int63(), 63
	}

	result := c.cache&0x01 == 1
	c.cache >>= 1
	c.remaining--

	return result
}

func dice(d int) int {
	var result int
	for i := 0; i < d; i++ {
		result += rand.Intn(6) + 1
	}
	return result
}

type planetnames struct {
	names     []string
	remaining int
}

func newPlanets() *planetnames {
	return &planetnames{}
}

func (p *planetnames) Name() string {
	if p.remaining == 0 {
		p.names = getMorePlanetNames()
		p.remaining = len(p.names)
	}

	result := p.names[0]
	p.names = p.names[1:]
	p.remaining--

	return result
}

func getMorePlanetNames() []string {
	res, err := http.Get("https://donjon.bin.sh/name/rpc-name.fcgi?type=SciFi+World&n=10&as_json=1")
	if err != nil {
		log.Fatal(err)
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	var names []string
	json.Unmarshal(body, &names)

	if len(names) < 1 {
		log.Fatal("No names could be scraped")
	}

	return names
}
