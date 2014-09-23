package server

import (
	"fmt"
	"net/http"

	"github.com/krane-io/krane/types"

	dockerAPIServer "github.com/docker/docker/api/server"
	dockerEngine "github.com/docker/docker/engine"
	dockerPkgVersion "github.com/docker/docker/pkg/version"
)

var ServerRoutes = map[string]map[string]types.HttpApiFunc{
	"GET": {
		"/_ping":                          GetCatchAll,
		"/events":                         GetCatchAll,
		"/info":                           dockerAPIServer.GetInfo,
		"/version":                        GetCatchAll,
		"/images/json":                    GetCatchAll,
		"/images/viz":                     GetCatchAll,
		"/images/search":                  GetCatchAll,
		"/images/{name:.*}/get":           GetCatchAll,
		"/images/{name:.*}/history":       GetCatchAll,
		"/images/{name:.*}/json":          GetCatchAll,
		"/containers/ps":                  dockerAPIServer.GetContainersJSON,
		"/containers/json":                dockerAPIServer.GetContainersJSON,
		"/containers/{name:.*}/export":    GetCatchAll,
		"/containers/{name:.*}/changes":   GetCatchAll,
		"/containers/{name:.*}/json":      GetCatchAll,
		"/containers/{name:.*}/top":       GetCatchAll,
		"/containers/{name:.*}/logs":      GetCatchAll,
		"/containers/{name:.*}/attach/ws": GetCatchAll,
	},
	"POST": {
		"/auth":                         GetCatchAll,
		"/commit":                       GetCatchAll,
		"/build":                        GetCatchAll,
		"/images/create":                GetCatchAll,
		"/images/load":                  GetCatchAll,
		"/images/{name:.*}/push":        GetCatchAll,
		"/images/{name:.*}/tag":         GetCatchAll,
		"/containers/create":            dockerAPIServer.PostContainersCreate,
		"/containers/{name:.*}/kill":    GetCatchAll,
		"/containers/{name:.*}/pause":   GetCatchAll,
		"/containers/{name:.*}/unpause": GetCatchAll,
		"/containers/{name:.*}/restart": GetCatchAll,
		"/containers/{name:.*}/start":   dockerAPIServer.PostContainersStart,
		"/containers/{name:.*}/stop":    dockerAPIServer.PostContainersStop,
		"/containers/{name:.*}/wait":    GetCatchAll,
		"/containers/{name:.*}/resize":  GetCatchAll,
		"/containers/{name:.*}/attach":  dockerAPIServer.PostContainersAttach,
		"/containers/{name:.*}/copy":    GetCatchAll,
	},
	"DELETE": {
		"/containers/{name:.*}": GetCatchAll,
		"/images/{name:.*}":     GetCatchAll,
	},
	"OPTIONS": {
		"": GetCatchAll,
	},
}

func GetCatchAll(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	fmt.Printf("%#v", r)
	return nil
}
