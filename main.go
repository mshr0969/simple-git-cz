package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type state int

const (
	choosePrefix state = iota
	enterMessage
	commitDone
)

var (
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	emojiMap          = map[string][]string{}
)

type model struct {
	choices      []string
	cursor       int
	selected     string
	message      textinput.Model // テキスト入力用のモデル
	quitting     bool
	currentState state
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Enter your commit message"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 40

	return model{
		choices: []string{
			"feat: A new feature",
			"fix: A bug fix",
			"docs: Documentation only changes",
			"style: Changes that do not affect the code meaning (white-space, formatting, etc.)",
			"refactor: A code change that neither fixes a bug nor adds a feature",
			"perf: A code change that improves performance",
			"test: Adding missing tests or correcting existing tests",
			"chore: Other changes that don't modify src or test files",
		},
		cursor:       0,
		message:      ti,
		currentState: choosePrefix,
	}
}

func (m model) Init() tea.Cmd {
	return textinput.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.currentState {
		case choosePrefix:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "up", "k":
				if m.cursor > 0 {
					m.cursor--
				}
			case "down", "j":
				if m.cursor < len(m.choices)-1 {
					m.cursor++
				}
			case "enter":
				prefix := strings.SplitN(m.choices[m.cursor], ":", 2)[0]
				m.selected = prefix + ": " + randomEmoji(prefix) + " "
				m.currentState = enterMessage
			}

		case enterMessage:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				m.commit(m.selected + m.message.Value())
				m.currentState = commitDone
				return m, tea.Quit
			}

			m.message, cmd = m.message.Update(msg)
		}
	}

	return m, cmd
}

func (m model) View() string {
	if m.quitting {
		return "Exiting...\n"
	}

	switch m.currentState {
	case choosePrefix:
		s := "Choose a commit message prefix:\n\n"
		for i, choice := range m.choices {
			cursor := " "
			line := itemStyle.Render(choice)

			if m.cursor == i {
				cursor = ">"
				line = selectedItemStyle.Render(choice)
			}

			s += fmt.Sprintf("%s %s\n", cursor, line)
		}
		return s
	case enterMessage:
		return fmt.Sprintf("Enter your commit message (starting with %s):\n\n%s%s", m.selected, m.selected, m.message.View())
	case commitDone:
		return "Commit complete!\n"
	}
	return ""
}

func (m *model) commit(commitMessage string) {
	cmd := exec.Command("git", "commit", "-m", commitMessage)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Println("Failed to commit:", err)
		os.Exit(1)
	}
}

func randomEmoji(prefix string) string {
	if emojis, ok := emojiMap[prefix]; ok {
		rand.Seed(time.Now().UnixNano())
		return emojis[rand.Intn(len(emojis))]
	}
	return ""
}

func loadEmojis(filename string) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Fatalf("Failed to read emoji file: %v", err)
	}

	err = json.Unmarshal(data, &emojiMap)
	if err != nil {
		log.Fatalf("Failed to parse emoji file: %v", err)
	}
}

func main() {
	emojiFile := os.Getenv("EMOJI_FILE")
	if emojiFile == "" {
		log.Fatalf("EMOJI_FILE is not set")
	}
	loadEmojis(emojiFile)

	m := initialModel()
	p := tea.NewProgram(m)

	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
