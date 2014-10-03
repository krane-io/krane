package types

type Fleet struct {
	Ships []Ship
}

type Ship struct {
	Id        string `json:"id"`
	Name      string `json:"name"`
	Fqdn      string `json:"fqdn"`
	Ip        string `json:"ipAddress"`
	State     string `json:"state"`
	Os        string `json:"os"`
	Plan      string `json:"plan"`
	Port      int    `json:"port"`
	Schema    string `json:"schema"`
	LocalPort int
	Touched   bool
}

func ConvertFromApiShipToShip(apiShip *APIShip) Ship {
	return Ship{
		Fqdn:   apiShip.Fqdn,
		Name:   apiShip.Name,
		Port:   apiShip.Port,
		Schema: "http",
	}
}
