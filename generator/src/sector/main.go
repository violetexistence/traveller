package sector

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type worldSize int

const (
	worldSize_0 worldSize = iota
	worldSize_1
	worldSize_2
	worldSize_3
	worldSize_4
	worldSize_5
	worldSize_6
	worldSize_7
	worldSize_8
	worldSize_9
	worldSize_A
)

type atmosphereType int

const (
	atmosphere_0 atmosphereType = iota
	atmosphere_1
	atmosphere_2
	atmosphere_3
	atmosphere_4
	atmosphere_5
	atmosphere_6
	atmosphere_7
	atmosphere_8
	atmosphere_9
	atmosphere_A
	atmosphere_B
	atmosphere_C
	atmosphere_D
	atmosphere_E
	atmosphere_F
)

type hydrographicType int

const (
	hydrographics_0 hydrographicType = iota
	hydrographics_1
	hydrographics_2
	hydrographics_3
	hydrographics_4
	hydrographics_5
	hydrographics_6
	hydrographics_7
	hydrographics_8
	hydrographics_9
	hydrographics_A
)

type populationType int

const (
	population_0 populationType = iota
	population_1
	population_2
	population_3
	population_4
	population_5
	population_6
	population_7
	population_8
	population_9
	population_A
)

type governmentType int

const (
	government_0 governmentType = iota
	government_1
	government_2
	government_3
	government_4
	government_5
	government_6
	government_7
	government_8
	government_9
	government_A
	government_B
	government_C
	government_D
)

type lawLevel int

const (
	lawlevel_0 lawLevel = iota
	lawlevel_1
	lawlevel_2
	lawlevel_3
	lawlevel_4
	lawlevel_5
	lawlevel_6
	lawlevel_7
	lawlevel_8
	lawlevel_9
)

type starportClass string

const (
	starport_X starportClass = "X"
	starport_E               = "E"
	starport_D               = "D"
	starport_C               = "C"
	starport_B               = "B"
	starport_A               = "A"
)

type techLevel int

const (
	technology_0 techLevel = iota
	technology_2
	technology_3
	technology_4
	technology_5
	technology_6
	technology_7
	technology_8
	technology_9
	technology_A
	technology_B
	technology_C
	technology_D
	technology_E
	technology_F
	technology_X
)

type sector struct {
	name  string
	hexes []hexInfo
}

type hexInfo struct {
	name       string
	location   string
	hz         int
	uwp        string
	bases      string
	remarks    string
	zone       string
	PBG        string
	allegiance string
	stars      string
	ix         string
	ex         string
	cx         string
	nobility   string
	worlds     int
	primary    star
}

var coin = newCoin()

type star struct {
	class spectralClass
	size  string
}

func getPrimary() star {
	class := spectralClass{
		letter:  getSpectralType(flux()),
		numeral: rollDecimal(0, 9),
	}

	size := getSpectralSize(class)

	return star{
		class: class,
		size:  size,
	}
}

func rollDecimal(min int, max int) int {
	if max <= min {
		panic(fmt.Sprintf("max must be greater than min {%d, %d}", min, max))
	}

	return rand.Intn(max-min+1) + min
}

func getSpectralType(fluxValue int) string {
	row := fluxValue + 6
	return spectralSizeMatrix[row][0]
}

var spectralSizeMatrix = [15][8]string{
	{"OB", "Ia", "Ia", "Ia", "II", "II", "II", "II"},
	{"A", "Ia", "Ia", "Ia", "II", "II", "II", "II"},
	{"A", "Ib", "Ib", "Ib", "III", "III", "III", "II"},
	{"F", "II", "II", "II", "IV", "IV", "IV", "II"},
	{"F", "III", "III", "III", "V", "V", "V", "III"},
	{"G", "III", "III", "IV", "V", "V", "V", "V"},
	{"G", "III", "III", "V", "V", "V", "V", "V"},
	{"K", "V", "III", "V", "V", "V", "V", "V"},
	{"K", "V", "V", "V", "V", "V", "V", "V"},
	{"M", "V", "V", "V", "V", "V", "V", "V"},
	{"M", "IV", "IV", "V", "VI", "VI", "VI", "VI"},
	{"M", "D", "D", "D", "D", "D", "D", "D"},
	{"BD", "IV", "IV", "V", "VI", "VI", "VI", "VI"},
	{"BD", "IV", "IV", "V", "VI", "VI", "VI", "VI"},
	{"BD", "IV", "IV", "V", "VI", "VI", "VI", "VI"},
}

var spectralSizeMatrixColumns = map[string]int{
	"Sp": 0,
	"O":  1,
	"B":  2,
	"A":  3,
	"F":  4,
	"G":  5,
	"K":  6,
	"M":  7,
}

type spectralClass struct {
	letter  string // O-M
	numeral int    // 0-9
}

func getSpectralSize(class spectralClass) string {
	row := flux() + 6
	col := spectralSizeMatrixColumns[class.letter]
	return spectralSizeMatrix[row][col]
}

func newSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	return s
}

