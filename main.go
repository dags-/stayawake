package main

import (
	"bufio"
	"os"

	"github.com/dags-/stayawake/wake"
)

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
