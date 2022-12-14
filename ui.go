package main

import (

	. "github.com/FunkyQChicken/fya/restaurant"

	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
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
    Bold(true).
    Render

  highlight = lipgloss.NewStyle().
    Foreground(lipgloss.Color("#EE6FF8")).
    Render

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
var header = "Order with style"

type stopSpinningMsg struct {msg tea.Msg}
type startSpinningMsg struct {message string; callback tea.Cmd}

func spinWhile(message string, f func() tea.Msg) tea.Cmd {
  return func() tea.Msg { 
    return startSpinningMsg {
      message: message,
      callback: func() tea.Msg { 
        return stopSpinningMsg{f()}
      },
    }
  }
}

type app struct {
  child tea.Model
  waitingOnIo bool
  spinner spinner.Model
  waitingMessage string
}

func initApp() app { 
  ret := app { 
    child: initChainPicker(), 
    waitingOnIo: false,
    spinner: spinner.New(),
    waitingMessage: "",
  } 
  ret.spinner.Spinner = spinner.Dot
  return ret
}

func (a app) Init() tea.Cmd { 
  return tea.Batch(a.child.Init(), spinner.Tick)
}

func (a app) View() string { 
  spin := "\n"
  if a.waitingOnIo {
    spin = fmt.Sprintf("   %s %s...\n", 
      a.spinner.View(), 
      a.waitingMessage)
  }
  return winStyle.Render(
    lipgloss.JoinVertical(
      0.0,
      titleStyle.Render(lipgloss.JoinHorizontal(lipgloss.Center, logo, header)),
      spin,
      bodyStyle.Render(a.child.View()),
    ))
}

func (a app) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  var spinCmd tea.Cmd = nil

  switch castMsg := msg.(type) {
    case tea.KeyMsg:
      switch castMsg.String() {
        case "ctrl+c":
          return a, tea.Quit
      }
    case tea.WindowSizeMsg:
      h, v := winStyle.GetFrameSize()
      castMsg.Height -=  v // to account for margins
      castMsg.Width = winStyle.GetWidth()
      castMsg.Width -= h
      
      castMsg.Height -= (7 + v) // to account for header
      castMsg.Height -= 1 // to account for spinnner

      msg = castMsg
    case startSpinningMsg:
      a.waitingMessage = castMsg.message
      a.waitingOnIo = true
      return a, castMsg.callback

    case stopSpinningMsg:
      msg = castMsg.msg
      a.waitingOnIo = false

    default:
      a.spinner, spinCmd = a.spinner.Update(msg)
  }
  var cmd tea.Cmd = nil
  if !a.waitingOnIo {
    child, c := a.child.Update(msg)
    cmd = c
    a.child = child
  }
  return a, tea.Batch(cmd, spinCmd)
}



type locationItem struct { c Location }
func (i locationItem) Title() string       { return i.c.GetAddress() }
func (i locationItem) Description() string { return i.c.GetDescription() }
func (i locationItem) FilterValue() string { return i.Title() + " " + i.Description() }

type locationPicker struct {
  list list.Model
}



type chainItem struct { c Chain }
func (i chainItem) Title() string       { return i.c.GetName() }
func (i chainItem) Description() string { return "" }
func (i chainItem) FilterValue() string { return i.Title() }



type menuItem struct { i FoodItem }
func (i menuItem) Title() string { 
  return fmt.Sprintf("%s - %s", i.i.Name, centsAsDollar(i.i.Cost)) 
}
func (i menuItem) Description() string { 
  return fmt.Sprintf("%s\n%d calories", i.i.Description, i.i.Calories)
}
func (i menuItem) FilterValue() string { return i.i.Name }


type discountItem struct { d Discount }
func (i discountItem) Title() string       { return i.d.Name }
func (i discountItem) Description() string { return i.d.Description }
func (i discountItem) FilterValue() string { return i.Title() }



