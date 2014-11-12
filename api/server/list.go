package server

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/engine"
	"github.com/krane-io/krane/api/server/client"
	"github.com/krane-io/krane/types"
	"net/url"
	"strconv"
)

func listContainer(job *engine.Job, configuration types.KraneConfiguration) <-chan *types.Ship {
	v := url.Values{}
	if all := job.GetenvBool("all"); all {
		v.Set("all", strconv.FormatBool(all))
	}

	ch := make(chan *types.Ship, len(configuration.Production.Fleet.Ships()))
	for _, ship := range configuration.Production.Fleet.Available() {
		go func(ship types.Ship) {

			cli := client.NewKraneClientApi(ship, false, job)
			body, statusCode, err := readBody(cli.Call("GET", "/containers/json?"+v.Encode(), nil, false))

			job.Logf("(%d) %s\n", statusCode, body)

			if err != nil {
				job.Logf("Error: %s", err.Error())
			}
			var resultShip types.Ship
			if (statusCode >= 200) && (statusCode < 300) {
				var containerList []types.Containers
				json.Unmarshal(body, &containerList)
				resultShip.Name = ship.Name
				resultShip.Fqdn = ship.Fqdn
				resultShip.Port = ship.Port
				resultShip.State = "operational"
				resultShip.Containers = containerList
				fmt.Printf("%#v", resultShip)
				ch <- &resultShip
			} else {
				ch <- nil
			}
		}(ship)
	}
	return ch
}

func listContainers(job *engine.Job) []*types.Ship {
	configuration := job.Eng.Hack_GetGlobalVar("configuration").(types.KraneConfiguration)
	results := listContainer(job, configuration)
	nShips := len(configuration.Production.Fleet.Available())
	ships := make([]*types.Ship, 0, nShips)
	for i := 0; i < nShips; i++ {
		result := <-results
		if result != nil {
			ships = append(ships, result)
		}
	}
	return ships
}