func New() tea.Model {
	return model{
		help:    help.New(),
		spinner: newSpinner(),
		waiting: true,
		message: "Generating sector data...",
	}
}

type model struct {
	help    help.Model
	spinner spinner.Model
	waiting bool
	message string
	sector  sector
	sub     int
}

type keyMap struct {
	Prev key.Binding
	Next key.Binding
	Save key.Binding
}

func (k keyMap) shortHelp() []key.Binding {
	return []key.Binding{k.Prev, k.Next, k.Save}
}

var defaultKeyMap = keyMap{
	Prev: key.NewBinding(
		key.WithKeys("h", "left"),
		key.WithHelp("←/h", "prev"),
	),
	Next: key.NewBinding(
		key.WithKeys("l", "right"),
		key.WithHelp("→/l", "next"),
	),
	Save: key.NewBinding(
		key.WithKeys("s"),
		key.WithHelp("s", "save"),
	),
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		m.spinner.Tick,
		generateSector(),
	)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch {
		case key.Matches(msg, defaultKeyMap.Prev):
			m.sub = applyMinimum(m.sub-1, 0)
		case key.Matches(msg, defaultKeyMap.Next):
			m.sub = applyMaximum(m.sub+1, 15)
		case key.Matches(msg, defaultKeyMap.Save):
			m.waiting = true
			m.message = "Saving sector data..."
			cmds = append(cmds,
				m.spinner.Tick,
				saveSector(m.sector),
			)
		}
	case sector:
		m.sector = msg
		m.waiting = false
		m.spinner = newSpinner()
	case saveSuccessful:
		m.waiting = false
		m.spinner = newSpinner()
	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m model) View() string {
	if m.waiting {
		return fmt.Sprintf("\n\n%s %s", m.spinner.View(), m.message)
	} else {
		str := fmt.Sprintf("\n\n%s Sector\n", m.sector.name)
		for _, world := range m.sector.hexes {
			str += fmt.Sprintf("\n%s %-20s %s", world.location, world.name, world.uwp)
		}
		str += fmt.Sprintf("\n %s", m.help.ShortHelpView(defaultKeyMap.shortHelp()))
		return str
	}
}

func generateSector() tea.Cmd {
	return func() tea.Msg {
		coin := newCoin()
		planets := newPlanets()

		var worlds []hexInfo

		for hx := 1; hx <= 32; hx++ {
			for hy := 1; hy <= 40; hy++ {
				locationCode := getLocationCode(hx, hy)

				if coin.Toss() {
					population := populationType(dice(2) - 2)
					starport := getStarportQuality(population)
					size := worldSize(dice(2) - 2)
					atmosphere := getAtmosphere(size)
					temperature := getSurfaceTemp(atmosphere)
					hydrographics := getHydrographics(size, atmosphere, temperature)
					government := getGovernment(population)
					law := getLawLevel(population, government)
					tech := getTechLevel(starport, size, atmosphere, hydrographics, population, government)
					uwp := strings.ToUpper(fmt.Sprintf("%s%x%x%x%x%x%x-%x", starport, size, atmosphere, hydrographics, population, government, law, tech))

					bases := getBases(starport)

					primary := getPrimary()

					hex := hexInfo{
						name:     planets.Name(),
						location: locationCode,
						uwp:      uwp,
						bases:    bases,
						primary:  primary,
					}

					worlds = append(worlds, hex)
				}
			}
		}

		var sector = sector{
			name:  planets.Name(),
			hexes: worlds,
		}

		return sector
	}
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func saveSector(sector sector) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Create(sector.name)
		check(err)
		defer f.Close()

		_, headerErr := f.WriteString("Hex\tName\tUWP\tBases\tRemarks\tZone\tPBG\tAllegiance\tStars\t{Ix}\t(Ex)\t[Cx]\tNobility\tW\n")
		check(headerErr)

		for _, w := range sector.hexes {
			_, worldErr := f.WriteString(
				fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
					w.location,
					w.name,
					w.uwp,
					w.bases,
					w.remarks,
					w.zone,
					w.PBG,
					w.allegiance,
					w.stars,
					w.ix,
					w.ex,
					w.cx,
					w.nobility,
					w.worlds,
				))
			check(worldErr)
		}

		f.Sync()

		return saveSuccessful{
			filename: f.Name(),
			worlds:   len(sector.hexes),
		}
	}
}

type saveSuccessful struct {
	filename string
	worlds   int
}

func getLocationCode(x int, y int) string {
	return fmt.Sprintf("%04d", x*100+y)
}

type requirement func(w hexInfo) bool

type uwpElementType int

const (
	St uwpElementType = iota
	Siz
	Atm
	Hyd
	Pop
	Gov
	Law
	TL = 8
)

func decodeHex(value string) int {
	decoded, err := strconv.ParseInt(value, 16, 32)

	if err != nil {
		panic(err)
	}

	return int(decoded)
}

func includesValue(allowed string, actual string) bool {
	for i := 0; i < len(allowed); i++ {
		next := string(allowed[i])

		if next == actual {
			return true
		}
	}
	return false
}

