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
	logger = log.New(os.Stdout, "", log.LstdFlags)
}

func Start() {
	cfg = loadCfg()
	ip := ip()
	host := hostname(ip)
	port := port(cfg.Port)
	audio = fmt.Sprintf(`http://%s:%s/audio.mp3`, ip, port)
	logger.Println("server running on:", "http://", host, ":", port)
	go serve(ip, port)
	go runLoop()
	if !cfg.Debug {
		log.SetOutput(&silent{})
	}
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

	logger.Printf("(%s) polling...\n", name)
	s, e := GetPlayerState(name)
	if e != nil {
		log.Println(e)
		return
	}
	logger.Printf("(%s) state: %s\n", name, s)

	switch s {
	case "BUFFERING":
	case "PLAYING":
		return
	case "IDLE":
		fallthrough
	case "PAUSED":
		fallthrough
	case "STOPPED":
		logger.Printf("(%s) casting: %s\n", name, audio)
		e := PlayAudio(name, audio, volume)
		if e != nil {
			logger.Println(name, e)
			logger.Printf("(%s) err: %s\n", name, e)
		} else {
			logger.Printf("(%s) cast complete\n", name)
		}
		return
	}
}
