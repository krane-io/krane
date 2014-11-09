package types

import (
	"errors"
	"fmt"
)

type KraneConfiguration struct {
	Production Enviroment `yaml:"production"`
	Driver     Driver
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
	Fleet    []Ship
	HighPort int
}

func (configuration *KraneConfiguration) UpdateShips(ships []Ship) error {
	for _, ship := range ships {
		ship.Touched = true
		err := configuration.UpdateShip(ship)
		if err != nil {
			configuration.Production.Fleet = append(configuration.Production.Fleet, ship)
			// if configuration.Production.Fleet == nil {
			// 	configuration.Production.Fleet = []Ship{ship}
			// } else {

			// }
		}
	}
	// Todo: Need to be able to remove ships
	return nil
}

func GetShip(fleet []Ship, name string) Ship {
	if len(fleet) > 0 {
		for _, ship := range fleet {
			if ship.Name == name || ship.Fqdn == name {
				return ship
			}
		}
	}
	return Ship{}
}

func (configuration *KraneConfiguration) GetShip(name string) Ship {
	if len(configuration.Production.Fleet) > 0 {
		for _, ship := range configuration.Production.Fleet {
			if ship.Name == name || ship.Fqdn == name {
				return ship
			}
		}
	}
	return Ship{}
}

func (configuration *KraneConfiguration) UpdateShip(newShip Ship) error {
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
