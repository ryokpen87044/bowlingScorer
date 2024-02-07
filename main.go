package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/filepicker"
	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/paginator"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/log"
)

var docStyle = lipgloss.NewStyle().Margin(1, 0)
var docColor = lipgloss.Color("#EE6FF8")
var docInactiveColor = lipgloss.Color("#626262")

type Model struct {
	Bowl Bowl

	logger     *log.Logger
	inputKeys  inputKeyMap
	selectKeys selectKeyMap
	keyHelp    help.Model
	scene      string
	modeSel    list.Model
	nameInput  textinput.Model
	data       string
	dataSel    filepicker.Model
	scoreInput textinput.Model
	scoreSel   paginator.Model
}
type Bowl struct {
	Name     string     `json:"name"`
	Pins     [21]string `json:"pins"`
	Scores   [11]int    `json:"scores"`
	MaxScore int        `json:"maxScore"`
	Times    int        `json:"times"`
	Archives []Archive  `json:"archives"`
}
type Archive struct {
	Time   string     `json:"time"`
	Pins   [21]string `json:"pins"`
	Scores [11]int    `json:"scores"`
}

type inputKeyMap struct {
	enter key.Binding
	quit  key.Binding
}
type selectKeyMap struct {
	enter key.Binding
	next  key.Binding
	prev  key.Binding
	quit  key.Binding
}

var inputKeys = inputKeyMap{
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "enter"),
	),
	quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
var upDownKeys = selectKeyMap{
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "enter"),
	),
	next: key.NewBinding(
		key.WithKeys("up", "k"),
		key.WithHelp("↑", "up"),
	),
	prev: key.NewBinding(
		key.WithKeys("down", "j"),
		key.WithHelp("↓", "down"),
	),
	quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}
var rightLeftKeys = selectKeyMap{
	enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("↵", "enter"),
	),
	next: key.NewBinding(
		key.WithKeys("left", "h"),
		key.WithHelp("←", "left"),
	),
	prev: key.NewBinding(
		key.WithKeys("right", "l"),
		key.WithHelp("→", "right"),
	),
	quit: key.NewBinding(
		key.WithKeys("q", "esc", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
}

func (k inputKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.enter, k.quit}
}
func (k selectKeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.next, k.prev, k.enter, k.quit}
}
func (k inputKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{}, {}}
}
func (k selectKeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{{}, {}}
}

type dish struct {
	state string
	desc  string
}

var menu = []list.Item{
	dish{state: "new user", desc: "Create new data."},
	dish{state: "existing user", desc: "Select saved data."},
}

func (d dish) Title() string       { return d.state }
func (d dish) Description() string { return d.desc }
func (d dish) FilterValue() string { return d.state }

func (m Model) Init() tea.Cmd {
	m.logger.Info("Launch the app.")
	return m.dataSel.Init()
}

