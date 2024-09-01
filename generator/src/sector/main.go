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
	population_B
	population_C
	population_D
	population_E
	population_F
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
	technology_1
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
	name                 string
	location             string
	hzVar                int
	uwp                  string
	bases                string
	remarks              string
	zone                 zoneType
	PBG                  string
	allegiance           string
	stars                string
	importance           int
	resources            int
	labor                int
	infrastructure       int
	efficiencies         int
	heterogeneity        int
	acceptance           int
	strangeness          int
	symbols              int
	nobility             []nobleTitle
	worlds               int
	primary              star
	populationMultiplier int
	belts                int
	gasGiants            int
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
	spectralType := spectralInfoMatrix[row][0]

	if spectralType == "OB" {
		if coin.Toss() {
			spectralType = "O"
		} else {
			spectralType = "B"
		}
	}

	return spectralType
}

var spectralInfoMatrix = [13][8]string{
	{"O", "Ia", "Ia", "Ia", "II", "II", "II", "II"},
	{"OB", "Ia", "Ia", "Ia", "II", "II", "II", "II"},
	{"A", "Ia", "Ia", "Ia", "II", "II", "II", "II"},
	{"A", "Ib", "Ib", "Ib", "III", "III", "III", "II"},
	{"F", "II", "II", "II", "IV", "IV", "IV", "II"},
	{"F", "III", "III", "III", "V", "V", "V", "III"},
	{"G", "III", "III", "V", "V", "V", "V", "V"},
	{"K", "V", "III", "V", "V", "V", "V", "V"},
	{"K", "V", "V", "V", "V", "V", "V", "V"},
	{"M", "V", "V", "V", "V", "V", "V", "V"},
	{"M", "IV", "IV", "V", "VI", "VI", "VI", "VI"},
	{"M", "D", "D", "D", "D", "D", "D", "D"},
	{"M", "D", "D", "D", "D", "D", "D", "D"},
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
	size := spectralInfoMatrix[row][col]

	switch {
	case size == "IV":
		if class.letter == "K" && class.numeral > 4 {
			return "V"
		}
	case size == "VI":
		if class.letter == "F" && class.numeral < 5 {
			return "V"
		}
	}
	return size
}

func getHzVar(star star) int {
	hzVar := 0
	dm := 0

	switch star.class.letter {
	case "M":
		dm += 2
	case "O", "B":
		dm -= 2
	}

	roll := flux() + dm

	switch {
	case roll < -5:
		hzVar = -2
	case roll > -6 && roll < -2:
		hzVar = -1
	case roll > 2 && roll < 6:
		hzVar = 1
	case roll > 5:
		hzVar = 2
	}

	return hzVar
}

type zoneType string

const (
	greenZone = "G"
	amberZone = "A"
	redZone   = "R"
)

func getWorlds(hex hexInfo) int {
	return 1 + hex.gasGiants + hex.belts + dice(2)
}

func getHeterogeneity(hex hexInfo) int {
	pop := getNumericUwpValue(hex.uwp, Pop)
	if pop == 0 {
		return 0
	}
	return applyRange(pop+flux(), 1, 0xF)
}

func getAcceptance(hex hexInfo) int {
	pop := getNumericUwpValue(hex.uwp, Pop)
	if pop == 0 {
		return 0
	}
	return applyRange(pop+hex.importance, 1, 0xF)
}

func getStrangeness(hex hexInfo) int {
	pop := getNumericUwpValue(hex.uwp, Pop)
	if pop == 0 {
		return 0
	}
	return applyMinimum(flux()+5, 1)
}

func getSymbols(hex hexInfo) int {
	pop := getNumericUwpValue(hex.uwp, Pop)
	if pop == 0 {
		return 0
	}
	return applyRange(flux()+getNumericUwpValue(hex.uwp, TL), 1, 0xF)
}

func getZone(hex hexInfo) zoneType {
	starport := string(hex.uwp[St])
	oppressionLevel := getNumericUwpValue(hex.uwp, Gov) + getNumericUwpValue(hex.uwp, Law)

	switch {
	case starport == "X":
		return redZone
	case oppressionLevel > 21:
		return redZone
	//case is(Pop, "0123456")(hex):
	//	return amberZone
	case oppressionLevel > 19:
		return amberZone
	}

	return greenZone
}

