package cast

import (
	"path/filepath"
	"os"
	"errors"
)

var (
	play string
	stat string
)

func playCmd() string {
	if play == "" {
		play = getCommand("_bin/play")
	}
	return play
}

func statCmd() string {
	if stat == "" {
		stat = getCommand("_bin/status")
	}
	return stat
}

func getCommand(path string) string {
	if exists(path) {
		return path
	}

	x, e := os.Executable()
	if e != nil {
		panic(e)
	}

	p :=  filepath.Join(filepath.Dir(x), path)
	if exists(p) {
		return p
	}

	panic(errors.New("file does not exist: " + p))
}

func exists(path string) bool {
	_, e := os.Stat(path)
	return e == nil
}