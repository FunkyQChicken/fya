package main


import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/list"
  tea "github.com/charmbracelet/bubbletea"
  "fmt"
	"github.com/charmbracelet/lipgloss"
)

var (
  winStyle = lipgloss.NewStyle().
    Width(40).
    Margin(1, 1).
    Padding(1, 1).
    BorderStyle(lipgloss.RoundedBorder())

  bodyStyle = lipgloss.NewStyle()

  inputStyle = lipgloss.NewStyle().
    Padding(0, 2).
    MarginBottom(1).
    BorderStyle(lipgloss.RoundedBorder())

  titleStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#9AFAFA")).
    Align(lipgloss.Center).
    Padding(0, 2).
    MarginBottom(1).
    BorderStyle(lipgloss.RoundedBorder())

      
)


type app struct {
  header string 
  child tea.Model
}
var header = "FYA; The better menu"
func initApp() app { 
  return app { 
    child: initChainPicker(), 
    //header: "FYA; The better menu",
  } 
}

func (a app) Init() tea.Cmd { return a.child.Init() }
func (a app) View() string { 
  return winStyle.Render(
    lipgloss.JoinVertical(
      0.0,
      titleStyle.Render(header),
      bodyStyle.Render(a.child.View()),
    ))
}
func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msgType := msg.(type) {
    case tea.KeyMsg:
      switch msgType.String() {
        case "ctrl+c":
          return a, tea.Quit
      }
    case tea.WindowSizeMsg:
      h, v := winStyle.GetFrameSize()
      msgType.Height -=  v // to account for margins
      msgType.Width -= h
      
      msgType.Height -= (1 + v) // to account for header

      msg = msgType
  }
  child, cmd := a.child.Update(msg)
  a.child = child
  return a, cmd
}



type locationItem struct { c location }
func (i locationItem) Title() string       { return i.c.GetDescription() }
func (i locationItem) Description() string { return i.c.GetAddress() }
func (i locationItem) FilterValue() string { return i.Title() + " " + i.Description() }

type locationPicker struct {
  list list.Model
}



type chainItem struct { c chain }
func (i chainItem) Title() string       { return i.c.GetName() }
func (i chainItem) Description() string { return "test" }
func (i chainItem) FilterValue() string { return i.Title() }

type picker struct {
  list list.Model
  chain chain
}



func initChainPicker() picker {
  chainItems := make([]list.Item, len(Chains))
  for i, v := range Chains {
    chainItems[i] = chainItem{v}
  }
  return picker {list.New(chainItems, list.NewDefaultDelegate(), 0, 0), nil}
}

func (p * picker) intoLocPicker() {
  locations := p.chain.Locations()
  locationItems := make([]list.Item, len(locations))
  for i, v := range locations {
    locationItems[i] = locationItem{v}
  }
  p.list.SetItems(locationItems)
  p.list.ResetFilter()
  p.list.ResetSelected()
}

func (p picker) Init() tea.Cmd { return nil }
func (p picker) View() string { return p.list.View() }
func (p picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case "enter":
          if p.chain == nil {
            ch, ok  := p.list.SelectedItem().(chainItem)
            if ok {
              p.chain = ch.c
              p.intoLocPicker()
              if p.chain.LoadCredentials() {
                return p, nil
              } else {
                sI := InitSignIn(p)
                cmd := sI.Init()
                return sI, cmd
              }
            }
          }
          return p, tea.Quit
      }
    case tea.WindowSizeMsg:
      p.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}



type SignIn struct {
  next tea.Model
  username textinput.Model
  password textinput.Model
  attempts int
}

func InitSignIn(next tea.Model) SignIn {
  username := textinput.New()
  username.Focus()
  username.Placeholder = "John Doe"
  username.Prompt = "> "

  password := textinput.New()
  password.Placeholder = "hunter2"
  password.EchoMode = textinput.EchoPassword
  password.Prompt = "> "

  return SignIn {
    next: next,
    username: username,
    password: password,
    attempts: 0,
  }
}

func (s SignIn) Init() tea.Cmd { header = "Please Sign In"; return textinput.Blink }

func (s SignIn) View() string {
  return lipgloss.JoinVertical(0, 
    inputStyle.Render(
      fmt.Sprintf("Username:\n%s", s.username.View()),
    ),
    inputStyle.Render(
      fmt.Sprintf("Password:\n%s", s.password.View()),
    ),
  )
}

func (s SignIn) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  if s.username.Focused() {
    s.username.Update(msg)
  } else {
    s.password.Update(msg)
  }
  return s, nil
}

