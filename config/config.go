package config

import (
	"github.com/krane-io/krane/types"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

// Function that return the path of Krane Configuration Files
func ConfigPath() (path string) {
	user, err := user.Current()
	if err != nil {
		log.Fatalln(err)
	}
	if user.Uid == "0" {
		path = "/etc/krane"
	} else {
		path = user.HomeDir + "/.krane"
	}
	//fmt.Printf("%#v", user)
	return path
}

// Returns Krane configuraton
func LoadConfigurationFile() (config types.KraneConfiguration) {
	ymlFile, err := os.Open(ConfigPath() + "/config.yml")
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer ymlFile.Close()
	b, _ := ioutil.ReadAll(ymlFile)
	yaml.Unmarshal(b, &config)
	return config
}