type picker struct {
  list list.Model
  chain Chain
  location Location
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

func (p * picker) intoFoodPicker(menu []FoodItem) {
  menuItems := make([]list.Item, len(menu))
  for i, v := range menu {
    menuItems[i] = menuItem{v}
  }
  p.list.Title = "Pick your items"
  p.list.SetItems(menuItems)
  p.list.ResetFilter()
  p.list.ResetSelected()
}

func (p * picker) intoDiscountPicker(discounts []Discount) {
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
            loc, _ := p.list.SelectedItem().(locationItem)
            p.location = loc.c
            return p, spinWhile("getting menu", func() tea.Msg {
              p.location.CreateCart()
              return p.location.Menu()
            })
          } else if !p.foodChosen {
            men, ok := p.list.SelectedItem().(menuItem)
            if ok {
              discounts := p.location.Discounts()
              var next tea.Model
              p.foodChosen = true
              if len(discounts) > 0 {
                p.intoDiscountPicker(discounts)
                next = p
              } else {
                next = initCartPreview(p.location, p)
              }
              item := men.i
              return initCustomize(item, p, next), nil
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
    case []FoodItem:
      p.intoFoodPicker(msg)
	}

	var cmd tea.Cmd
	p.list, cmd = p.list.Update(msg)
	return p, cmd
}




type SignInField struct {
  fieldname string
  input textinput.Model
}

func (s SignInField) View() string {
    return inputStyle.Render(fmt.Sprintf("  %s:\n%s", s.fieldname, s.input.View()))
}

func initSignInField(fieldname string, placeholder string) SignInField {
  input := textinput.New()
  input.Placeholder = placeholder
  input.Prompt = "> "
  return SignInField{fieldname, input}
}

type SignIn struct {
  fields []SignInField
  currField int
  next picker
  attempts int
}


func InitSignIn(next picker) SignIn {
  requiredFields := next.chain.LoginFields()
  fields := make([]SignInField, 0, len(requiredFields))
  
  for fieldname, placeholder := range requiredFields {
    fields = append(fields, initSignInField(fieldname, placeholder))
  }
  currField := 0
  fields[currField].input.Focus()

  return SignIn {
    fields: fields,
    next: next,
    currField: currField,
    attempts: 0,
  }
}

func (s SignIn) Init() tea.Cmd {return textinput.Blink }

func (s SignIn) View() string {
  ret := fauxBlue.Render("Credentials not found, please provide auth token")+ "\n"
  for _, field := range s.fields {
    ret = lipgloss.JoinVertical(0,
      ret,
      field.View())
  }

  return ret
}

func (s * SignIn) moveFocus(offset int) tea.Cmd {
  nextField := (s.currField + offset + len(s.fields)) % len(s.fields)
  s.fields[s.currField].input.Blur()
  s.currField = nextField
  return s.fields[nextField].input.Focus()
}

func (s * SignIn) tryLogin() bool {
  loginDict := make(map[string] string, len(s.fields))
  for _, field := range s.fields {
    loginDict[field.fieldname] = field.input.Value()
  }
  return s.next.chain.Login(loginDict)
}

func (s SignIn) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  var foc tea.Cmd
  switch msg := msg.(type) {
    case tea.KeyMsg:
      switch msg.String() {
        case "enter":
          if s.currField == len(s.fields) - 1 {
            if (s.tryLogin()) {
              s.next.intoLocPicker()
              return s.next, nil 
            }
          } 
          foc = s.moveFocus(1)
        case "tab":
          foc = s.moveFocus(1)
        case "shift+tab":
          foc = s.moveFocus(-1)
      }
    case tea.WindowSizeMsg:
      width := msg.Width - inputStyle.GetHorizontalBorderSize() 
      for i := range s.fields {
        s.fields[i].input.Width = width - 3
      }
  }

  cmds := make([]tea.Cmd, len(s.fields) + 1)
  for i := range s.fields {
    s.fields[i].input, cmds[i + 1] = s.fields[i].input.Update(msg)
  }
  cmds[0] = foc
  return s, tea.Batch(cmds...)
}



