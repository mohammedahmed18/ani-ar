package main

import (
	"fmt"
	"os"

	"github.com/ani/ani-ar/gui"
	tea "github.com/charmbracelet/bubbletea"
)

func main() {

	p := tea.NewProgram(gui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}