func is(elementType uwpElementType, allowed string) requirement {
	return func(w hexInfo) bool {
		var actual = string(w.uwp[elementType])
		return includesValue(allowed, actual)
	}
}

type definition struct {
	code    string
	require []requirement
}

func define(code string, req ...requirement) definition {
	return definition{
		code:    code,
		require: req,
	}
}

func starport(allowed string) requirement {
	return func(w hexInfo) bool {
		actual := string(w.uwp[St])
		for i := 0; i < len(allowed); i++ {
			next := string(allowed[i])
			if next == actual {
				return true
			}
		}
		return false
	}
}

var tradeCodes = []definition{
	// Planetary
	//
	define("As", is(Siz, "0"), is(Atm, "0"), is(Hyd, "0")),
	define("De", is(Atm, "23456789"), is(Hyd, "0")),
	define("Fl", is(Atm, "ABC"), is(Hyd, "123456789A")),
	define("Ga", is(Siz, "678"), is(Atm, "568"), is(Hyd, "567")),
	define("He", is(Siz, "3456789ABC"), is(Atm, "123")),
	define("Ic", is(Atm, "01"), is(Hyd, "123456789A")),
	define("Oc", is(Siz, "ABCDEF"), is(Atm, "3456789DEF"), is(Hyd, "A")),
	define("Va", is(Atm, "0")),
	define("Wa", is(Siz, "3456789"), is(Atm, "3456789DEF"), is(Hyd, "A")),
	// Population
	//
	define("Di", is(Pop, "0"), is(Gov, "0"), is(Law, "0"), is(TL, "123456789ABCDEF")),
	define("Ba", is(Pop, "0"), is(Gov, "0"), is(Law, "0"), is(St, "EX")),
	define("Lo", is(Pop, "123")),
	define("Ni", is(Pop, "456")),
	define("Ph", is(Pop, "8")),
	define("Hi", is(Pop, "9ABCDEF")),
	// Economic
	//
	define("Pa", is(Atm, "456789"), is(Hyd, "45678"), is(Pop, "48")),
	define("Ag", is(Atm, "456789"), is(Hyd, "45678"), is(Pop, "567")),
	define("Na", is(Atm, "0123"), is(Hyd, "0123"), is(Pop, "6789ABCDEF")),
	define("Px", is(Atm, "23AB"), is(Hyd, "12345"), is(Pop, "3456"), is(Law, "6789")),
	define("Pi", is(Atm, "012479"), is(Pop, "78")),
	define("In", is(Atm, "012479ABC"), is(Pop, "9ABCDEF")),
	define("Po", is(Atm, "2345"), is(Hyd, "0123")),
	define("Pr", is(Atm, "68"), is(Pop, "59")),
	define("Ri", is(Atm, "68"), is(Pop, "678")),
	// Climate
	//

}

func getBases(starport starportClass) string {
	var naval, scout bool

	switch starport {
	case "A":
		naval = dice(2) < 7
		scout = dice(2) < 5
	case "B":
		naval = dice(2) < 6
		scout = dice(2) < 6
	case "C":
		scout = dice(2) < 7
	case "D":
		scout = dice(2) < 8
	}

	result := ""

	if naval {
		result += "N"
	}

	if scout {
		result += "S"
	}

	return result
}

func getStarportQuality(population populationType) starportClass {
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

func getAtmosphere(size worldSize) atmosphereType {
	roll := dice(2) - 7 + int(size)

	return atmosphereType(applyMinimum(roll, 0))
}

func getSurfaceTemp(atmosphere atmosphereType) int {
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

func getHydrographics(size worldSize, atmosphere atmosphereType, temperature int) hydrographicType {
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

	return hydrographicType(applyMinimum(roll, 0))
}

func getGovernment(population populationType) governmentType {
	var result int

	if population > 0 {
		result = dice(2) - 7 + int(population)
	}

	return governmentType(applyRange(result, 0, int(government_D)))
}

func getLawLevel(population populationType, government governmentType) lawLevel {
	var result int

	if population > 0 {
		result = dice(2) - 7 + int(government)
	}

	return lawLevel(applyRange(result, 0, 9))
}

func getTechLevel(starport starportClass, size worldSize, atmosphere atmosphereType, hydrographics hydrographicType, population populationType, government governmentType) techLevel {
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

	environmentalMin := 0
	switch atmosphere {
	case 0, 1:
		environmentalMin = 8
	case 2, 3:
		environmentalMin = 5
	case 4, 7, 9:
		environmentalMin = 3
	case 10:
		environmentalMin = 8
	case 11:
		environmentalMin = 9
	case 12:
		environmentalMin = 10
	case 13, 14:
		environmentalMin = 5
	case 15:
		environmentalMin = 8
	}

	return techLevel(applyRange(result, environmentalMin, 15))
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

func applyRange(value int, floor int, ceiling int) int {
	return applyMinimum(applyMaximum(value, ceiling), floor)
}

func assertRange(value int, floor int, ceiling int) int {
	if value > ceiling || value < floor {
		log.Fatalf("oops! %d is outside bounds of range %d - %d", value, floor, ceiling)
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

func flux() int {
	return dice(1) - dice(1)
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
