package wake

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type Config struct {
	Port    string   `json:"port"`
	Devices []string `json:"devices"`
}

func loadCfg() *Config {
	logger.Println("loading config")
	var c Config
	d, e := ioutil.ReadFile("config.json")
	if e == nil {
		e = json.Unmarshal(d, &c)
		if e == nil {
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
