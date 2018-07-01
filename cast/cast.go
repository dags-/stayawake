package cast

import (
	"fmt"
	"os/exec"
	"time"
)

func (info *DeviceInfo) Play(mp3 string) (error) {
	device := (*info)["deviceName"]

	Log("casting wakeup noise")
	if e := execSync(playCmd(), "--name", device, "volume", "0.2"); e != nil {
		return e
	}

	time.Sleep(time.Second)

	// execute play playCommand: `cast --name <device> media play <mp3_address>
	if e := execSync(playCmd(), "--name", device, "media", "play", mp3); e != nil {
		return e
	}

	// wait for audio to play a bit
	time.Sleep(time.Second * 2)

	// get the device status
	s, e := GetStatus(*info)
	if  e != nil {
		return e
	}
	
	if s.App != "none" {
		return nil
	}

	// quit the default media receiver app
	Log("wakeup complete")
	if e := execSync(playCmd(), "--name", device, "quit"); e != nil {
		return e
	}

	// reset volume
	 return execSync(playCmd(), "--name", device, "volume", "0.75")
}

func execSync(cmd string, args ...string) error {
	c := exec.Command(cmd, args...)
	e := c.Start()
	if e != nil {
		return e
	}
	return c.Wait()
}

func Log(message ...interface{}) {
	fmt.Println(time.Now().Format(time.Stamp), ":", fmt.Sprint(message...))
}