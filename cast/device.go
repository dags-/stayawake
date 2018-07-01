package cast

import (
	"os/exec"
	"bufio"
	"strings"
	"errors"
	"regexp"
)

type DeviceInfo map[string]string

type Status struct {
	State string
	App   string
}

var (
	keyVal        = regexp.MustCompile("([^;\\[\\]]+)=([^;\\[\\]]+)")
	status        = regexp.MustCompile("(.*) - (.*) \\({(.*)}\\)")
)

func GetStatus(device DeviceInfo) (*Status, error) {
	id := device["uuid"]
	if id == "" {
		return nil, errors.New("unable to determine device id")
	}

	s, e := cmd(statCmd(), "--uuid", id, "status")
	if e != nil {
		return nil, e
	}

	// last line of the output is the device's status
	var last string
	for s.Scan() {
		last = s.Text()
	}

	if strings.HasPrefix(last, "Chromecast is idle") {
		return &Status{State: "idle", App: "none"}, nil
	}

	// (state) - (application) ({metadata}, time=###, volume=###)
	match := status.FindAllStringSubmatch(last, -1)
	if len(match) > 0 {
		groups := match[0]
		if len(groups) > 2 {
			var status Status
			status.State = strings.ToLower(groups[1])
			status.App = groups[2]
			return &status, nil
		}
	}

	return nil, errors.New("unable to read status")
}

func GetDevice(name string) (*DeviceInfo, error) {
	s, e := cmd(statCmd(), "list")
	if e != nil {
		return nil, e
	}

	for s.Scan() {
		t := s.Text()

		// device info printed between brackets [ ]
		if t[0] == '[' {
			device := DeviceInfo{}
			match := keyVal.FindAllStringSubmatch(t, -1)
			for i, entry := range match {
				if i > 0 && len(entry) == 3 {
					k := strings.TrimSpace(entry[1])
					v := strings.TrimSpace(entry[2])
					device[k] = v
				}
			}
			// deviceName matches the name?
			if device["deviceName"] == name {
				return &device, nil
			}
		}
	}

	return nil, errors.New("device not found")
}

func cmd(cmd string, args ...string) (*bufio.Scanner, error) {
	c := exec.Command(cmd, args...)
	o, e := c.StdoutPipe()
	if e != nil {
		return nil, e
	}
	return bufio.NewScanner(o), c.Start()
}
