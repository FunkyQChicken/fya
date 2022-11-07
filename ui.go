package main


import (
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/list"
  tea "github.com/charmbracelet/bubbletea"
  "fmt"
	"github.com/charmbracelet/lipgloss"
)


var logo = "" +
" ▄▄▄▄▄▄▄ ▄▄   ▄▄ ▄▄▄▄▄▄▄   \n" +
"█       █  █ █  █       █  \n" +
"█    ▄▄▄█  █▄█  █   ▄   █  \n" +
"█   █▄▄▄█       █  █▄█  █  \n" +
"█    ▄▄▄█▄     ▄█       █  \n" +
"█   █     █   █ █   ▄   █  \n" +
"█▄▄▄█     █▄▄▄█ █▄▄█ █▄▄█  \n"

var (
  smolPad = lipgloss.NewStyle().
    Padding(0, 2)

  fauxBlue = lipgloss.NewStyle().
		Background(lipgloss.Color("62")).
		Foreground(lipgloss.Color("230"))

  winStyle = lipgloss.NewStyle().
    Width(60).
    Margin(1, 1).
    Padding(1, 1).
    BorderStyle(lipgloss.ThickBorder())

  bodyStyle = lipgloss.NewStyle()

  inputStyle = lipgloss.NewStyle().
    Padding(0, 2).
    MarginBottom(1)

  titleStyle = lipgloss.NewStyle().
    Bold(true).
    Foreground(lipgloss.Color("#9AFAFA")).
    Align(lipgloss.Center).
    Padding(0, 2).
    MarginBottom(1)
      
	subtle = lipgloss.NewStyle().
    Foreground(lipgloss.Color("241"))

  right = lipgloss.NewStyle().
    Align(lipgloss.Right).
    PaddingLeft(2)

  bold = lipgloss.NewStyle().
    Bold(true)
)

func centsAsDollar(cents int) string {
  negative := ""
  if cents < 0 {
    cents *= -1
    negative = "-"
  }
  dollars := cents / 100
  cents    = cents % 100
  var padding string
  if cents < 10 {
    padding = "0" 
  } else {
    padding = ""
  }
  return fmt.Sprintf("%s$%d.%s%d", negative, dollars, padding, cents)
}

type app struct {
  child tea.Model
}
var header = "Order with style"
func initApp() app { 
  return app { 
    child: initChainPicker(), 
  } 
}

func (a app) Init() tea.Cmd { return a.child.Init() }
func (a app) View() string { 
  return winStyle.Render(
    lipgloss.JoinVertical(
      0.0,
      titleStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, logo, header)),
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
      
      msgType.Height -= (7 + v) // to account for header

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
  return fmt.Sprintf("%s - %s", i.i.name, centsAsDollar(i.i.cost)) 
}
func (i menuItem) Description() string { 
  return fmt.Sprintf("%s\n%d calories", i.i.description, i.i.calories)
}
func (i menuItem) FilterValue() string { return i.i.name }


type discountItem struct { d discount }
func (i discountItem) Title() string       { return i.d.name }
func (i discountItem) Description() string { return i.d.description }
func (i discountItem) FilterValue() string { return i.Title() }



type picker struct {
  list list.Model
  chain chain
  location location
  foodChosen bool
}

func initChainPicker() picker {
  chainItems := make([]list.Item, len(Chains))
  for i, v := range Chains {
    chainItems[i] = chainItem{v}
  }
  lst := list.New(chainItems, list.NewDefaultDelegate(), 0, 0)
  lst.Title = "Pick your restaurant chain"
  return picker {lst, nil, nil, false}
}

func (p * picker) intoLocPicker() {
  locations := p.chain.Locations()
  locationItems := make([]list.Item, len(locations))
  for i, v := range locations {
    locationItems[i] = locationItem{v}
  }
  p.list.Title = "Pick your restaurant location"
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
}

func (p * picker) intoDiscountPicker(discounts []discount) {
  discountItems := make([]list.Item, len(discounts))
  for i, v := range discounts {
    discountItems[i] = discountItem{v}
  }
  p.list.Title = "Pick discounts"
  p.list.SetItems(discountItems)
  p.list.ResetFilter()
  p.list.ResetSelected()
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
              if p.chain.LoadCredentials() {
                p.intoLocPicker()
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
          } else if !p.foodChosen {
            men, ok := p.list.SelectedItem().(menuItem)
            if ok {
              p.location.AddItem(men.i)
              discounts := p.location.Discounts()
              p.foodChosen = true
              if len(discounts) > 0 {
                p.intoDiscountPicker(discounts)
                return p, nil
              } else {
                return initCartPreview(p.location, p), nil
              }
            }
          } else {
            disc, ok := p.list.SelectedItem().(discountItem)
            if ok {
              p.location.ApplyDiscounts(disc.d)
              return initCartPreview(p.location, p), nil
            }
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
  token textinput.Model
  attempts int
}

func InitSignIn(next picker) SignIn {
  token := textinput.New()
  token.Focus() 
  token.Placeholder = "1fd341-..."
  token.Prompt = "> "

  return SignIn {
    next: next,
    token: token,
    attempts: 0,
  }
}

func (s SignIn) Init() tea.Cmd {return textinput.Blink }

func (s SignIn) View() string {
  return lipgloss.JoinVertical(0, 
    "  " + fauxBlue.Render("Credentials not found, please provide auth token")+ "\n",
    inputStyle.Render(
      fmt.Sprintf("Auth Token:\n%s", s.token.View()),
    ),
  )
}

func (s SignIn) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case "enter":
          if s.next.chain.Login(s.token.Value()) {
            s.next.intoLocPicker()
            return s.next, nil
          }
      }
    case tea.WindowSizeMsg:
      width := msg.Width - inputStyle.GetHorizontalBorderSize() 
      s.token.Width = width - 3
  }

  var cmd tea.Cmd
  s.token, cmd = s.token.Update(msg)
  return s, cmd
}



type cartPreview struct { 
  location location 
  items []cartItem
  prev picker
}

func initCartPreview(location location, prev picker) cartPreview{
  return cartPreview {
    location: location,
    items: location.Cart(),
    prev: prev,
  }   
}

func (c cartPreview) Init() tea.Cmd { return nil }
func (c cartPreview) Update(msg tea.Msg) (tea.Model, tea.Cmd) { 
  switch msg := msg.(type) {
  case tea.KeyMsg:
    switch msg.String() {
      case "y", "Y":
        c.location.Checkout()
        header = "Order Placed!"
        return c, tea.Quit
      case "b", "B", "n", "N":
        c.prev.foodChosen = false
        c.prev.intoFoodPicker()
        return c.prev, nil
      case "c", "C", "q", "Q":
        header = "Order Canceled!"
        return c, tea.Quit
    }
  }
  return c, nil
}

func (c cartPreview) View() string {
  costs := ""
  items := ""
  totalCost := 0
  for _, it := range c.items {
    items += it.description + "\n"
    totalCost += it.cost
    costs += centsAsDollar(it.cost) +"\n"
  }
  return smolPad.Render(fmt.Sprintf("%s\n\n%s\nTotal Cost: %s\n\n%s", 
    fauxBlue.Render("Would you like to place your order?"),
    lipgloss.JoinHorizontal(lipgloss.Top, items, right.Render(costs)),
    bold.Render(centsAsDollar(totalCost)),
    subtle.Render("[Y]es/[N]o/[C]ancel")))
}
