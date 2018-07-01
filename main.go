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
	"os/exec"
	"path/filepath"
)

type Config struct {
	Port    string   `json:"port"`
	Devices []string `json:"devices"`
}

type Instance struct {
	Info  *cast.DeviceInfo
	Pause *time.Time
	Idle  *time.Time
}

var (
	lock      sync.RWMutex
	cfg       *Config
	instances  = make(map[string]*Instance)
)

func init() {
	install("play", "/cmd/cast", "github.com/barnybug/go-cast")
	install("status", "", "github.com/vishen/go-chromecast")
}

func main() {
	cfg = loadCfg()
	ip := ip()
	host := hostname(ip)
	port := port(cfg.Port)
	cast.Log("server running on: ", "http://", host, ":", port)
	go serve(ip, port)
	go handleInput()
	runLoop(fmt.Sprintf(`http://%s:%s/audio.mp3`, ip, port))
}

func runLoop(audio string) {
	i := time.Minute * 5
	cfg := loadCfg()

	for {
		// start monitor task for each device
		wg := &sync.WaitGroup{}
		for _, name := range cfg.Devices {
			i, e := getInstance(name)
			if e != nil {
				cast.Log(e)
				continue
			}
			wg.Add(1)
			go monitor(i, audio, wg)
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

func monitor(i *Instance, audio string, wg *sync.WaitGroup) {
	defer wg.Done()
	cast.Log("monitoring ", (*i.Info)["deviceName"])

	s, e := cast.GetStatus(*i.Info)
	if e != nil {
		cast.Log(e)
		return
	}

	if s.State == "idle" {
		if i.Idle == nil {
			t := time.Now()
			i.Idle = &t
			return
		}
		if time.Since(*i.Idle) > time.Duration(time.Minute*10) {
			i.Idle = nil
			e = i.Info.Play(audio)
			if e != nil {
				cast.Log(e)

			}
		}
		return
	}

	if s.State == "paused" {
		if i.Pause == nil {
			t := time.Now()
			i.Pause = &t
			return
		}
		if time.Since(*i.Pause) > time.Duration(time.Minute * 10) {
			i.Pause = nil
			e = i.Info.Play(audio)
			if e != nil {
				cast.Log(e)
			}
		}
		return
	}

	i.Pause = nil
	i.Idle = nil
}

func hostname(ip string) string {
	if n, e := os.Hostname(); e == nil {
		return n
	}
	return ip
}

func ip() string {
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

func getInstance(name string) (*Instance, error) {
	lock.Lock()
	defer lock.Unlock()
	if d, ok := instances[name]; ok {
		return d, nil
	}

	info, e := cast.GetDevice(name)
	if e != nil {
		return nil, e
	}
	cast.Log("got info for device: ", name)
	d := &Instance{Info: info,}
	instances[name] = d
	return d, nil
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
	lock.Lock()
	defer lock.Unlock()

	if r.Method == "GET" {
		cast.Log("received config GET request")
		w.Header().Set("Content-Type", "application/json")
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

func install(name, exe, path string) {
	f := filepath.Join("_bin", name)
	if _, e := os.Stat(path); e == nil {
		return
	}

	os.Mkdir("_bin", os.ModePerm)

	c := exec.Command("go", "get", "-u", path)
	c.Start()
	c.Wait()

	c = exec.Command("go", "build", "-o", f, path + exe)
	c.Start()
	c.Wait()

	cast.Log("installed ", name)
}