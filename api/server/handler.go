package server

import (
	"fmt"
	"net/http"

	"github.com/krane-io/krane/types"
	"strings"

	dockerAPI "github.com/docker/docker/api"
	dockerAPIServer "github.com/docker/docker/api/server"
	dockerEngine "github.com/docker/docker/engine"
	dockerPkgVersion "github.com/docker/docker/pkg/version"
	dockerUtils "github.com/docker/docker/utils"
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
		"/ships/json":                     GetShipsJSON,
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
		"/ships/create":                 PostShipsCreate,
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

func GetShipsJSON(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	var (
		err error
		job = eng.Job("ships")
	)

	streamJSON(job, w, false)

	if err = job.Run(); err != nil {
		return err
	}
	return nil
}

func PostShipsCreate(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}

	var (
		job = eng.Job("commission", r.Form.Get("name"))
	)

	if err := checkForJson(r); err != nil {
		return err
	}

	if err := job.DecodeEnv(r.Body); err != nil {
		return err
	}

	streamJSON(job, w, false)

	if err := job.Run(); err != nil {
		return err
	}
	return nil
}

func streamJSON(job *dockerEngine.Job, w http.ResponseWriter, flush bool) {
	w.Header().Set("Content-Type", "application/json")
	if job.GetenvBool("lineDelim") {
		w.Header().Set("Content-Type", "application/x-json-stream")
	}

	if flush {
		job.Stdout.Add(dockerUtils.NewWriteFlusher(w))
	} else {
		job.Stdout.Add(w)
	}
}

//If we don't do this, POST method without Content-type (even with empty body) will fail
func parseForm(r *http.Request) error {
	if r == nil {
		return nil
	}
	if err := r.ParseForm(); err != nil && !strings.HasPrefix(err.Error(), "mime:") {
		return err
	}
	return nil
}

// Check to make sure request's Content-Type is application/json
func checkForJson(r *http.Request) error {
	ct := r.Header.Get("Content-Type")

	// No Content-Type header is ok as long as there's no Body
	if ct == "" {
		if r.Body == nil || r.ContentLength == 0 {
			return nil
		}
	}

	// Otherwise it better be json
	if dockerAPI.MatchesContentType(ct, "application/json") {
		return nil
	}
	return fmt.Errorf("Content-Type specified (%s) must be 'application/json'", ct)
}
