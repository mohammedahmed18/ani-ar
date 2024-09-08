package player

import (
	"errors"
	"os/exec"
)

type Player struct {
	bin string

	execute func(url, title string) error
}

var players []Player = []Player{
	{
		bin: "mpv",
		execute: func(u, t string) error {
			_, err := exec.Command(
				"mpv",
				"--title="+t,
				u).Output()
			if err != nil {
				return err
			}
			return nil
		},
	},
	{
		bin: "vlc",
		execute: func(u, t string) error {
			_, err := exec.Command(
				"vlc",
				"--play-and-exit",
				"--meta-title="+t,
				u).Output()
			if err != nil {
				return err
			}
			return nil
		},
	},
}

func RunVideo(url, title string) error {
	for _, player := range players {
		exist := commandExists(player.bin)
		if exist {
			player.execute(url, title)
			return nil
		}
	}
	return errors.New("you don't any players to play the episode try installing vlc or mpv")
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
