package server

import (
	"github.com/docker/docker/engine"
	"github.com/krane-io/krane/types"
	"strings"
)

func shipWithContainerId(job *engine.Job, id string) *types.Ship {
	ships := listContainers(job)
	for _, ship := range ships {
		if ship != nil {
			for _, container := range ship.Containers {
				if strings.HasPrefix(container.ID, id) {
					configuration := job.Eng.Hack_GetGlobalVar("configuration").(types.KraneConfiguration)
					for _, shipClean := range configuration.Production.Fleet.Ships() {
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
