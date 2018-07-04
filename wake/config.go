package wake

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
)

type Config struct {
	Port    string   `json:"port"`
	Devices []string `json:"devices"`
}

func loadCfg() *Config {
	log.Println("loading config")
	var c Config
	d, e := ioutil.ReadFile("config.json")
	if e == nil {
		e = json.Unmarshal(d, &c)
		if e == nil {
			return &c
		}
	}
	log.Println(e)
	c = Config{Devices: []string{}, Port: "0"}
	saveCfg(&c)
	return &c
}

func saveCfg(c *Config) {
	log.Println("saving config")
	d, e := json.MarshalIndent(c, "", "  ")
	if e != nil {
		log.Println(e)
		return
	}
	e = ioutil.WriteFile("config.json", d, os.ModePerm)
	if e != nil {
		log.Println(e)
	}
}
