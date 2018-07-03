package wake

import (
	"fmt"
	"sync"
	"time"

	"github.com/dags-/stayawake/cast"
)

type Instance struct {
	Device *cast.Device
	Pause  *time.Time
	Idle   *time.Time
}

var (
	cfg       *Config
	lock      sync.RWMutex
	audio     string
	timeout   = time.Duration(time.Minute * 5)
	interval  = time.Duration(time.Minute * 10)
	instances = make(map[string]*Instance)
)

func Start() {
	cfg = loadCfg()
	ip := ip()
	host := hostname(ip)
	port := port(cfg.Port)
	audio = fmt.Sprintf(`http://%s:%s/audio.mp3`, ip, port)
	cast.Log("server running on: ", "http://", host, ":", port)
	go serve(ip, port)
	go runLoop()
}

func runLoop() {
	cfg := loadCfg()

	for {
		t := time.Now()

		// start monitor task for each device
		wg := &sync.WaitGroup{}
		for _, name := range cfg.Devices {
			i, e := instance(name)
			if e != nil {
				cast.Log(e)
				continue
			}
			wg.Add(1)
			go i.poll(wg)
		}

		// wait until all tasks complete
		wg.Wait()

		// sleep for remaining time
		r := interval - time.Since(t)
		if r.Seconds() > 0 {
			time.Sleep(r)
		}
	}
}

func instance(name string) (*Instance, error) {
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
	d := &Instance{Device: info,}
	instances[name] = d
	return d, nil
}

func (i *Instance) poll(wg *sync.WaitGroup) {
	defer wg.Done()
	cast.Log("polling ", i.Device.Name)

	s, e := i.Device.GetStatus()
	if e != nil {
		cast.Log(e)
		return
	}

	if s.State == "idle" {
		i.idle()
		return
	}

	if s.State == "paused" {
		i.paused()
		return
	}

	cast.Log("neither!? ", s.State, " - ", s.App)

	i.Pause = nil
	i.Idle = nil
}

func (i *Instance) idle() {
	if i.Idle == nil {
		t := time.Now()
		i.Idle = &t
		i.Pause = nil
		return
	}
	if time.Since(*i.Idle) > timeout {
		i.Idle = nil
		i.Pause = nil
		e := i.Device.Play(audio)
		if e != nil {
			cast.Log(e)
		}
	}
}

func (i *Instance) paused() {
	if i.Pause == nil {
		t := time.Now()
		i.Idle = nil
		i.Pause = &t
		return
	}
	if time.Since(*i.Pause) > timeout {
		i.Idle = nil
		i.Pause = nil
		e := i.Device.Play(audio)
		if e != nil {
			cast.Log(e)
		}
	}
}
