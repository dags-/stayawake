package wake

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	logger *log.Logger
)

type silent struct{}

type options struct {
	audio        string        // the mp3 url
	volume       float64       // volume to play audio at
	attempts     int           // number of attempts at getting status/playing audio
	pollInterval time.Duration // time between device polls
	castInterval time.Duration // time between casts
}

func init() {
	logger = log.New(os.Stdout, "", log.LstdFlags)
}

func Start() {
	cfg := loadCfg()

	ip := ip()
	host := hostname(ip)
	port := port(cfg.Port)
	audio := fmt.Sprintf(`http://%s:%s/audio.mp3`, ip, port)
	logger.Printf("server running on: http://%s:%s", host, port)

	options := &options{
		audio:        audio,
		volume:       0.75,
		attempts:     5,
		castInterval: time.Duration(cfg.CastInterval) * time.Minute,
		pollInterval: time.Duration(cfg.PollInterval) * time.Minute,
	}

	manager := newManager()
	manager.setDevices(cfg.Devices)

	go manager.startServer(ip, port)
	go manager.startLoop(options)

	if !cfg.Debug {
		log.SetOutput(&silent{})
	}
}

func (w *silent) Write(data []byte) (int, error) {
	return 0, nil
}
