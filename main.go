package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/dags-/stayawake/cast"
	"github.com/dags-/stayawake/wake"
)

func init() {
	install("bin/play", "/cmd/cast", "github.com/barnybug/go-cast")
	install("bin/status", "", "github.com/vishen/go-chromecast")
}

func main() {
	wake.Start()
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		in := s.Text()
		if in == "stop" {
			os.Exit(0)
		}
	}
}

func install(file, exe, pkg string) {
	f, e := filepath.Abs(file)
	if e != nil {
		panic(e)
	}

	if _, e := os.Stat(f); e == nil {
		fmt.Println(file, "exists, skipping install")
		return
	}

	os.Mkdir(filepath.Dir(f), os.ModePerm)

	c := exec.Command("go", "get", "-u", pkg)
	c.Start()
	c.Wait()

	c = exec.Command("go", "build", "-o", f, pkg+exe)
	c.Start()
	c.Wait()

	cast.Log("installed ", file)
}
