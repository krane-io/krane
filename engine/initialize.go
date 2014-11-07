package engine

import (
	"fmt"
	"github.com/docker/docker/builtins"
	dockerEngine "github.com/docker/docker/engine"
	"github.com/krane-io/krane/api/server"
	"github.com/krane-io/krane/ssh"
	"github.com/krane-io/krane/types"
	"log"
	"net/url"
	"strconv"
)

func InitializeDockerEngine(configuration types.KraneConfiguration) (eng *dockerEngine.Engine) {
	eng = dockerEngine.New()

	eng.Hack_SetGlobalVar("configuration", configuration)

	// Load default plugins
	builtins.Register(eng)

	eng.Register("server_krane_api", server.ServeApi)
	eng.Register("ssh_tunnel", ssh.Tunnel)

	listenURL := &url.URL{
		Scheme: configuration.Production.Server.Host.Schema,
		Host:   configuration.Production.Server.Host.Fqdn + ":" + strconv.Itoa(configuration.Production.Server.Host.Port),
	}

	job := eng.Job("server_krane_api", listenURL.String())

	parameters := url.Values{}

	ships, err := configuration.Driver.List(parameters)
	if err != nil {
		log.Fatalf("unable to get list of ships from %s", configuration.Driver.Name())
	}

	configuration.UpdateShips(ships)
	eng.Hack_SetGlobalVar("configuration", configuration)

	for _, ship := range configuration.Production.Fleet {
		fmt.Printf("We are going to queue %s\n", ship.Fqdn)
		ssh_job := eng.Job("ssh_tunnel", ship.Fqdn, "false")
		ssh_job.Run()
	}

	job.SetenvBool("Logging", true)
	job.SetenvBool("AutoRestart", true)
	job.Setenv("ExecDriver", "native")

	job.SetenvBool("EnableCors", true)

	if err := job.Run(); err != nil {
		log.Fatalf("Unable to spawn the test daemon: %s", err)
	}

	return eng
}
