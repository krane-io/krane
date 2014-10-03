package main

import (
	"fmt"
	"github.com/krane-io/krane/api/client"
	"github.com/krane-io/krane/config"
	"github.com/krane-io/krane/drivers/concerto"
	"github.com/krane-io/krane/engine"
	"log"
	"os"
	"strings"
	"unsafe"

	dockerAPI "github.com/docker/docker/api"
	flag "github.com/docker/docker/pkg/mflag"
	dockerUtils "github.com/docker/docker/utils"
)

func main() {
	configuration := config.LoadConfigurationFile()

	flag.Parse()

	if *flVersion {
		showVersion()
		return
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}

	if flHosts.Len() == 0 {
		defaultHost := os.Getenv("DOCKER_HOST")

		if os.Getenv("KRANE_HOST") != "" {
			defaultHost = os.Getenv("KRANE_HOST")
		}

		if unsafe.Sizeof(configuration) > 0 {
			defaultHost = fmt.Sprintf("%s:%d", configuration.Production.Server.Host.Fqdn, configuration.Production.Server.Host.Port)
		} else if defaultHost == "" || *flDaemon {
			// If we do not have a host, default to unix socket
			defaultHost = fmt.Sprintf("unix://%s", dockerAPI.DEFAULTUNIXSOCKET)
		}
		if _, err := dockerAPI.ValidateHost(defaultHost); err != nil {
			log.Fatal(err)
		}
		flHosts.Set(defaultHost)
	}

	if *flDaemon {

		if configuration.Production.Server.Driver == "concerto" {
			concerto := concerto.NewDriver()
			configuration.Driver = concerto
		}
		engine.InitializeDockerEngine(configuration)
		return
	}

	if flHosts.Len() > 1 {
		log.Fatal("Please specify only one -H")
	}

	protoAddrParts := strings.SplitN(flHosts.GetAll()[0], "://", 2)

	var (
		cli *client.KraneCli
	)

	cli = client.NewKraneCli(os.Stdin, os.Stdout, os.Stderr, protoAddrParts[0], protoAddrParts[1], nil)

	if err := cli.Cmd(flag.Args()...); err != nil {
		if sterr, ok := err.(*dockerUtils.StatusError); ok {
			if sterr.Status != "" {
				log.Println(sterr.Status)
			}
			os.Exit(sterr.StatusCode)
		}
		log.Fatal(err)
	}

}

// TODO: We need to do version similar to https://github.com/docker/docker/blob/master/dockerversion/dockerversion.go
func showVersion() {
	fmt.Printf("Krane version .......\n")
}
