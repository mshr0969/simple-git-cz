package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"strings"
	"time"

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
)

var emojiMap = map[string][]string{
	"feat":     {"✨", "🚀", "🎉"},
	"fix":      {"🐛", "🔧", "🚑️"},
	"docs":     {"📚", "✏️", "📝"},
	"style":    {"🎨", "💄", "🎯"},
	"refactor": {"♻️", "🛠️", "🔄"},
	"perf":     {"⚡", "🔥", "💨"},
	"test":     {"✅", "🧪", "📊"},
	"chore":    {"🧹", "📦", "🔒"},
}

type model struct {
	choices      []string
	cursor       int
	selected     string
	message      string
	quitting     bool
	currentState state
}

func (m model) Init() tea.Cmd {
	// No initial command, just return nil
	return nil
}

func initialModel() model {
	return model{
		choices: []string{
			"feat: A new feature",
			"fix: A bug fix",
			"docs: Documentation only changes",
			"style: Changes that do not affect the meaning of the code (white-space, formatting, missing semi-colons, etc)",
			"refactor: A code change that neither fixes a bug nor adds a feature",
			"perf: A code change that improves performance",
			"test: Adding missing tests or correcting existing tests",
			"chore: Other changes that don't modify src or test files",
		},
		cursor:       0,
		selected:     "",
		message:      "",
		currentState: choosePrefix,
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.commit(m.selected + m.message)
				m.currentState = commitDone
				return m, tea.Quit
			case "backspace":
				if len(m.message) > 0 {
					m.message = m.message[:len(m.message)-1]
				}
			default:
				if msg.Type == tea.KeyRunes {
					m.message += msg.String()
				}
			}
		}
	}

	return m, nil
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
		return fmt.Sprintf("Enter your commit message (starting with %s):\n\n%s%s", m.selected, m.selected, m.message)
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

func main() {
	m := initialModel()

	p := tea.NewProgram(m)

	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
