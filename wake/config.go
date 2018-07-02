package wake

import (
	"encoding/json"
	"io/ioutil"
	"os"

	"github.com/dags-/stayawake/cast"
)

type Config struct {
	Port    string   `json:"port"`
	Devices []string `json:"devices"`
}

func loadCfg() *Config {
	cast.Log("loading config")
	var c Config
	d, e := ioutil.ReadFile("config.json")
	if e == nil {
		e = json.Unmarshal(d, &c)
		if e == nil {
			return &c
		}
	}
	cast.Log(e)
	c = Config{Devices: []string{}, Port: "0"}
	saveCfg(&c)
	return &c
}

func saveCfg(c *Config) {
	cast.Log("saving config")
	d, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		cast.Log(e)
		return
	}
	e = ioutil.WriteFile("config.json", d, os.ModePerm)
	if e != nil {
		cast.Log(e)
	}
}
