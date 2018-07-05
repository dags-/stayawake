package wake

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"strconv"
)

type Config struct {
	Port         string   `json:"port"`
	Debug        bool     `json:"debug"`
	PollInterval int      `json:"poll_interval"`
	CastInterval int      `json:"cast_interval"`
	Devices      []string `json:"devices"`
}

func loadCfg() *Config {
	logger.Println("loading config")
	var c Config
	d, e := ioutil.ReadFile("config.json")
	if e == nil {
		e = json.Unmarshal(d, &c)
		if e == nil {
			ensure(&c)
			saveCfg(&c)
			return &c
		}
	}
	logger.Println(e)
	c = Config{Devices: []string{}, Port: "0"}
	saveCfg(&c)
	return &c
}

func saveCfg(c *Config) {
	logger.Println("saving config")
	d, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		logger.Println(e)
		return
	}
	e = ioutil.WriteFile("config.json", d, os.ModePerm)
	if e != nil {
		logger.Println(e)
	}
}

func ensure(c *Config) {
	if c.PollInterval < 1 {
		c.PollInterval = 4
	}
	if c.CastInterval < 1 {
		c.CastInterval = 10
	}
	if _, e := strconv.Atoi(c.Port); e != nil {
		c.Port = "0"
	}
}
