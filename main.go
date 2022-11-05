package main

import (
//	"github.com/charmbracelet/bubbles"
  "fmt"
  "os"
	tea "github.com/charmbracelet/bubbletea"
)


func main() {
  p := tea.NewProgram(initApp())
  if err := p.Start(); err != nil {
    fmt.Printf("Error: %v", err)
    os.Exit(1)
  }
}
