package main

import (
	"bufio"
	"log"
	"os"

	"github.com/dags-/stayawake/wake"
)

type wr struct {}

func (w *wr) Write(data []byte) (int, error) {
	return 0, nil
}

func main() {
	log.SetOutput(&wr{})

	wake.Start()
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		in := s.Text()
		if in == "stop" {
			os.Exit(0)
		}
	}
}