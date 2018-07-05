package wake

import (
	"sync"
	"time"
)

type device struct {
	name     string
	lastCast time.Time
}

func (d *device) poll(o *options, wg *sync.WaitGroup) {
	defer wg.Done()

	if d.lastCast.IsZero() {
		d.lastCast = time.Now()
		return
	}

	if time.Since(d.lastCast) < o.castInterval {
		return
	}

	s, e := getState(d.name, o.attempts)
	if e != nil {
		logger.Printf("(%s) poll err: %s", d.name, e)
		return
	}

	d.processState(o, s)
}

func (d *device) processState(o *options, s string) {
	switch s {
	case "IDLE", "PAUSED", "STOPPED":
		logger.Printf("(%s) casting: %s\n", d.name, o.audio)
		e := playAudio(d.name, o.audio, o.volume, o.attempts)
		if e != nil {
			logger.Printf("(%s) cast err: %s\n", d.name, e)
		} else {
			d.lastCast = time.Now()
			logger.Printf("(%s) cast complete\n", d.name)
		}
	default:
		// BUFFERING, PLAYING
		d.lastCast = time.Now()
	}
}

func getState(name string, attempts int) (string, error) {
	var err error
	for i := 0; i < attempts; i++ {
		s, e := GetPlayerState(name)
		if e == nil {
			return s, e
		}
		err = e
		time.Sleep(time.Second)
	}
	return "", err
}

func playAudio(name, audio string, volume float64, attempts int) (error) {
	var err error
	for i := 0; i < attempts; i++ {
		e := PlayAudio(name, audio, volume)
		if e == nil {
			return nil
		}
		err = e
		time.Sleep(time.Second)
	}
	return err
}
