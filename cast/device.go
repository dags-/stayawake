package cast

import (
	"bufio"
	"errors"
	"os/exec"
	"regexp"
	"strings"
)

type Device struct {
	Name string
	UUID string
}

type Status struct {
	State string
	App   string
}

var (
	keyVal = regexp.MustCompile("([^;\\[\\]]+)=([^;\\[\\]]+)")
	status = regexp.MustCompile("(.*) - (.*) \\({(.*)}\\)")
)

func GetDevice(name string) (*Device, error) {
	_, s, e := cmd(statCmd(), "list")
	if e != nil {
		return nil, e
	}

	for s.Scan() {
		t := s.Text()

		// device info printed between brackets [ ]
		if t[0] == '[' {
			props := make(map[string]string)
			match := keyVal.FindAllStringSubmatch(t, -1)
			for i, entry := range match {
				if i > 0 && len(entry) == 3 {
					k := strings.TrimSpace(entry[1])
					v := strings.TrimSpace(entry[2])
					props[k] = v
				}
			}

			device := &Device{
				Name: props["deviceName"],
				UUID: props["uuid"],
			}

			// deviceName matches the name?
			if device.Name == name {
				return device, nil
			}
		}
	}

	return nil, errors.New("device not found")
}

func cmd(cmd string, args ...string) (*exec.Cmd, *bufio.Scanner, error) {
	c := exec.Command(cmd, args...)
	o, e := c.StdoutPipe()
	if e != nil {
		return nil, nil, e
	}
	return c, bufio.NewScanner(o), c.Start()
}
