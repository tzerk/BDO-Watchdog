package lib

import (
	"strings"
	"io/ioutil"
	"os"
	"gopkg.in/yaml.v2"
	"errors"
)

// Telegram and program settings (config.yml)
// Guide: http://sweetohm.net/article/go-yaml-parsers.en.html
type Config struct {
	Token string
	Botid string
	Chatid string
	TimeStamp bool
	Message string
	StayAlive bool
	Process string
	TimeBetweenChecksInS int
	FailLimit int
	KillOnDC bool
	ShutdownOnDC bool
	KillCoherentUI bool
}

func Read_Settings(ex string) (config Config, err error) {

	defer func() {
		if r := recover(); r != nil {
			err = errors.New("Unable to find config.yml, created a new one.")
		}
	}()

	////  SETTINGS
	//--------------------------------------------------------------------------------------------------------------
	// YAML PARSING
	newex := strings.Replace(ex, "BDO-Watchdog.exe", "config.yml", -1)
	// This is necessary for dynamic builds in Jetbrains Gogland IDE
	//newex = strings.Replace(ex, "Application.exe", "config.yml", -1)

	source, err := ioutil.ReadFile(newex)

	if err != nil {
		// in theory, using yml.Marshal() would be more elegant, but we want to preserve the yaml comments
		// as well as set some default values/hints
		defconf :=
			"## Get updates here: https://github.com/tzerk/BDO-Watchdog/releases/\r\n" +
				"## Telegram Bot Settings\r\n" +
				"token: \r\n" +
				"botid: \r\n" +
				"chatid: \r\n" +
				"timestamp: true\r\n" +
				"message: BDO disconnected \r\n" +
				"\r\n" +
				"## Program Settings\r\n" +
				"stayalive: false\r\n" +
				"process: BlackDesert64.exe\r\n" +
				"timebetweenchecksins: 60\r\n" +
				"faillimit: 2\r\n" +
				"\r\n" +
				"# These settings require the .exe to be run with admin rights! \r\n" +
				"killondc: true\r\n" +
				"shutdownondc: false\r\n" +
				"killcoherentui: false"
		ioutil.WriteFile("config.yml", []byte(defconf), os.FileMode(int(0777)))
		panic(err)
	}
	err = yaml.Unmarshal(source, &config)
	if err != nil {
		panic(err)
	}

	return config, err
}