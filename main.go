package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// Different states of the application
type state int

const (
	choosePrefix state = iota
	enterMessage
	commitDone
)

type model struct {
	choices      []string
	cursor       int
	selected     string
	message      string
	quitting     bool
	currentState state
}

// Init is called when the program starts. It returns an initial command if any.
func (m model) Init() tea.Cmd {
	// No initial command, just return nil
	return nil
}

// initialModel creates the initial state of the model
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

// Update handles messages based on user input.
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
				// Store the selected prefix and switch to enterMessage state
				m.selected = strings.SplitN(m.choices[m.cursor], ": ", 2)[0] + ": "
				m.currentState = enterMessage
			}

		case enterMessage:
			switch msg.String() {
			case "ctrl+c", "q":
				m.quitting = true
				return m, tea.Quit
			case "enter":
				// Perform git commit and switch to commitDone state
				m.commit(m.selected + m.message)
				m.currentState = commitDone
				return m, tea.Quit
			case "backspace":
				// Remove the last character from the message
				if len(m.message) > 0 {
					m.message = m.message[:len(m.message)-1]
				}
			default:
				// Append typed characters to the message
				if msg.Type == tea.KeyRunes {
					m.message += msg.String()
				}
			}
		}
	}

	return m, nil
}

// View renders the UI.
func (m model) View() string {
	if m.quitting {
		return "Exiting...\n"
	}

	switch m.currentState {
	case choosePrefix:
		s := "Choose a commit message prefix:\n\n"
		for i, choice := range m.choices {
			cursor := " "
			if m.cursor == i {
				cursor = ">"
			}

			s += fmt.Sprintf("%s %s\n", cursor, choice)
		}
		return s
	case enterMessage:
		return fmt.Sprintf("Enter your commit message (starting with %s):\n\n%s%s", m.selected, m.selected, m.message)
	case commitDone:
		return "Commit complete!\n"
	}
	return ""
}

// commit runs git commit with the chosen prefix and message
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

func main() {
	m := initialModel()

	// Start the program
	p := tea.NewProgram(m)

	if err := p.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
