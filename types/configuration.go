package types

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
	Fleet      Fleet
	HighPort   int
	SshProfile string `yaml:"ssl_profile"`
}