func (m Model) strike(pin string) ([21]string, int) {
	if n, err := strconv.Atoi(pin); err == nil {
		if n < 10 {
			if n == 0 {
				m.Bowl.Pins[m.Bowl.Times] = "G"
			} else {
				m.Bowl.Pins[m.Bowl.Times] = pin
			}
			m.Bowl.Times += 1
		} else if n == 10 {
			m.Bowl.Pins[m.Bowl.Times] = "X"
			m.Bowl.Times += 2
		}
	} else {
		if regexp.MustCompile(`^[xX]$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "X"
			m.Bowl.Times += 2
		} else if regexp.MustCompile(`^[gG]$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "G"
			m.Bowl.Times += 1
		} else if regexp.MustCompile(`^-$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "G"
			m.Bowl.Times += 1
		} else if regexp.MustCompile(`^[fF]$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "F"
			m.Bowl.Times += 1
		}
	}
	return m.Bowl.Pins, m.Bowl.Times
}
func (m Model) spare(pin string) ([21]string, int) {
	if n, err := strconv.Atoi(pin); err == nil {
		if l, e := strconv.Atoi(m.Bowl.Pins[m.Bowl.Times-1]); e == nil {
			if n < 10-l {
				if n == 0 {
					m.Bowl.Pins[m.Bowl.Times] = "-"
				} else {
					m.Bowl.Pins[m.Bowl.Times] = pin
				}
				m.Bowl.Times += 1
			} else if n == 10-l {
				m.Bowl.Pins[m.Bowl.Times] = "/"
				m.Bowl.Times += 1
			}
		} else {
			if n < 10 {
				if n == 0 {
					m.Bowl.Pins[m.Bowl.Times] = "-"
				} else {
					m.Bowl.Pins[m.Bowl.Times] = pin
				}
				m.Bowl.Times += 1
			} else if n == 10 {
				m.Bowl.Pins[m.Bowl.Times] = "/"
				m.Bowl.Times += 1
			}
		}
	} else {
		if regexp.MustCompile(`^/$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "/"
			m.Bowl.Times += 1
		} else if regexp.MustCompile(`^[gG]$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "-"
			m.Bowl.Times += 1
		} else if regexp.MustCompile(`^-$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "-"
			m.Bowl.Times += 1
		} else if regexp.MustCompile(`^[fF]$`).MatchString(pin) {
			m.Bowl.Pins[m.Bowl.Times] = "F"
			m.Bowl.Times += 1
		}
	}
	return m.Bowl.Pins, m.Bowl.Times
}
func (m Model) score() [11]int {
	frameNumber := 0
	for i, pin := range m.Bowl.Pins {
		if i < 18 {
			if i%2 == 0 {
				frameNumber++
			}
			if p, err := strconv.Atoi(pin); err == nil {
				if i%2 != 0 {
					m.Bowl.Scores[frameNumber] = m.Bowl.Scores[frameNumber-1]
					if q, err := strconv.Atoi(m.Bowl.Pins[i-1]); err == nil {
						m.Bowl.Scores[frameNumber] += p + q
					} else {
						m.Bowl.Scores[frameNumber] += p
					}
				}
			} else {
				switch pin {
				case "X":
					if m.Bowl.Pins[i+3] != "yet" || m.Bowl.Pins[i+4] != "yet" {
						m.Bowl.Scores[frameNumber] = m.Bowl.Scores[frameNumber-1]
						m.Bowl.Scores[frameNumber] += 10
						if p, e := strconv.Atoi(m.Bowl.Pins[i+2]); e == nil {
							m.Bowl.Scores[frameNumber] += p
							if m.Bowl.Pins[i+3] == "/" {
								m.Bowl.Scores[frameNumber] += 10 - p
							} else {
								if p, e := strconv.Atoi(m.Bowl.Pins[i+3]); e == nil {
									m.Bowl.Scores[frameNumber] += p
								} else {
									m.Bowl.Scores[frameNumber] += 0
								}
							}
						} else if m.Bowl.Pins[i+2] == "X" {
							m.Bowl.Scores[frameNumber] += 10
							if frameNumber == 9 {
								if p, e := strconv.Atoi(m.Bowl.Pins[i+3]); e == nil {
									m.Bowl.Scores[frameNumber] += p
								} else if m.Bowl.Pins[i+3] == "X" {
									m.Bowl.Scores[frameNumber] += 10
								} else {
									m.Bowl.Scores[frameNumber] += 0
								}
							} else {
								if p, e := strconv.Atoi(m.Bowl.Pins[i+4]); e == nil {
									m.Bowl.Scores[frameNumber] += p
								} else if m.Bowl.Pins[i+4] == "X" {
									m.Bowl.Scores[frameNumber] += 10
								} else {
									m.Bowl.Scores[frameNumber] += 0
								}
							}
						} else {
							m.Bowl.Scores[frameNumber] += 0
							if m.Bowl.Pins[i+3] == "/" {
								m.Bowl.Scores[frameNumber] += 10
							} else {
								if p, e := strconv.Atoi(m.Bowl.Pins[i+3]); e == nil {
									m.Bowl.Scores[frameNumber] += p
								} else {
									m.Bowl.Scores[frameNumber] += 0
								}
							}
						}
					}
				case "/":
					if m.Bowl.Pins[i+1] != "yet" {
						m.Bowl.Scores[frameNumber] = m.Bowl.Scores[frameNumber-1]
						m.Bowl.Scores[frameNumber] += 10
						if p, e := strconv.Atoi(m.Bowl.Pins[i+1]); e == nil {
							m.Bowl.Scores[frameNumber] += p
						} else if m.Bowl.Pins[i+1] == "X" {
							m.Bowl.Scores[frameNumber] += 10
						} else {
							m.Bowl.Scores[frameNumber] += 0
						}
					}
				default:
					if regexp.MustCompile(`^[-GF]$`).MatchString(pin) {
						if i%2 != 0 {
							m.Bowl.Scores[frameNumber] = m.Bowl.Scores[frameNumber-1]
							if p, e := strconv.Atoi(m.Bowl.Pins[i-1]); e == nil {
								m.Bowl.Scores[frameNumber] += p
							} else {
								m.Bowl.Scores[frameNumber] += 0
							}
						}
					}
				}
			}
		} else {
			if i == 18 {
				frameNumber++
			}
			if m.Bowl.Times == 21 {
				m.Bowl.Scores[frameNumber] = m.Bowl.Scores[frameNumber-1]
				if p, err := strconv.Atoi(m.Bowl.Pins[18]); err == nil {
					if m.Bowl.Pins[19] != "/" {
						m.Bowl.Scores[frameNumber] += p
					}
				} else {
					switch m.Bowl.Pins[18] {
					case "X":
						m.Bowl.Scores[frameNumber] += 10
					default:
						m.Bowl.Scores[frameNumber] += 0
					}
				}
				if p, err := strconv.Atoi(m.Bowl.Pins[19]); err == nil {
					if m.Bowl.Pins[20] != "/" {
						m.Bowl.Scores[frameNumber] += p
					}
				} else {
					switch m.Bowl.Pins[19] {
					case "X":
						m.Bowl.Scores[frameNumber] += 10
					case "/":
						m.Bowl.Scores[frameNumber] += 10
					default:
						m.Bowl.Scores[frameNumber] += 0
					}
				}
				if p, err := strconv.Atoi(m.Bowl.Pins[20]); err == nil {
					m.Bowl.Scores[frameNumber] += p
				} else {
					switch m.Bowl.Pins[20] {
					case "X":
						m.Bowl.Scores[frameNumber] += 10
					case "/":
						m.Bowl.Scores[frameNumber] += 10
					default:
						m.Bowl.Scores[frameNumber] += 0
					}
				}
			}
		}
	}
	return m.Bowl.Scores
}
func (m Model) maxScore() int {
	if m.Bowl.Times < 19 {
		if m.Bowl.Times%2 != 0 {
			m.Bowl.Pins[m.Bowl.Times] = "/"
			m.Bowl.Times++
		}
		for i := m.Bowl.Times; i <= 20; i++ {
			if i > 17 {
				m.Bowl.Pins[i] = "X"
			} else if i%2 == 0 {
				m.Bowl.Pins[i] = "X"
			}
		}
	} else if m.Bowl.Times < 21 {
		for i := m.Bowl.Times; i <= 20; i++ {
			switch m.Bowl.Pins[i-1] {
			case "X":
				m.Bowl.Pins[i] = "X"
			case "/":
				m.Bowl.Pins[i] = "X"
			default:
				m.Bowl.Pins[i] = "/"
			}
		}
	}
	m.Bowl.Times = 21
	ms := m.score()
	return ms[10]
}
func (m Model) addScore(str string) Bowl {
	if m.Bowl.Times > 17 {
		times := m.Bowl.Times
		switch m.Bowl.Times {
		case 18:
			m.Bowl.Pins, m.Bowl.Times = m.strike(str)
			if m.Bowl.Pins[times] == "X" {
				m.Bowl.Times--
			}
		case 19:
			if m.Bowl.Pins[times-1] == "X" {
				m.Bowl.Pins, m.Bowl.Times = m.strike(str)
				if m.Bowl.Pins[times] == "X" {
					m.Bowl.Times--
				}
			} else {
				m.Bowl.Pins, m.Bowl.Times = m.spare(str)
				if m.Bowl.Pins[times] != "/" {
					m.Bowl.Times = 21
				}
			}
		case 20:
			if m.Bowl.Pins[times-1] == "X" || m.Bowl.Pins[times-1] == "/" {
				m.Bowl.Pins, m.Bowl.Times = m.strike(str)
			} else {
				m.Bowl.Pins, m.Bowl.Times = m.spare(str)
			}
			if m.Bowl.Pins[times] != "yet" {
				m.Bowl.Times = 21
			}
		}
	} else if m.Bowl.Times%2 == 0 {
		m.Bowl.Pins, m.Bowl.Times = m.strike(str)
	} else {
		m.Bowl.Pins, m.Bowl.Times = m.spare(str)
	}
	m.Bowl.Scores = m.score()
	m.Bowl.MaxScore = m.maxScore()
	return m.Bowl
}
func (m Model) nextGame() (Bowl, paginator.Model) {
	a := Archive{
		Time:   time.Now().Format("2006/01/02 15:04:05 -0700 MST"),
		Pins:   m.Bowl.Pins,
		Scores: m.Bowl.Scores,
	}
	m.Bowl.Archives = append(m.Bowl.Archives, a)

	m.Bowl.Pins = initPins()
	m.Bowl.Scores = initScores()
	m.Bowl.MaxScore = 300
	m.Bowl.Times = 0

	m.scoreSel.SetTotalPages(len(m.Bowl.Archives))
	n := m.scoreSel.TotalPages - m.scoreSel.Page
	for i := 0; i <= n; i++ {
		m.scoreSel.NextPage()
	}
	return m.Bowl, m.scoreSel
}

func (m Model) nameCheck() string {
	if m.nameInput.Value() == "" {
		return m.Bowl.Name
	} else {
		re := regexp.MustCompile(`[\\/:*?"<>|]`)
		if re.MatchString(m.nameInput.Value()) {
			m.logger.Error("Found an invalid value. Initialize to appropriate values.")
			return re.ReplaceAllString(m.nameInput.Value(), "-")
		}
		return m.nameInput.Value()
	}
}
func pinsCheck(pins [21]string) ([21]string, bool) {
	for i, pin := range pins {
		if p, err := strconv.Atoi(pin); err == nil {
			if p > 9 {
				return initPins(), false
			}
		} else {
			if i > 18 {
				switch i {
				case 19:
					if pins[i-1] == "X" {
						switch pin {
						case "X":
							continue
						case "G":
							continue
						case "F":
							continue
						case "yet":
							continue
						default:
							return initPins(), false
						}
					} else {
						switch pin {
						case "/":
							continue
						case "-":
							continue
						case "F":
							continue
						case "yet":
							continue
						default:
							return initPins(), false
						}
					}
				case 20:
					if pins[i-1] == "X" || pins[i-1] == "/" {
						switch pin {
						case "X":
							continue
						case "G":
							continue
						case "F":
							continue
						case "yet":
							continue
						default:
							return initPins(), false
						}
					} else {
						switch pin {
						case "/":
							continue
						case "-":
							continue
						case "F":
							continue
						case "yet":
							continue
						default:
							return initPins(), false
						}
					}
				}
			} else if i%2 == 0 {
				switch pin {
				case "X":
					continue
				case "G":
					continue
				case "F":
					continue
				case "yet":
					continue
				default:
					return initPins(), false
				}
			} else {
				switch pin {
				case "/":
					continue
				case "-":
					continue
				case "F":
					continue
				case "yet":
					continue
				default:
					return initPins(), false
				}
			}
		}
	}
	return pins, true
}
func scoresCheck(scores [11]int) ([11]int, bool) {
	lim := 0
	for _, score := range scores {
		if score > lim {
			return initScores(), false
		}
		lim += 30
	}
	return scores, true
}
func (m Model) dataCheck() Bowl {
	msg := "Found an invalid value. Initialize to appropriate values."
	if m.Bowl.Name == "" {
		m.logger.Error(msg)
		return initBowl()
	}
	if m.Bowl.MaxScore > 300 {
		m.logger.Error(msg)
		return initBowl()
	}
	if m.Bowl.Times > 21 {
		m.logger.Error(msg)
		return initBowl()
	}

	flg := true
	m.Bowl.Pins, flg = pinsCheck(m.Bowl.Pins)
	if !flg {
		m.Bowl.Scores = initScores()
		m.logger.Error(msg)
	}
	m.Bowl.Scores, flg = scoresCheck(m.Bowl.Scores)
	if !flg {
		m.Bowl.Pins = initPins()
		m.logger.Error(msg)
	}
	for i := range m.Bowl.Archives {
		m.Bowl.Archives[i].Pins, flg = pinsCheck(m.Bowl.Archives[i].Pins)
		if !flg {
			m.Bowl.Archives[i].Scores = initScores()
			m.logger.Error(msg)
		}
		m.Bowl.Archives[i].Scores, flg = scoresCheck(m.Bowl.Archives[i].Scores)
		if !flg {
			m.Bowl.Archives[i].Pins = initPins()
			m.logger.Error(msg)
		}
	}
	return m.Bowl
}
func (m Model) write() {
	if _, err := os.Stat("data"); err != nil {
		if e := os.Mkdir("data", 0777); e == nil {
			m.logger.Info("Create a directory named \"data\".")
		} else {
			m.logger.Fatal("Failed to create a directory named \"data\".")
		}
	}
	if file, err := os.Create(filepath.Join("data", fmt.Sprintf("%s.json", m.Bowl.Name))); err == nil {
		m.logger.Info(fmt.Sprintf("Create a JSON file named \"%s.json\".", m.Bowl.Name))
		encoder := json.NewEncoder(file)
		encoder.SetIndent("", "  ")
		if e := encoder.Encode(m.Bowl); e == nil {
			m.logger.Info("Encode data.")
		} else {
			m.logger.Fatal("Failed to encode data.")
		}
	} else {
		m.logger.Fatal(fmt.Sprintf("Failed to create a JSON file named \"%s.json\".", m.Bowl.Name))
	}
}
func (m Model) read() Bowl {
	fileName := filepath.Base(m.data)
	if file, err := os.Open(m.data); err == nil {
		m.logger.Info(fmt.Sprintf("Reading a JSON file named \"%s\".", fileName))
		if e := json.NewDecoder(file).Decode(&m.Bowl); e == nil {
			m.logger.Info("Decode data.")
		} else {
			m.logger.Fatal("Failed to decode data.")
		}
	} else {
		m.logger.Fatal(fmt.Sprintf("Failed to reading a JSON file named \"%s\".", fileName))
	}
	m.Bowl = m.dataCheck()
	return m.Bowl
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.dataSel, cmd = m.dataSel.Update(msg)
	switch m.scene {
	case "dataGenMode":
		m.nameInput, cmd = m.nameInput.Update(msg)
	case "dataSelMode":
		if didSelect, path := m.dataSel.DidSelectFile(msg); didSelect {
			m.data = path
		}
	case "mgmtScore":
		m.scoreInput, cmd = m.scoreInput.Update(msg)
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.scene {
		case "modeSelect":
			m.selectKeys = upDownKeys
			switch {
			case key.Matches(msg, m.selectKeys.enter):
				m.logger.Info("Current mode is \"Mode Selection\".")
				switch m.modeSel.Cursor() {
				case 0:
					m.logger.Info("\"Data Generation\" mode is selected.")
					m.scene = "dataGenMode"
				case 1:
					m.logger.Info("\"Data Selection\" mode is selected.")
					m.scene = "dataSelMode"
				}
			case key.Matches(msg, m.selectKeys.next):
				m.modeSel.CursorUp()
			case key.Matches(msg, m.selectKeys.prev):
				m.modeSel.CursorDown()
			case key.Matches(msg, m.selectKeys.quit):
				m.logger.Info("Close the app.")
				return m, tea.Quit
			}

		case "dataGenMode":
			switch {
			case key.Matches(msg, m.inputKeys.enter):
				m.logger.Info("Current mode is \"Data Generation\".")
				m.Bowl = initBowl()
				m.logger.Info(fmt.Sprintf("\"%s\" is typed.", m.nameInput.Value()))
				m.Bowl.Name = m.nameCheck()
				m.nameInput.Reset()
				m.scene = "mgmtScore"
			case key.Matches(msg, m.inputKeys.quit):
				m.logger.Info("Close the app.")
				return m, tea.Quit
			}

		case "dataSelMode":
			m.selectKeys = upDownKeys
			switch {
			case key.Matches(msg, m.selectKeys.enter):
				m.logger.Info("Current mode is \"Data Selection\".")
				m.Bowl = m.read()
				m.scoreSel.SetTotalPages(len(m.Bowl.Archives))
				m.scene = "mgmtScore"
			case key.Matches(msg, m.selectKeys.quit):
				m.logger.Info("Close the app.")
				return m, tea.Quit
			}

		case "mgmtScore":
			m.selectKeys = rightLeftKeys
			if m.Bowl.Times == 21 {
				m.logger.Info("Game start.")
				m.Bowl, m.scoreSel = m.nextGame()
			}
			switch {
			case key.Matches(msg, m.selectKeys.enter):
				m.logger.Info("Current mode is \"Management Score\".")
				m.logger.Info(fmt.Sprintf("\"%s\" is typed.", m.scoreInput.Value()))
				times := m.Bowl.Times
				m.Bowl = m.addScore(m.scoreInput.Value())
				if times != m.Bowl.Times {
					m.logger.Info("Update Score.")
				} else {
					m.logger.Warn("Invalid value. Type again.")
				}
				m.scoreInput.Reset()
				if m.Bowl.Times == 21 {
					m.logger.Info("Game over.")
					m.scoreInput.Placeholder = "Let's go to the next game!"
				} else {
					m.scoreInput.Placeholder = "How many pins were knocked down?"
				}
			case key.Matches(msg, m.selectKeys.next):
				m.scoreSel.PrevPage()
			case key.Matches(msg, m.selectKeys.prev):
				m.scoreSel.NextPage()
			case key.Matches(msg, m.selectKeys.quit):
				m.write()
				m.logger.Info("Close the app.")
				return m, tea.Quit
			}
		}
	}
	return m, cmd
}

func (m Model) scoreDrawing() string {
	scoreDrawing := strings.Builder{}
	scoreDrawing.WriteString("┏━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━┳━━━━━┓┏━━━━━┓\n")
	if m.Bowl.Times != 21 {
		scoreDrawing.WriteString("┃ 1 ┃ 2 ┃ 3 ┃ 4 ┃ 5 ┃ 6 ┃ 7 ┃ 8 ┃ 9 ┃ 10  ┃┃ MAX ┃\n")
	} else {
		scoreDrawing.WriteString("┃ 1 ┃ 2 ┃ 3 ┃ 4 ┃ 5 ┃ 6 ┃ 7 ┃ 8 ┃ 9 ┃ 10  ┃┃ RES ┃\n")
	}
	scoreDrawing.WriteString("┗━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━━━┛┗━━━━━┛\n")

	scoreDrawing.WriteString("┏━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┓┏━━━━━┓\n")
	pinsLine := "┃"
	for _, pin := range m.Bowl.Pins {
		pinStr := ""
		if pin == "yet" {
			pinStr = " "
		} else {
			pinStr = pin
		}
		pinsLine = fmt.Sprintf("%s%s┃", pinsLine, pinStr)
	}
	scoreDrawing.WriteString(fmt.Sprintf("%s┃     ┃\n", pinsLine))
	scoreDrawing.WriteString("┃ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━╋━┻━┻━┫┃")
	if m.Bowl.MaxScore < 10 {
		scoreDrawing.WriteString(fmt.Sprintf("  %d  ┃\n", m.Bowl.MaxScore))
	} else if m.Bowl.MaxScore < 100 {
		scoreDrawing.WriteString(fmt.Sprintf("  %d ┃\n", m.Bowl.MaxScore))
	} else {
		scoreDrawing.WriteString(fmt.Sprintf(" %d ┃\n", m.Bowl.MaxScore))
	}
	scoresLine := "┃"
	for i, score := range m.Bowl.Scores {
		if i == 0 {
			continue
		}
		scoreStr := ""
		if score == -1 {
			scoreStr = "   "
		} else {
			scoreStr = strconv.Itoa(score)
			if score < 100 {
				scoreStr = fmt.Sprintf(" %-02s", scoreStr)
			}
		}
		if i < 10 {
			scoresLine = fmt.Sprintf("%s%s┃", scoresLine, scoreStr)
		} else {
			scoresLine = fmt.Sprintf("%s %s ┃", scoresLine, scoreStr)
		}
	}
	scoreDrawing.WriteString(fmt.Sprintf("%s┃     ┃\n", scoresLine))
	scoreDrawing.WriteString("┗━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━━━┛┗━━━━━┛\n")

	archivesLen := len(m.Bowl.Archives)
	if archivesLen > 0 {
		high := 0
		low := 300
		sum := 0

		for _, archive := range m.Bowl.Archives {
			a := archive.Scores[10]
			if a > high {
				high = a
			}
			if a < low {
				low = a
			}
			sum += a
		}
		avg := sum / archivesLen
		scoreDrawing.WriteString(fmt.Sprintf(
			"    Game:%-02s  Total:%-04s  Avg:%-03s  H/G:%-03s  L/G:%-03s\n\n",
			strconv.Itoa(archivesLen+1),
			strconv.Itoa(sum),
			strconv.Itoa(avg),
			strconv.Itoa(high),
			strconv.Itoa(low),
		))
	} else {
		scoreDrawing.WriteString("    Game:1   Total:----  Avg:---  H/G:---  L/G:---\n\n")
	}
	return scoreDrawing.String()
}
func (m Model) archivesScoreDrawing() string {
	archiveScoresDrawing := strings.Builder{}
	start, end := m.scoreSel.GetSliceBounds(len(m.Bowl.Archives))
	for i, arc := range m.Bowl.Archives[start:end] {
		archiveScoresDrawing.WriteString(fmt.Sprintf(" Game %-06s[%s]\n", strconv.Itoa(start+i+1), arc.Time))
		archiveScoresDrawing.WriteString("┏━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┳━┓\n")
		archivePinsLine := "┃"
		for _, pin := range m.Bowl.Archives[i].Pins {
			archivePinStr := ""
			if pin == "yet" {
				archivePinStr = " "
			} else {
				archivePinStr = pin
			}
			archivePinsLine = fmt.Sprintf("%s%s┃", archivePinsLine, archivePinStr)
		}
		archiveScoresDrawing.WriteString(fmt.Sprintf("%s\n", archivePinsLine))
		archiveScoresDrawing.WriteString("┃ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━┫ ┗━╋━┻━┻━┫\n")

		archiveScoresLine := "┃"
		for i, score := range m.Bowl.Archives[i].Scores {
			if i == 0 {
				continue
			}
			archiveScoresStr := ""
			if score == -1 {
				archiveScoresStr = "   "
			} else {
				archiveScoresStr = strconv.Itoa(score)
				if score < 100 {
					archiveScoresStr = fmt.Sprintf(" %-02s", archiveScoresStr)
				}
			}
			if i < 10 {
				archiveScoresLine = fmt.Sprintf("%s%s┃", archiveScoresLine, archiveScoresStr)
			} else {
				archiveScoresLine = fmt.Sprintf("%s %s ┃", archiveScoresLine, archiveScoresStr)
			}

		}
		archiveScoresDrawing.WriteString(fmt.Sprintf("%s\n", archiveScoresLine))
		archiveScoresDrawing.WriteString("┗━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━┻━━━━━┛\n")
	}
	archiveScoresDrawing.WriteString(fmt.Sprintf("  %s\n", m.scoreSel.View()))
	return archiveScoresDrawing.String()
}

func (m Model) modeSelectScene() string {
	modeSelectScene := strings.Builder{}
	modeSelectScene.WriteString(fmt.Sprintf("%s\n", m.modeSel.View()))
	return modeSelectScene.String()
}
func (m Model) dataGenModeScene() string {
	dataGenModeScene := strings.Builder{}
	dataGenModeScene.WriteString(fmt.Sprintf("%s\n", m.modeSel.View()))
	dataGenModeScene.WriteString(fmt.Sprintf("%s\n\n", m.nameInput.View()))
	return dataGenModeScene.String()
}
func (m Model) dataSelModeScene() string {
	dataSelModeScene := strings.Builder{}
	dataSelModeScene.WriteString(fmt.Sprintf("%s\n", m.modeSel.View()))
	dataSelModeScene.WriteString(fmt.Sprintf("%s\n", m.dataSel.View()))
	return dataSelModeScene.String()
}
func (m Model) mgmtScoreScene() string {
	mgmtScoreScene := strings.Builder{}
	if len(m.Bowl.Archives) > 0 {
		mgmtScoreScene.WriteString(m.archivesScoreDrawing())
	}
	mgmtScoreScene.WriteString(m.scoreDrawing())
	mgmtScoreScene.WriteString(fmt.Sprintf("%s\n\n", m.scoreInput.View()))
	return mgmtScoreScene.String()
}
func (m Model) infoLine() (string, string) {
	name := fmt.Sprintf(" Player: %s\n\n", m.Bowl.Name)
	switch m.scene {
	case "modeSelect":
		m.selectKeys = upDownKeys
	case "dataGenMode":
		return name, lipgloss.PlaceHorizontal(40, 1, m.keyHelp.View(m.inputKeys))
	case "dataSelMode":
		m.selectKeys = upDownKeys
	case "mgmtScore":
		m.selectKeys = rightLeftKeys
	}
	return name, lipgloss.PlaceHorizontal(44, 1, m.keyHelp.View(m.selectKeys))
}

func (m Model) View() string {
	view := strings.Builder{}
	name, helper := m.infoLine()
	switch m.scene {
	case "modeSelect":
		view.WriteString(m.modeSelectScene())
	case "dataGenMode":
		view.WriteString(m.dataGenModeScene())
	case "dataSelMode":
		view.WriteString(m.dataSelModeScene())
	case "mgmtScore":
		view.WriteString(name)
		view.WriteString(m.mgmtScoreScene())
	}
	view.WriteString(helper)
	return docStyle.Render(view.String())
}

func initPins() [21]string {
	var pins [21]string
	for i := range pins {
		pins[i] = "yet"
	}
	return pins
}
func initScores() [11]int {
	var scores [11]int
	for i := range scores {
		if i == 0 {
			scores[i] = 0
		} else {
			scores[i] = -1
		}
	}
	return scores
}
func initBowl() Bowl {
	var archives []Archive
	return Bowl{
		Name:     time.Now().Format("20060102-150405MST"),
		Pins:     initPins(),
		Scores:   initScores(),
		MaxScore: 300,
		Times:    0,
		Archives: archives,
	}
}
func initLogger() *log.Logger {
	if _, err := os.Stat("logs"); err != nil {
		if e := os.Mkdir("logs", 0777); e != nil {
			fmt.Println("Error running program:", err)
			os.Exit(1)
		}
	}
	timer := time.Now().Format("20060102-150405MST")
	logfile, err := os.Create(filepath.Join("logs", fmt.Sprintf("%s.log", timer)))
	if err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
	logger := log.NewWithOptions(logfile, log.Options{
		ReportCaller:    true,
		ReportTimestamp: true,
		TimeFormat:      "2006/01/02 15:04:05 -0700 MST",
	})
	return logger
}
func initKeyHelp() help.Model {
	keyHelp := help.New()
	keyHelp.Styles.ShortKey = lipgloss.NewStyle().Foreground(docInactiveColor)
	keyHelp.Styles.ShortDesc = lipgloss.NewStyle().Foreground(docInactiveColor)
	keyHelp.Styles.ShortSeparator = lipgloss.NewStyle().Foreground(docInactiveColor)
	return keyHelp
}
func initModeSel() list.Model {
	modeSel := list.New(menu, list.NewDefaultDelegate(), 22, 6)
	modeSel.Title = "Mode selection"
	modeSel.SetShowTitle(false)
	modeSel.SetShowHelp(false)
	modeSel.SetShowStatusBar(false)
	modeSel.SetFilteringEnabled(false)
	modeSel.SetShowPagination(false)
	modeSel.Styles.FilterCursor = lipgloss.NewStyle().Foreground(docColor)
	return modeSel
}
func initNameInput() textinput.Model {
	nameInput := textinput.New()
	nameInput.CharLimit = 37
	nameInput.Placeholder = "What is your name?"
	nameInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(docInactiveColor)
	nameInput.Focus()
	return nameInput
}
func initDataSel() filepicker.Model {
	dataSel := filepicker.New()
	dataSel.AllowedTypes = []string{".json"}
	dir, _ := os.Getwd()
	if _, err := os.Stat("data"); err != nil {
		if e := os.Mkdir("data", 0777); e != nil {
			fmt.Println("Error running program:", e)
			os.Exit(1)
		}
	}
	dataSel.CurrentDirectory = filepath.Join(dir, "data")
	dataSel.Styles.Selected = lipgloss.NewStyle().Foreground(docColor)
	dataSel.Styles.EmptyDirectory =
		lipgloss.NewStyle().
			Foreground(docInactiveColor).
			PaddingLeft(3).
			SetString("No Files Found.\n")
	return dataSel
}
func initScoreInput() textinput.Model {
	scoreInput := textinput.New()
	scoreInput.CharLimit = 2
	scoreInput.Placeholder = "How many pins were knocked down?"
	scoreInput.PlaceholderStyle = lipgloss.NewStyle().Foreground(docInactiveColor)
	scoreInput.Focus()
	return scoreInput
}
func initScoreSel() paginator.Model {
	scoreSel := paginator.New()
	scoreSel.PerPage = 3
	scoreSel.Type = paginator.Dots
	scoreSel.ActiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#ffffff")).Render("•")
	scoreSel.InactiveDot = lipgloss.NewStyle().Foreground(lipgloss.Color("#555555")).Render("•")
	return scoreSel
}
func initModel() tea.Model {
	return Model{
		Bowl:       initBowl(),
		logger:     initLogger(),
		inputKeys:  inputKeys,
		selectKeys: upDownKeys,
		keyHelp:    initKeyHelp(),
		scene:      "modeSelect",
		modeSel:    initModeSel(),
		nameInput:  initNameInput(),
		dataSel:    initDataSel(),
		scoreInput: initScoreInput(),
		scoreSel:   initScoreSel(),
	}
}

func main() {
	if _, err := tea.NewProgram(initModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Println("Error running program:", err)
		os.Exit(1)
	}
}