func getNumericUwpValue(uwp string, element uwpElementType) int {
	return decodeHex(string(uwp[element]))
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
		planets := newPlanets()

		var worlds []hexInfo

		for hx := 1; hx <= 32; hx++ {
			for hy := 1; hy <= 40; hy++ {
				locationCode := getLocationCode(hx, hy)

				if rollDecimal(1, 20) < 8 {
					population := getPopulation()
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
					var stars string
					if primary.size == "D" {
						if primary.class.letter == "B" {
							stars = "BD"
						} else {
							stars = "D"
						}
					} else {
						stars = fmt.Sprintf("%s%d %s", primary.class.letter, primary.class.numeral, primary.size)
					}
					hzVar := getHzVar(primary)

					hex := hexInfo{
						name:     planets.Name(),
						location: locationCode,
						uwp:      uwp,
						bases:    bases,
						primary:  primary,
						hzVar:    hzVar,
						stars:    stars,
					}

					hex.zone = getZone(hex)
					hex.remarks = getTradeCodes(hex)

					hex.populationMultiplier = getPopulationMultiplier(hex)
					hex.belts = applyMinimum(dice(1)-3, 0)
					hex.gasGiants = applyMinimum(dice(2)/2-2, 0)

					hex.allegiance = "Gc"
					hex.importance = getImportanceExtension(hex)

					hex.resources = getResources(hex)
					hex.labor = getLabor(hex)
					hex.infrastructure = getInfrastructure(hex)
					hex.efficiencies = getEfficiencies(hex)

					hex.heterogeneity = getHeterogeneity(hex)
					hex.acceptance = getAcceptance(hex)
					hex.strangeness = getStrangeness(hex)
					hex.symbols = getSymbols(hex)

					hex.nobility = getNobility(hex)

					hex.worlds = getWorlds(hex)

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

func getPopulationMultiplier(hex hexInfo) int {
	if getNumericUwpValue(hex.uwp, Pop) == 0 {
		return 0
	}
	return rollDecimal(1, 9)
}

type baseLetter string

const (
	navalBase      baseLetter = "N"
	navalDepot                = "D"
	scoutBase                 = "S"
	waystation                = "W"
	militaryBase              = "M"
	scientificBase            = "E"
	diplomaticBase            = "P"
	culturalBase              = "C"
)

func hasBase(required baseLetter) requirement {
	return func(w hexInfo) bool {
		for i := 0; i < len(w.bases); i++ {
			found := string(w.bases[i])
			if found == string(required) {
				return true
			}
		}
		return false
	}
}

type tradeCode string

const (
	asteroid  tradeCode = "As"
	desert              = "De"
	fluid               = "Fl"
	garden              = "Ga"
	hell                = "He"
	iceCapped           = "Ic"
	ocean               = "Oc"
	vacuum              = "Va"
	water               = "Wa"
	satellite           = "Sa"
	locked              = "Lk"

	dieback        = "Di"
	barren         = "Ba"
	lowPop         = "Lo"
	nonIndustrial  = "Ni"
	preHighPop     = "Ph"
	highPopulation = "Hi"

	preAg         = "Pa"
	agricultural  = "Ag"
	nonAg         = "Na"
	prisonExile   = "Px"
	preIndustrial = "Pi"
	industrial    = "In"
	poor          = "Po"
	preRich       = "Pr"
	rich          = "Ri"
	lowTech       = "Lt"
	highTech      = "Ht"

	frozen       = "Fr"
	hot          = "Ho"
	cold         = "Co"
	tropic       = "Tr"
	tundra       = "Tu"
	twilightZone = "Tz"

	farming      = "Fa"
	mining       = "Mi"
	militaryRule = "Mr"
	penalColony  = "Pe"
	reserve      = "Re"

	subsectorCapitol = "Cp"
	sectorCapitol    = "Cs"
	capitol          = "Cx"
	colony           = "Cy"

	forbidden      = "Fo"
	puzzle         = "Pz"
	dangerous      = "Da"
	dataRepository = "Ab"
	ancientSite    = "An"
)

func hasTradeCode(code tradeCode) requirement {
	return func(w hexInfo) bool {
		hasThese := strings.Split(w.remarks, " ")
		for i := 0; i < len(hasThese); i++ {
			next := tradeCode(hasThese[i])
			if next == code {
				return true
			}
		}
		return false
	}
}

func getImportanceExtension(hex hexInfo) int {
	value := 0

	switch {
	case is(St, "AB")(hex):
		value += 1
	case is(St, "DEX")(hex):
		value -= 1
	}

	tech := techLevel(getNumericUwpValue(hex.uwp, TL))
	if tech > technology_F {
		value += 1
	}

	if tech > technology_9 {
		value += 1
	}

	if tech < technology_9 {
		value -= 1
	}

	// trade codes
	for _, c := range []tradeCode{agricultural, highPopulation, industrial, rich} {
		if hasTradeCode(c)(hex) {
			value += 1
		}
	}

	// population
	if getNumericUwpValue(hex.uwp, Pop) < 7 {
		value -= 1
	}

	if hasBase("N")(hex) && hasBase("S")(hex) {
		value += 1
	}

	if hasBase("W")(hex) {
		value += 1
	}

	return value
}

func getResources(hex hexInfo) int {
	resources := dice(2)
	if getNumericUwpValue(hex.uwp, TL) > 7 {
		resources += hex.gasGiants + hex.belts
	}
	return applyRange(resources, 0, 0xF)
}

func getLabor(hex hexInfo) int {
	return applyMinimum(getNumericUwpValue(hex.uwp, Pop)-1, 0)
}

func getInfrastructure(hex hexInfo) int {
	var infrastructure int

	population := getNumericUwpValue(hex.uwp, Pop)

	switch population {
	case 0:
		infrastructure = 0
	case 1, 2, 3:
		infrastructure = hex.importance
	case 4, 5, 6:
		infrastructure = dice(1) + hex.importance
	default:
		infrastructure = dice(2) + hex.importance
	}

	return applyRange(infrastructure, 0, 0xF)
}

func getEfficiencies(_ hexInfo) int {
	value := flux()
	if value == 0 {
		return 1
	}
	return value
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

type nobleTitle string

const (
	knight    nobleTitle = "B"
	baronet              = "c"
	baron                = "C"
	marquis              = "D"
	viscount             = "e"
	count                = "E"
	duke                 = "f"
	grandDuke            = "F"
	archduke             = "G"
)

var any requirement = func(_ hexInfo) bool {
	return true
}

func or(r ...requirement) requirement {
	return func(w hexInfo) bool {
		for _, next := range r {
			if next(w) {
				return true
			}
		}
		return false
	}
}

func and(r ...requirement) requirement {
	return func(w hexInfo) bool {
		for _, next := range r {
			if !next(w) {
				return false
			}
		}
		return true
	}
}

func isImportant(min int) requirement {
	return func(w hexInfo) bool {
		return w.importance >= min
	}
}

func excludeTradeCodes(codes ...tradeCode) requirement {
	var requirements []requirement
	for _, next := range codes {
		hasIt := hasTradeCode(next)
		requirements = append(requirements, func(h hexInfo) bool {
			return !hasIt(h)
		})
	}

	return and(requirements...)
}

func includeTradeCodes(codes ...tradeCode) requirement {
	var requirements []requirement
	for _, next := range codes {
		hasIt := hasTradeCode(next)
		requirements = append(requirements, hasIt)
	}
	return or(requirements...)
}

type nobilityEntry struct {
	title    nobleTitle
	requires requirement
}

var nobilityRequirements = []nobilityEntry{
	{title: knight, requires: any},
	{title: baronet, requires: includeTradeCodes(preAg, preRich)},
	{title: baron, requires: includeTradeCodes(agricultural, rich)},
	{title: marquis, requires: hasTradeCode(preIndustrial)},
	{title: viscount, requires: hasTradeCode(preHighPop)},
	{title: count, requires: includeTradeCodes(industrial, highPopulation)},
	{title: duke, requires: and(isImportant(4), excludeTradeCodes(subsectorCapitol, capitol, sectorCapitol))},
	{title: duke, requires: includeTradeCodes(capitol, subsectorCapitol)},
}

func getNobility(hex hexInfo) []nobleTitle {
	var nobility []nobleTitle

	for _, next := range nobilityRequirements {
		if next.requires(hex) {
			nobility = append(nobility, next.title)
		}
	}

	return nobility
}

func trimBrackets(val string) string {
	return strings.Trim(val, "[]")
}

func formatZone(value zoneType) string {
	if value == greenZone {
		return ""
	}
	return string(value)
}

func saveSector(sector sector) tea.Cmd {
	return func() tea.Msg {
		f, err := os.Create(sector.name)
		check(err)
		defer f.Close()

		_, headerErr := f.WriteString("Hex\tName\tUWP\tBases\tRemarks\tZone\tPBG\tAllegiance\tStars\t{Ix}\t(Ex)\t[Cx]\tNobility\tW\n")
		check(headerErr)

		for _, h := range sector.hexes {
			_, worldErr := f.WriteString(
				fmt.Sprintf("%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\t%d\n",
					h.location,
					h.name,
					h.uwp,
					h.bases,
					h.remarks,
					formatZone(h.zone),
					fmt.Sprintf("%d%d%d", h.populationMultiplier, h.belts, h.gasGiants),
					h.allegiance,
					h.stars,
					fmt.Sprintf("{ %d }", h.importance),
					strings.ToUpper(fmt.Sprintf("(%x%x%x%+d)", h.resources, h.labor, h.infrastructure, h.efficiencies)),
					strings.ToUpper(fmt.Sprintf("[%x%x%x%x]", h.heterogeneity, h.acceptance, h.strangeness, h.symbols)),
					strings.Join(strings.Split(strings.Trim(fmt.Sprintf("%v", h.nobility), "[]"), " "), ""),
					h.worlds,
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

type requirement func(h hexInfo) bool

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
	code    tradeCode
	require []requirement
}

func define(code tradeCode, req ...requirement) definition {
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

func hasHzVariance(hzVar int) requirement {
	return func(w hexInfo) bool {
		return w.hzVar == hzVar
	}
}

var tradeCodes = []definition{
	// Planetary
	//
	define("As", is(Siz, "0"), is(Hyd, "0")),
	define("De", is(Atm, "23456789"), is(Hyd, "0")),
	define("Fl", is(Atm, "ABC"), is(Hyd, "123456789A")),
	define("Ga", is(Siz, "678"), is(Atm, "568"), is(Hyd, "567")),
	define("He", is(Siz, "3456789ABC"), is(Atm, "2479ABC"), is(Hyd, "012")),
	define("Ic", is(Atm, "01"), is(Hyd, "123456789A")),
	define("Oc", is(Siz, "ABCDEF"), is(Atm, "3456789DEF"), is(Hyd, "A")),
	define("Va", is(Atm, "0")),
	define("Wa", is(Siz, "3456789"), is(Atm, "3456789DEF"), is(Hyd, "A")),
	// Population
	//
	define("Di", is(Pop, "0"), is(Gov, "0"), is(Law, "0"), is(TL, "123456789ABCDEF")),
	define("Ba", is(Pop, "0"), is(Gov, "0"), is(Law, "0"), is(St, "EX"), is(TL, "0")),
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
	define("Ri", is(Atm, "68"), is(Pop, "678"), is(Gov, "456789")),
	define(lowTech, is(Pop, "123456789ABCDEF"), is(TL, "12345")),
	define(highTech, is(TL, "CDEF")),
	// Climate
	//
	define("Fr", is(Siz, "23456789"), is(Hyd, "123456789A"), hasHzVariance(2)),
	define("Ho", hasHzVariance(-1)),
	define("Co", hasHzVariance(1)),
	define("Tr", is(Siz, "6789"), is(Atm, "456789"), is(Hyd, "34567"), hasHzVariance(-1)),
	define("Tu", is(Siz, "6789"), is(Atm, "456789"), is(Hyd, "34567"), hasHzVariance(1)),
	// Secondary
	//
	define("Re", is(Pop, "01234"), is(Gov, "6"), is(Law, "045")),
	// Political
	//
	define("Cy", is(Pop, "01234"), is(Gov, "6"), is(Law, "0123")),
}

func getTradeCodes(hex hexInfo) string {
	var codes []tradeCode

	for _, def := range tradeCodes {
		match := true
		for _, req := range def.require {
			if !req(hex) {
				match = false
				break
			}
		}
		if match {
			codes = append(codes, def.code)
		}
	}

	if getNumericUwpValue(hex.uwp, Gov) == 6 {
		if !includeTradeCodes(militaryRule, prisonExile, reserve)(hex) {
			codes = append(codes, militaryRule)
		}
	}

	return strings.Trim(fmt.Sprintf("%v", codes), "[]")
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

func getPopulation() populationType {
	roll := dice(2) - 2

	if roll == 10 {
		return populationType(dice(2) + 3)
	}
	return populationType(roll)
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
	case 9:
		dm += 1
	case 0xA:
		dm += 2
	}

	switch population {
	case 1, 2, 3, 4, 5:
		dm += 1
	case 9:
		dm += 2
	case 0xA, 0xB, 0xC, 0xD, 0xE, 0xF:
		dm += 4
	}

	switch government {
	case 0, 5:
		dm += 1
	case 7:
		dm += 2
	case 0xD:
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
	current   string
}

func newPlanets() *planetnames {
	return &planetnames{}
}

var usedNames = map[string]int{}

func (p *planetnames) Name() string {
	p.Next()
	_, exists := usedNames[p.current]
	if exists {
		return p.Name()
	} else {
		usedNames[p.current] = 1
		return p.current
	}
}

func (p *planetnames) Next() {
	if p.remaining == 0 {
		p.names = getMorePlanetNames()
		p.remaining = len(p.names)
	}

	result := p.names[0]
	p.names = p.names[1:]
	p.remaining--
	p.current = result
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
