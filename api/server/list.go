package server

import (
	"encoding/json"
	"net/url"
	"strconv"

	dockerEngine "github.com/docker/docker/engine"

	"github.com/krane-io/krane/api/server/client"

	"github.com/krane-io/krane/config"
	"github.com/krane-io/krane/types"
)

func listContainer(job *dockerEngine.Job, configuration config.KraneConfiguration) <-chan *types.APIShip {
	v := url.Values{}
	if all := job.GetenvBool("all"); all {
		v.Set("all", strconv.FormatBool(all))
	}

	ch := make(chan *types.APIShip, len(configuration.Production.Fleet))
	for _, ship := range configuration.Production.Fleet {
		go func(ship types.Ship) {

			cli := client.NewKraneClientApi(ship, false, job)

			body, statusCode, err := readBody(cli.Call("GET", "/containers/json?"+v.Encode(), nil, false))

			job.Logf("(%d) %s\n", statusCode, body)

			if err != nil {
				job.Logf("Error: %s", err.Error())
			}
			var resultShip types.APIShip
			if (statusCode >= 200) && (statusCode < 300) {
				var containerList []types.APIContainers
				json.Unmarshal(body, &containerList)
				resultShip.Name = ship.Name
				resultShip.Fqdn = ship.Fqdn
				resultShip.Port = ship.Port
				resultShip.Status = true
				resultShip.Containers = containerList
				ch <- &resultShip
			} else {
				ch <- nil
			}
		}(ship)
	}
	return ch
}

func listContainers(job *dockerEngine.Job) []*types.APIShip {
	configuration := job.Eng.Hack_GetGlobalVar("configuration").(config.KraneConfiguration)
	results := listContainer(job, configuration)
	nShips := len(configuration.Production.Fleet)
	ships := make([]*types.APIShip, 0, nShips)
	for i := 0; i < nShips; i++ {
		result := <-results
		if result != nil {
			ships = append(ships, result)
		}
	}
	return ships
}
