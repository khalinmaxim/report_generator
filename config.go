package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"time"
)

type Config struct {
	LogLevel string   `json:"loglevel"`
	Database Database `json:"database"`
	Outpath  string   `json:"outpath"`
	Format   string   `json:"format"`
	Dry      bool     `json:"dry"`
	Test     bool     `json:"test"`
	Testpath string   `json:"testpath"`
}

func (config *Config) read() {
	byteValue, err := ioutil.ReadFile("reporter.cfg")
	if err == nil {
		err = json.Unmarshal(byteValue, &config)
		if err == nil {
			log.WithFields(log.Fields{
				"loglevel":       config.LogLevel,
				"drymode":        config.Dry,
				"testmode":       config.Test,
				"host":           config.Database.Host,
				"port":           config.Database.Port,
				"name":           config.Database.Name,
				"user":           config.Database.UserName,
				"channel":        config.Database.Channel,
				"channeltimeout": config.Database.Timeout * time.Second,
				"outpath":        config.Outpath,
				"testpath":       config.Testpath,
			}).Info("Config")
		} else {
			log.Error(string(byteValue))
			log.Fatal(err)
		}
	} else {
		log.Fatal(err)
	}
}
