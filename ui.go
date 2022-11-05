package main


import (
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
  tea "github.com/charmbracelet/bubbletea"
  "fmt"
	"github.com/charmbracelet/lipgloss"
)
var docStyle = lipgloss.NewStyle().Margin(1, 2)



type app struct {
  child tea.Model
}

func initApp() app { return app { child: initChainPicker()  } }

func (a app) Init() tea.Cmd { return a.child.Init() }
func (a app) View() string { return a.child.View() }
func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case "ctrl+c":
          return a, tea.Quit
      }
  }
  child, cmd := a.child.Update(msg)
  a.child = child
  return a, cmd
}





type chainItem struct { c chain }
func (i chainItem) Title() string       { return i.c.GetName() }
func (i chainItem) Description() string { return "" }
func (i chainItem) FilterValue() string { return i.Title() }

type chainPicker struct {
  list list.Model
}

func initChainPicker() chainPicker {
  chainItems := make([]list.Item, len(Chains))
  for i, v := range Chains {
    chainItems[i] = chainItem{v}
  }
  return chainPicker {list.New(chainItems, list.NewDefaultDelegate(), 0, 0)}
}

func (p chainPicker) Init() tea.Cmd { return nil }
func (p chainPicker) View() string { return docStyle.Render(p.list.View()) }
func (p chainPicker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		p.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}



type prompt struct { textInput textinput.Model }

func initPrompt() prompt {
  ti := textinput.New()
  ti.Placeholder = "10"
  ti.Focus()
  ti.CharLimit = 20
  ti.Width = 20
  return prompt {
    textInput: ti,
  };
}

func (m prompt) Init() tea.Cmd { return textinput.Blink }

func (m prompt) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case "enter":
          return m, tea.Quit
      }
  }
  var cmd tea.Cmd
  m.textInput, cmd = m.textInput.Update(msg)
  return m, cmd
}

func (m prompt) View() string {
  return fmt.Sprintf("How long should I sleep for?\n%s", m.textInput.View())
}



