package main

import (
//	"github.com/charmbracelet/bubbles"
  "fmt"
  "os"
	tea "github.com/charmbracelet/bubbletea"
  "log"
)

const LogFile = "fya.log"

func main() {
	var f *os.File
	f, _ = os.OpenFile(LogFile, os.O_TRUNC | os.O_CREATE | os.O_RDWR, 0666)
  log.SetOutput(f)
  log.Println("#####################################################")
  log.Println("# WARNING, THIS FILE CONTAINS SENSITIVE INFORMATION #")
  log.Println("#####################################################")

  p := tea.NewProgram(initApp())
  if err := p.Start(); err != nil {
    fmt.Printf("Error: %v", err)
    os.Exit(1)
  }
}
