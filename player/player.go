package player

import (
	"errors"
	"log"
	"os/exec"
)

type Player struct {
	bin string

	execute func(url, title string) *exec.Cmd
}

var players []Player = []Player{
	{
		bin: "mpv",
		execute: func(u, t string) *exec.Cmd {
			cmd := exec.Command(
				"mpv",
				"--title="+t,
				u)
			return cmd
		},
	},
	{
		bin: "vlc",
		execute: func(u, t string) *exec.Cmd {
			cmd := exec.Command(
				"vlc",
				"--play-and-exit",
				"--meta-title="+t,
				u)
			return cmd
		},
	},
}

func RunVideo(url, title string) error {
	for _, player := range players {
		exist := commandExists(player.bin)
		if exist {
			cmd := player.execute(url, title)
			err := cmd.Start()
			if err != nil {
				return err
			}
			log.Printf("video played with PID %d\n", cmd.Process.Pid)
			return nil
		}
	}
	return errors.New("you don't any players to play the episode try installing vlc or mpv")
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
