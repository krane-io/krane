package engine

import (
	"fmt"
	"log"
	"net/url"
	"strconv"

	"github.com/krane-io/krane/hacks"
	"github.com/krane-io/krane/ssh"
	"github.com/krane-io/krane/types"

	dockerBuiltins "github.com/docker/docker/builtins"
	dockerEngine "github.com/docker/docker/engine"
)

func InitializeDockerEngine(configuration types.KraneConfiguration) (eng *dockerEngine.Engine) {
	eng = dockerEngine.New()
	// Load default plugins
	dockerBuiltins.Register(eng)

	eng.Register("server_krane_api", ServeApi)
	eng.Register("ssh_tunnel", ssh.Tunnel)

	listenURL := &url.URL{
		Scheme: configuration.Production.Server.Host.Schema,
		Host:   configuration.Production.Server.Host.Fqdn + ":" + strconv.Itoa(configuration.Production.Server.Host.Port),
	}

	job := eng.Job("server_krane_api", listenURL.String())

	fmt.Printf("%v", configuration)

	hacks.DockerSetGlobalConfig(job, configuration)

	ssh_job := eng.Job("ssh_tunnel")
	ssh_job.Run()

	job.SetenvBool("Logging", true)
	job.SetenvBool("AutoRestart", true)
	job.Setenv("ExecDriver", "native")

	job.SetenvBool("EnableCors", true)

	if err := job.Run(); err != nil {
		log.Fatalf("Unable to spawn the test daemon: %s", err)
	}

	return eng
}
