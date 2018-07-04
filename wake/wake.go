package wake

import (
	"fmt"
	"log"
	"sync"
	"time"
)

var (
	cfg      *Config
	lock     sync.RWMutex
	audio    string
	volume   float64
	interval = time.Duration(time.Minute * 15)
)

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
	log.Println("polling", name)

	s, e := GetPlayerState(name)
	if e != nil {
		log.Println(e)
		return
	}

	switch s {
	case "BUFFERING":
	case "PLAYING":
		return
	case "IDLE":
	case "PAUSED":
	case "STOPPED":
		PlayAudio(name, audio, volume)
		break
	default:
		log.Println("unknown status: ", s)
	}
}
