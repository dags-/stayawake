package cast

import (
	"errors"
	"strings"
)

func (d *Device) GetStatus() (*Status, error) {
	if d.UUID == "" {
		return nil, errors.New("unable to determine device id")
	}

	_, s, e := cmd(statCmd(), "--uuid", d.UUID, "status")
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
