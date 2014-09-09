package main

import (
	"io/ioutil"
	"log"

	"gopkg.in/yaml.v1"
)

// appConfigT: the yaml config file we get on app startup will be parsed into
// this
type appConfigT struct {
	// default config for uploads
	IncomingIP              string `yaml:"IncomingIP"`
	IncomingPort            uint   `yaml:"IncomingPort"`
	UploadChunkSizeKB       uint   `yaml:"UploadChunkSizeKB"`
	UploadSendAhead         uint   `yaml:"UploadSendAhead"`
	UploadMaxIdleDurationS  uint   `yaml:"UploadMaxIdleDurationS"`
	StorageDir              string `yaml:"StorageDir"`
	HandoverTimeoutS        uint   `yaml:"HandoverTimeoutS"`
	HandoverConfirmTimeoutS uint   `yaml:"HandoverConfirmTimeoutS"`
}

func LoadConfig(path string) (c *appConfigT, e error) {
	// read config file
	var fileContent []byte
	fileContent, e = ioutil.ReadFile(path)
	if e != nil {
		log.Printf("Couldn't read config file %s: %s", path, e.Error())
		return
	}

	// parse config file
	c = new(appConfigT)
	e = yaml.Unmarshal(fileContent, c)
	if e != nil {
		log.Printf("Couldn't parse config file %s: %s", path, e.Error())
		return
	}

	// TODO: fiddle in other sources for config vars: env vars, command line
	return
}
