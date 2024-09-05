package main

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/ani/ani-ar/gui"
)

func main() {
	p := tea.NewProgram(gui.InitialModel())
	if _, err := p.Run(); err != nil {
		fmt.Printf("Alas, there's been an error: %v", err)
		os.Exit(1)
	}
}

// func main() {
// 	var f fetcher.Anime4up
//
// 	u := f.GetLazyVideoUrl(
// 		"https://aname4up.shop/episode/death-note-%D8%A7%D9%84%D8%AD%D9%84%D9%82%D8%A9-p4nba",
// 	)
//
// 	println(u)
// }
