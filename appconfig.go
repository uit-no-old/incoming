package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path"

	"bitbucket.org/kardianos/osext"
	"gopkg.in/yaml.v1"
)

// appConfigT: the yaml config file we get on app startup will be parsed into
// this
type appConfigT struct {
	// default config for uploads
	IncomingIP                  string `yaml:"IncomingIP"`
	IncomingPort                uint   `yaml:"IncomingPort"`
	UploadChunkSizeKB           uint   `yaml:"UploadChunkSizeKB"`
	UploadSendAhead             uint   `yaml:"UploadSendAhead"`
	UploadMaxIdleDurationS      uint   `yaml:"UploadMaxIdleDurationS"`
	WebsocketConnectionTimeoutS uint   `yaml:"WebsocketConnectionTimeoutS"`
	StorageDir                  string `yaml:"StorageDir"`
	HandoverTimeoutS            uint   `yaml:"HandoverTimeoutS"`
	HandoverConfirmTimeoutS     uint   `yaml:"HandoverConfirmTimeoutS"`
}

func LoadConfig() (c *appConfigT, e error) {
	// find out which file to load
	fPath := ""
	if _, e = os.Stat("incoming_cfg.yaml"); e == nil {
		fPath = "incoming_cfg.yaml"
	} else {
		programDir, _ := osext.ExecutableFolder()
		candPath := path.Join(programDir, "incoming_cfg.yaml")
		if _, e := os.Stat(candPath); e == nil {
			fPath = candPath
		}
	}

	if fPath == "" {
		e = fmt.Errorf("didn't find config file anywhere!")
		return
	}

	var fileContent []byte
	fileContent, e = ioutil.ReadFile(fPath)
	if e != nil {
		log.Printf("Couldn't read config file %s: %s", fPath, e.Error())
		return
	}

	// parse config file
	c = new(appConfigT)
	e = yaml.Unmarshal(fileContent, c)
	if e != nil {
		log.Printf("Couldn't parse config file %s: %s", fPath, e.Error())
		return
	}

	// TODO: fiddle in other sources for config vars: env vars, command line
	return
}
