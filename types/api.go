package types

type APIPort struct {
	PrivatePort int64
	PublicPort  int64
	Type        string
	IP          string
}

type APIContainers struct {
	ID         string `json:"Id"`
	Image      string
	Command    string
	Created    int64
	Status     string
	Ports      []APIPort
	SizeRw     int64
	SizeRootFs int64
	Names      []string
}

type APIShip struct {
	Name string
	Status bool
	Fqdn string
	Port int
	Containers []APIContainers
}
