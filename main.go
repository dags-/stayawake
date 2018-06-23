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

func main() {
	cfg := loadCfg()
	ip := ip()
	host := hostname()
	if host == "" {
		host = ip
	}
	port := port(cfg.Port)
	addr := fmt.Sprintf(`http://%s:%s/`, ip, port)
	cast.Log("server running on: ", addr)
	go serve(ip, port)
	go handleInput()
	runLoop(addr + "/audio.mp3")
}

func runLoop(audio string) {
	i := time.Minute * 10

	for {
		// read device names from config
		cfg := loadCfg()

		// start monitor task for each device
		wg := &sync.WaitGroup{}
		for _, device := range cfg.Devices {
			wg.Add(1)
			go monitor(device, audio, wg)
		}

		// wait until all tasks complete
		t := time.Now()
		wg.Wait()

		// sleep for remaining time
		r := i - time.Since(t)
		if r.Seconds() > 0 {
			time.Sleep(r)
		}
	}
}

func monitor(device, audio string, wg *sync.WaitGroup) {
	defer wg.Done()
	cast.Log("monitoring ", device)

	// get status from the device
	s, e := cast.GetStatus(device)
	if e != nil {
		cast.Log(e)
		return
	}

	// cast the noise file if nothing else is playing
	if s == "no applications running" {
		e = cast.Play(device, audio)
		if e != nil {
			cast.Log(e)
		}
	}
}

func hostname() string {
	if n, e := os.Hostname(); e == nil {
		return n
	}
	return ""
}

func ip() string {
	n, e := os.Hostname()
	if e == nil {
		return n
	}
	conn, e := net.Dial("udp", "8.8.8.8:80")
	if e != nil {
		panic(e)
	}
	defer conn.Close()
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String()
}

func port(port string) string {
	if port == "" {
		port = "0"
	}
	add := fmt.Sprintf("127.0.0.1:%s", port)
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
		w.Header().Set("Content-Type", "application/json")
		cfg := loadCfg()
		e := json.NewEncoder(w).Encode(cfg)
		if e != nil {
			cast.Log(e)
		}
		return
	}

	if r.Method == "POST" && r.Header.Get("Content-Type") == "application/json" {
		cast.Log("received config POST request")
		var cfg Config
		e := json.NewDecoder(r.Body).Decode(&cfg)
		if e == nil {
			saveCfg(&cfg)
		} else {
			cast.Log(e)
		}
		return
	}

	cast.Log("rejected ", r.Method, " request from ", r.RemoteAddr)
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