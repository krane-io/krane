package server

import (
	"strings"

	dockerEngine "github.com/docker/docker/engine"

	"github.com/krane-io/krane/hacks"
	"github.com/krane-io/krane/types"
)

func shipWithContainerId(job *dockerEngine.Job, id string) *types.Ship {
	ships := listContainers(job)
	for _, ship := range ships {
		if ship != nil {
			for _, container := range ship.Containers {
				if strings.HasPrefix(container.ID, id) {
					configuration := hacks.DockerGetGlobalConfig(job)
					for _, shipClean := range configuration.Production.Fleet {
						if string(shipClean.Fqdn) == string(ship.Fqdn) {
							return &shipClean
						}
					}
				}
			}

		}
	}
	return nil
}
