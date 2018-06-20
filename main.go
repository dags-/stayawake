package main

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"strings"
	"time"
	"os"
)

func main() {
	ip := getIp()
	port := getPort()
	mp3 := fmt.Sprintf(`http://%s:%s/noise.mp3`, ip, port)
	go serveFiles(ip, port)
	go handleStop()
	monitor(mp3)
}

func log(message string) {
	fmt.Println(time.Now().Format(time.Stamp), ":", message)
}

func doPanic(e error) {
	if e != nil {
		panic(e)
	}
}

func getIp() string {
	conn, e := net.Dial("udp", "8.8.8.8:80")
	doPanic(e)
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func getPort() string {
	l, e := net.Listen("tcp", "127.0.0.1:0")
	doPanic(e)
	defer l.Close()
	parts := strings.Split(l.Addr().String(), ":")
	return parts[1]
}

func serveFiles(ip, port string) {
	u := fmt.Sprintf(`%s:%s`, ip, port)
	h := http.FileServer(http.Dir("audio"))
	http.ListenAndServe(u, h)
}

func handleStop() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		in := s.Text()
		if in == "stop" {
			os.Exit(0)
		}
	}
}

func monitor(mp3 string) {
	for {
		log("monitoring cast audio")
		s := getStatus()
		if s == "No applications running" {
			play(mp3)
		}
		time.Sleep(time.Minute)
	}
}

func play(mp3 string) {
	log("casting wakeup noise")
	c := exec.Command(getCommand(), "--name", "Eneby", "media", "play", mp3)
	e := c.Start()

	time.Sleep(time.Second * 10)

	if getStatus() == "[Default Media Receiver] Default Media Receiver" {
		log("wakeup complete")
		c = exec.Command(getCommand(), "--name", "Eneby", "quit")
		e = c.Start()
		doPanic(e)
	} else {
		log("wakeup interrupted")
	}
}

func getStatus() string {
	for {
		c := exec.Command(getCommand(), "--name", "Eneby", "status")
		o, e := c.StdoutPipe()
		doPanic(e)

		s := bufio.NewScanner(o)
		e = c.Start()
		doPanic(e)

		for s.Scan() {
			if s.Text() == "Connected" && s.Scan() {
				return s.Text()
			}
		}

		time.Sleep(5 * time.Second)
	}
}

func getCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "./bin/cast-mac-amd64.dms"
	case "windows":
		return "./bin/cast-windows-amd64.exe"
	case "linux":
		if runtime.GOARCH == "arm" {
			return "./bin/cast-linux-arm.dms"
		}
		return "./bin/cast-linux-amd64.dms"
	}
	panic(errors.New("unsupported platform"))
}