package types

type KraneConfiguration struct {
	Production Enviroment `yaml:"production"`
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
	Fleet []Ship
}

type Ship struct {
	Fqdn      string
	Name      string
	Port      int
	Schema    string
	LocalPort int
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

func ConvertFromApiShipToShip(apiShip *APIShip) Ship {
	return Ship{
		Fqdn:   apiShip.Fqdn,
		Name:   apiShip.Name,
		Port:   apiShip.Port,
		Schema: "http",
	}
}
