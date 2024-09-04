package main

import "github.com/ani/ani-ar/fetcher"

// import (
// 	"fmt"
// 	"os"

// 	"github.com/ani/ani-ar/gui"
// 	tea "github.com/charmbracelet/bubbletea"
// )

func main() {

	f := &fetcher.Anime4up{}
	f.GetLazyVideoUrl("https://aname4up.shop/episode/death-note-%d8%a7%d9%84%d8%ad%d9%84%d9%82%d8%a9-dlgqd/")
	// p := tea.NewProgram(gui.InitialModel())
	// if _, err := p.Run(); err != nil {
	// 	fmt.Printf("Alas, there's been an error: %v", err)
	// 	os.Exit(1)
	// }
}