type cartPreview struct { 
  location Location 
  items []CartItem
  prev picker
}

func initCartPreview(location Location, prev picker) cartPreview{
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
    items += it.Description + "\n"
    totalCost += it.Cost
    costs += centsAsDollar(it.Cost) +"\n"
  }
  return smolPad.Render(fmt.Sprintf("%s\n\n%s\nTotal Cost: %s\n\n%s", 
    fauxBlue.Render("Would you like to place your order?"),
    lipgloss.JoinHorizontal(lipgloss.Top, items, right.Render(costs)),
    bold(centsAsDollar(totalCost)),
    subtle.Render("[Y]es/[N]o/[C]ancel")))
}

type customize struct {
  food FoodItem
  options []FoodOption
  prev picker
  next tea.Model
  curr int
}

func initCustomize(food FoodItem, prev picker, next tea.Model) tea.Model {
  return customize {
    food: food,
    options: prev.location.GetCustomizations(food),
    prev: prev,
    next: next,
    curr: 0,
  }
}

func (c customize) Init() tea.Cmd {
  return nil
}

func (c customize) View() string {
  views := make([]string, len(c.options))
  for i, opt := range c.options {
    views[i] = inputStyle.Render(ViewFoodOption(opt))
  }
  views[c.curr] = highlight(views[c.curr])
  return lipgloss.JoinVertical(0, views...)
}

func (c customize) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
  switch msg := msg.(type) {
    case tea.KeyMsg:
      switch  msg.String() {
        case "enter":
          c.prev.location.AddItem(c.food, c.options)
          return c.next, nil
        case "tab":
          c.curr = uwrap(c.curr, len(c.options))
        case "shift+tab":
          c.curr = dwrap(c.curr, len(c.options))
        default:
          c.options[c.curr] = UpdateFoodOption(c.options[c.curr], msg.String()) 
      }
  }
  return c, nil
}

func ViewFoodOption(o FoodOption) string {
  switch o := o.(type) {
    case FoodOptionSelectOne:
      return fmt.Sprintf("%s: < %s >", o.Name, o.Options[o.Curr])

    case  FoodOptionSelectNumber:
      return fmt.Sprintf("<%d> %s", o.Num, o.Name)

    case FoodOptionsGroup:
      ret := o.Name 
      for i, fo := range o.Options {
        curr := ViewFoodOption(fo)
        if i == o.Selected {
          curr = bold(curr)
        }
        ret = lipgloss.JoinVertical(0, ret, curr)
      }
      return ret
    default:
      return fmt.Sprintf("Error: %#v\n",0)
  }
}

func dwrap(curr int, max int) int {
  return (curr + max - 1) % max
}
func uwrap(curr int, max int) int {
  return (curr + 1) % max
}

func UpdateFoodOption(o FoodOption, key string) FoodOption {
  switch o := o.(type) {
    case FoodOptionSelectOne:
      switch key {
        case "l", "right": o.Curr = uwrap(o.Curr, len(o.Options))
        case "h", "left": o.Curr = dwrap(o.Curr, len(o.Options))
      }
      return o

    case  FoodOptionSelectNumber:
      switch key {
        // TODO: Doesn't account for min
        case "l", "left": o.Num = uwrap(o.Num, o.Max)
        case "h", "right": o.Num = dwrap(o.Num, o.Max)
      }
      return o

    case FoodOptionsGroup:
      switch key {
        case "k", "up": o.Selected = dwrap(o.Selected, len(o.Options))
        case "j", "down": o.Selected = uwrap(o.Selected, len(o.Options))
        default: o.Options[o.Selected] = UpdateFoodOption(o.Options[o.Selected], key)
      }
      return o
  }
  return o
}
