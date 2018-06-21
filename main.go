package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/dags-/stayawake/cast"
)

type Config struct {
	Port    string   `json:"port"`
	Devices []string `json:"devices"`
}

var (
	noise = ""
	cfg   = loadCfg()
	lock  = sync.RWMutex{}
)

func main() {
	host := host() // hostname/ip address
	port := port() // server port
	addr := fmt.Sprintf(`http://%s:%s`, host, port)
	noise = addr + "/noise.mp3"
	cast.Log("server running on: ", addr)
	go serve(host, port)
	go handleInput()
	runLoop()
}

func runLoop() {
	for {
		// read device names from config
		lock.Lock()
		devices := cfg.Devices
		lock.Unlock()

		// start monitor task for each device
		wg := sync.WaitGroup{}
		for _, device := range devices {
			wg.Add(1)
			go monitor(device, wg)
		}

		// wait until all tasks complete
		t := time.Now()
		wg.Wait()

		// sleep for remaining time
		d := time.Since(t)
		r := time.Minute - d
		if r.Seconds() > 0 {
			time.Sleep(r)
		}
	}
}

func monitor(device string, wg sync.WaitGroup) {
	defer wg.Done()

	// get status from the device
	s, e := cast.GetStatus(device)
	if e != nil {
		cast.Log(e)
		return
	}

	// cast the noise file if nothing else is playing
	if s == "no applications running" {
		cast.Play(device, noise)
	}
}

func host() string {
	n, e := os.Hostname()
	if e == nil {
		return n + ".local"
	}
	conn, e := net.Dial("udp", "8.8.8.8:80")
	if e != nil {
		panic(e)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func port() string {
	lock.Lock()
	defer lock.Unlock()
	add := fmt.Sprintf("127.0.0.1:%s", cfg.Port)
	l, e := net.Listen("tcp", add)
	if e != nil {
		panic(e)
	}
	defer l.Close()
	parts := strings.Split(l.Addr().String(), ":")
	return parts[1]
}

func loadCfg() *Config {
	cast.Log("loading config")
	var c Config
	d, e := ioutil.ReadFile("config.json")
	if e == nil {
		e = json.Unmarshal(d, &c)
		if e == nil {
			return &c
		}
	}
	cast.Log(e)
	c = Config{Devices: []string{}, Port: "0"}
	saveCfg(&c)
	return &c
}

func saveCfg(c *Config) {
	cast.Log("saving config")
	d, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		cast.Log(e)
		return
	}
	e = ioutil.WriteFile("config.json", d, os.ModePerm)
	if e != nil {
		cast.Log(e)
	}
}

func serve(ip, port string) {
	addr := fmt.Sprintf(`%s:%s`, ip, port)
	fs := http.FileServer(http.Dir("_web"))
	m := http.NewServeMux()
	m.HandleFunc("/config", handleConfig)
	m.Handle("/", fs)
	http.ListenAndServe(addr, m)
}

func handleConfig(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		cast.Log("received config GET request")
		lock.Lock()
		defer lock.Unlock()
		w.Header().Set("Content-Type", "application/json")
		e := json.NewEncoder(w).Encode(cfg)
		if e != nil {
			cast.Log(e)
		}
	} else if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
		cast.Log("received config POST request")
		lock.Lock()
		defer lock.Unlock()
		e := json.NewDecoder(r.Body).Decode(cfg)
		if e != nil {
			cast.Log(e)
		} else {
			saveCfg(cfg)
		}
	} else {
		cast.Log("rejected ", r.Method, " request from ", r.RemoteAddr)
	}
}

func handleInput() {
	s := bufio.NewScanner(os.Stdin)
	for s.Scan() {
		in := s.Text()
		if in == "stop" {
			os.Exit(0)
		}
	}
}
