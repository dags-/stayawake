package cast

import (
	"os/exec"
	"time"
)

func (d *Device) Play(mp3 string) (error) {
	Log("casting wakeup noise")
	if e := execSync(playCmd(), "--name", d.Name, "volume", "0.2"); e != nil {
		return e
	}
	time.Sleep(time.Second)


	// execute play playCommand: `cast --name <device> media play <mp3_address>
	if e := execSync(playCmd(), "--name", d.Name, "media", "play", mp3); e != nil {
		return e
	}

	// wait for audio to play a bit
	time.Sleep(time.Second * 2)

	// get the device status
	s, e := d.GetStatus()
	if  e != nil {
		return e
	}
	
	if s.App != "none" {
		return nil
	}

	// quit the default media receiver app
	Log("wakeup complete")
	if e := execSync(playCmd(), "--name", d.Name, "quit"); e != nil {
		return e
	}

	// reset volume
	 return execSync(playCmd(), "--name", d.Name, "volume", "0.75")
}

func execSync(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	e := c.Start()
	if e != nil {
		return e
	}
	return c.Wait()
}