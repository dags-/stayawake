package cast

import (
	"os/exec"
)

func (d *Device) Play(mp3 string) (error) {
	// execute quit command: `cast --name <device> quit`
	if e := execSync(playCmd(), "--name", d.Name, "quit"); e != nil {
		return e
	}
	// execute volume command: `cast --name <device> volume <vol>`
	if e := execSync(playCmd(), "--name", d.Name, "volume", "0.25"); e != nil {
		return e
	}
	// execute play playCommand: `cast --name <device> media play <mp3_address>
	Log("casting wakeup call")
	if e := execSync(playCmd(), "--name", d.Name, "media", "play", mp3); e != nil {
		return e
	}
	// reset volume
	return nil
}

func execSync(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	return c.Run()
}
