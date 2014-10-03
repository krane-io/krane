package config

import (
	"errors"
	"fmt"
	"github.com/krane-io/krane/drivers/base"
	"github.com/krane-io/krane/types"
	"gopkg.in/yaml.v1"
	"io/ioutil"
	"log"
	"os"
	"os/user"
)

type KraneConfiguration struct {
	Production Enviroment `yaml:"production"`
	Driver     base.Driver
}

type Enviroment struct {
	Server struct {
		Driver string
		Host   struct {
			Fqdn   string
			Name   string
			Port   int
			Schema string
		}
	}
	Fleet    []types.Ship
	HighPort int
}

func (configuration *KraneConfiguration) UpdateShips(ships []types.Ship) error {
	for _, ship := range ships {
		ship.Touched = true
		err := configuration.UpdateShip(ship)
		if err != nil {
			configuration.Production.Fleet = append(configuration.Production.Fleet, ship)
			// if configuration.Production.Fleet == nil {
			// 	configuration.Production.Fleet = []types.Ship{ship}
			// } else {

			// }
		}
	}
	// Todo: Need to be able to remove ships
	return nil
}

func (configuration *KraneConfiguration) GetShip(name string) types.Ship {
	if len(configuration.Production.Fleet) > 0 {
		for _, ship := range configuration.Production.Fleet {
			if ship.Name == name || ship.Fqdn == name {
				return ship
			}
		}
	}
	return types.Ship{}
}

func (configuration *KraneConfiguration) UpdateShip(newShip types.Ship) error {
	if len(configuration.Production.Fleet) > 0 {
		for index, ship := range configuration.Production.Fleet {
			if ship.Name == newShip.Name && ship.Fqdn == newShip.Fqdn {
				configuration.Production.Fleet[index].Fqdn = newShip.Fqdn
				configuration.Production.Fleet[index].Id = newShip.Id
				configuration.Production.Fleet[index].Ip = newShip.Ip
				if newShip.LocalPort > 0 {
					configuration.Production.Fleet[index].LocalPort = newShip.LocalPort
				}
				configuration.Production.Fleet[index].Name = newShip.Name
				configuration.Production.Fleet[index].Os = newShip.Os
				configuration.Production.Fleet[index].Plan = newShip.Plan
				configuration.Production.Fleet[index].Port = newShip.Port
				configuration.Production.Fleet[index].State = newShip.State
				return nil
			}
		}
	}
	return errors.New(fmt.Sprintf("Can not update Ship %s", newShip.Fqdn))
}

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
func LoadConfigurationFile() (config KraneConfiguration) {
	ymlFile, err := os.Open(ConfigPath() + "/config.yml")
	if err != nil {
		log.Fatalln("Error opening file:", err)
	}
	defer ymlFile.Close()
	b, _ := ioutil.ReadAll(ymlFile)
	yaml.Unmarshal(b, &config)
	return config
}
