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
    Width(60).
    Margin(1, 1).
    Padding(1, 1).
    BorderStyle(lipgloss.RoundedBorder())

  bodyStyle = lipgloss.NewStyle()

  inputStyle = lipgloss.NewStyle().
    Padding(0, 2).
    MarginBottom(1)

  titleStyle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#9AFAFA")).
    Align(lipgloss.Center).
    Padding(0, 2).
    MarginBottom(1).
    BorderStyle(lipgloss.RoundedBorder())

      
)

func centsAsDollar(cents int) string {
  dollars := cents / 100
  cents    = cents % 100
  var padding string
  if cents < 10 {
    padding = "0" 
  } else {
    padding = ""
  }
  return fmt.Sprintf("%d.%s%d", dollars, padding, cents)
}

type app struct {
  header string 
  child tea.Model
}
var header = "FYA; Order with style"
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
      msgType.Width = winStyle.GetWidth()
      msgType.Width -= h
      
      msgType.Height -= (1 + v) // to account for header

      msg = msgType
  }
  child, cmd := a.child.Update(msg)
  a.child = child
  return a, cmd
}



type locationItem struct { c location }
func (i locationItem) Title() string       { return i.c.GetAddress() }
func (i locationItem) Description() string { return i.c.GetDescription() }
func (i locationItem) FilterValue() string { return i.Title() + " " + i.Description() }

type locationPicker struct {
  list list.Model
}



type chainItem struct { c chain }
func (i chainItem) Title() string       { return i.c.GetName() }
func (i chainItem) Description() string { return "" }
func (i chainItem) FilterValue() string { return i.Title() }



type menuItem struct { i item }
func (i menuItem) Title() string { 
  return fmt.Sprintf("%s - $%s", i.i.name, centsAsDollar(i.i.cost)) 
}
func (i menuItem) Description() string { 
  return fmt.Sprintf("%s\n%d calories", i.i.description, i.i.calories)
}
func (i menuItem) FilterValue() string { return i.i.name }



type picker struct {
  list list.Model
  chain chain
  location location
}

func initChainPicker() picker {
  chainItems := make([]list.Item, len(Chains))
  for i, v := range Chains {
    chainItems[i] = chainItem{v}
  }
  lst := list.New(chainItems, list.NewDefaultDelegate(), 0, 0)
  lst.Title = "Pick your restraunt chain"
  return picker {lst, nil, nil}
}

func (p * picker) intoLocPicker() {
  locations := p.chain.Locations()
  locationItems := make([]list.Item, len(locations))
  for i, v := range locations {
    locationItems[i] = locationItem{v}
  }
  p.list.Title = "Pick your restraunt location"
  p.list.SetItems(locationItems)
  p.list.ResetFilter()
  p.list.ResetSelected()
}

func (p * picker) intoFoodPicker() {
  p.location.CreateCart()
  menu := p.location.Menu()
  menuItems := make([]list.Item, len(menu))
  for i, v := range menu {
    menuItems[i] = menuItem{v}
  }
  p.list.Title = "Pick your items"
  p.list.SetItems(menuItems)
  p.list.ResetFilter()
  p.list.ResetSelected()
  list.NewDefaultDelegate()
}


func (p picker) Init() tea.Cmd { return nil }
func (p picker) View() string { return p.list.View() }
func (p picker) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case "enter", "space":
          if p.list.FilterState() == list.Filtering {
            break
          }
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
          } else if p.location == nil {
            loc, ok := p.list.SelectedItem().(locationItem)
            if ok {
              p.location = loc.c
              p.intoFoodPicker()
              return p, nil
            }
          } else {
            men, _ := p.list.SelectedItem().(menuItem)
            p.location.AddItem(men.i)
            // Todo go to discount prompt
            return p, nil
          }
      }
    case tea.WindowSizeMsg:
      p.list.SetSize(msg.Width, msg.Height)
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}



type SignIn struct {
  next picker
  username textinput.Model
  password textinput.Model
  attempts int
}

func InitSignIn(next picker) SignIn {
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

func (s SignIn) Init() tea.Cmd {return textinput.Blink }

func (s SignIn) View() string {
  return lipgloss.JoinVertical(0, 
    "Credentials not found, please sign in\n",
    inputStyle.Render(
      fmt.Sprintf("Username:\n%s", s.username.View()),
    ),
    inputStyle.Render(
      fmt.Sprintf("Password:\n%s", s.password.View()),
    ),
  )
}

func (s * SignIn) toggleFocus() tea.Cmd {
  if s.username.Focused() {
    s.username.Blur()
    return s.password.Focus()
  } else {
    s.password.Blur()
    return s.username.Focus()
  }
}

func (s SignIn) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  var foc tea.Cmd = nil
  switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case "tab", "shift+tab":
          foc = s.toggleFocus()
        case "enter":
          if s.username.Focused() {
            foc = s.toggleFocus()
          } else {
            if (s.next.chain.Login(s.username.Value(), s.password.Value())) {
              return s.next, nil
            } else {
              foc = s.toggleFocus()
            }
          }
      }
    case tea.WindowSizeMsg:
      width := msg.Width - inputStyle.GetHorizontalBorderSize() 
      s.username.Width = width - 3
      s.password.Width = width - 3
  }

  var pcmd tea.Cmd
  var ucmd tea.Cmd
  s.username, ucmd = s.username.Update(msg)
  s.password, pcmd = s.password.Update(msg)
  return s, tea.Batch(ucmd, pcmd, foc)
}

