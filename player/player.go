package player

import (
	"errors"
	"os/exec"
)

type Player struct {
	bin     string
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

func RunVideo(url, title string) (*exec.Cmd, error) {
	for _, player := range players {
		exist := commandExists(player.bin)
		if exist {
			cmd := player.execute(url, title)
			err := cmd.Start()
			if err != nil {
				return nil, err
			}
			//		log.Printf("video played with PID %d\n", cmd.Process.Pid)
			return cmd, nil
		}
	}
	return nil, errors.New("you don't any players to play the episode try installing vlc or mpv")
}

func commandExists(cmd string) bool {
	_, err := exec.LookPath(cmd)
	return err == nil
}
