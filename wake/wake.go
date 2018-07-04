package wake

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var (
	cfg      *Config
	lock     sync.RWMutex
	logger   *log.Logger
	audio    string
	volume   float64
	interval = time.Duration(time.Minute * 15)
)

func init() {
	logger = log.New(os.Stdout, "", 0)
}

func Start() {
	cfg = loadCfg()
	ip := ip()
	host := hostname(ip)
	port := port(cfg.Port)
	audio = fmt.Sprintf(`http://%s:%s/audio.mp3`, ip, port)
	log.Println("server running on:", "http://", host, ":", port)
	go serve(ip, port)
	go runLoop()
}

func runLoop() {
	cfg := loadCfg()

	for {
		t := time.Now()
		wg := &sync.WaitGroup{}
		for _, name := range cfg.Devices {
			wg.Add(1)
			go poll(name, wg)
		}

		wg.Wait()
		r := interval - time.Since(t)
		if r.Seconds() > 0 {
			time.Sleep(r)
		}
	}
}

func poll(name string, wg *sync.WaitGroup) {
	defer wg.Done()

	logger.Println(name, "polling...")
	s, e := GetPlayerState(name)
	if e != nil {
		log.Println(e)
		return
	}
	logger.Println(name, "state:", s)

	switch s {
	case "BUFFERING":
	case "PLAYING":
		return
	case "IDLE":
	case "PAUSED":
	case "STOPPED":
		logger.Println(name, "casting ", audio)
		e := PlayAudio(name, audio, volume)
		if e != nil {
			logger.Println(name, e)
		} else {
			logger.Print(name, "cast complete")
		}
		break
	default:
		logger.Println(name, "unknown state:", s)
	}
}
