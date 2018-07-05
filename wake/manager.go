package wake

import (
	"sync"
	"time"
)

type manager struct {
	lock    sync.RWMutex
	devices map[string]*device
}

func newManager() *manager {
	return &manager{
		devices: make(map[string]*device),
	}
}

func (m *manager) startLoop(o *options) {
	for {
		t := time.Now()
		m.tick(o)
		remaining := o.pollInterval - time.Since(t)
		if remaining > 0 {
			time.Sleep(remaining)
		}
	}
}

func (m *manager) tick(o *options) {
	m.lock.Lock()
	defer m.lock.Unlock()
	wg := &sync.WaitGroup{}
	wg.Add(len(m.devices))
	for _, device := range m.devices {
		go device.poll(o, wg)
	}
	wg.Wait()
}

func (m *manager) setDevices(devices []string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	dif := make(map[string]bool)
	for _, name := range devices {
		dif[name] = true
	}

	for name := range m.devices {
		if _, keep := dif[name]; !keep {
			dif[name] = false
		}
	}

	for name, insert := range dif {
		_, exists := m.devices[name]
		if insert && !exists {
			m.devices[name] = &device{name: name}
		} else if !insert && exists {
			delete(m.devices, name)
		}
	}
}
