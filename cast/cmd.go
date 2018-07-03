package cast

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

var (
	play string
	stat string
)

func Log(message ...interface{}) {
	fmt.Println(time.Now().Format(time.Stamp), ":", fmt.Sprint(message...))
}

func GetFile(path string) (string, error) {
	if exists(path) {
		return path, nil
	}

	x, e := os.Executable()
	if e != nil {
		return "", e
	}

	path = filepath.Join(filepath.Dir(x), path)
	if exists(path) {
		return path, nil
	}
	return path, errors.New("file not found")
}

func playCmd() string {
	if play == "" {
		play = mustCommand("bin/play")
	}
	return play
}

func statCmd() string {
	if stat == "" {
		stat = mustCommand("bin/status")
	}
	return stat
}

func mustCommand(path string) string {
	p, e := GetFile(path)
	if e != nil {
		panic(e)
	}
	return p
}

func exists(path string) bool {
	_, e := os.Stat(path)
	return e == nil
}
