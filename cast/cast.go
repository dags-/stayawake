package cast

import (
	"bufio"
	"errors"
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

var (
	// get the cast command depending on os/arch
	command = getCommand()
)

func GetStatus(device string) (string, error) {
	// make 3 attempts
	for i := 1; i <= 3; i++ {
		// execute status command: `cast --name <device> status`
		c := exec.Command(command, "--name", device, "status")
		o, e := c.StdoutPipe()
		if e != nil {
			Log(e)
			time.Sleep(5 * time.Second)
			continue
		}

		// wrap command output in a scanner
		s := bufio.NewScanner(o)
		e = c.Start()
		if e != nil {
			Log(e)
			time.Sleep(5 * time.Second)
			continue
		}

		// read command output, status comes after 'connected'
		for s.Scan() {
			line := strings.ToLower(s.Text())
			if line == "connected" && s.Scan() {
				return s.Text(), nil
			}
			if line == "timeout exceeded" {
				Log(line, " (device: ", device, " #", i, ")")
			}
		}
	}
	// all 3 attempts failed, return error
	return "", errors.New("unable to retrieve status of " + device)
}

func Play(device, mp3 string) (error) {
	Log("casting wakeup noise")
	// execute play command: `cast --name <device> media play <mp3_address>
	c := exec.Command(command, "--name", device, "media", "play", mp3)
	e := c.Start()

	// wait for audio to play a bit
	time.Sleep(time.Second * 15)

	// get the device status
	s, e := GetStatus(device)
	if  e != nil {
		return e
	}

	// if something other than the 'default media receiver' is using the device then stop here
	if s != "[default media receiver] default media receiver" {
		Log("wakeup interrupted")
		return nil
	}

	// quit the default media receiver app
	Log("wakeup complete")
	c = exec.Command(command, "--name", device, "quit")
	return c.Start()
}

func Log(message ...interface{}) {
	fmt.Println(time.Now().Format(time.Stamp), ":", fmt.Sprint(message...))
}

func getCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "_bin/cast-mac-amd64.dms"
	case "windows":
		return "_bin/cast-windows-amd64.exe"
	case "linux":
		// rasp pi
		if runtime.GOARCH == "arm" {
			return "_bin/cast-linux-arm.dms"
		}
		return "_bin/cast-linux-amd64.dms"
	}
	// can't run on this platform
	panic(errors.New("unsupported platform"))
}