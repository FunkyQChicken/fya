package main

import (
//	"github.com/charmbracelet/bubbles"
  "fmt"
  "os"
  "log"
  "errors"

	tea "github.com/charmbracelet/bubbletea"

  "github.com/FunkyQChicken/fya/restaurant"
)

var FyaDir = os.ExpandEnv("$HOME/.fya/")
var LogFile = FyaDir + "fya.log"

func main() {
  restaurant.BaseDir = FyaDir

  err := os.Mkdir(FyaDir, 0755)
 
  if err != nil && ! errors.Is(err, os.ErrExist) {
    log.Fatalf("ERROR: problem setting up the directory for FYA: %s\n", err)
  }
	var f *os.File
	f, err = os.OpenFile(LogFile, os.O_TRUNC | os.O_CREATE | os.O_RDWR, 0666)
  if err != nil {
    log.Fatalf("ERROR: problem setting up the log file for FYA: %s\n", err)
  }
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
