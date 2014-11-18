package server

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/pkg/version"
	"github.com/krane-io/krane/api/shipyard"
	"net/http"
)

func AttachJobs(eng *engine.Engine) error {
	for name, method := range map[string]engine.Handler{
		"containers":        Containers,
		"ships":             shipyard.List,
		"plans":             shipyard.Plan,
		"commission":        shipyard.Commission,
		"decomission":       shipyard.Decomission,
		"create":            ContainerCreate,
		"stop":              ContainerStop,
		"container_inspect": ContainerAttach,
		"start":             ContainerStart,
	} {
		if err := eng.Register(name, method); err != nil {
			return err
		}
	}
	return nil
}

var ServerRoutes = map[string]map[string]HttpApiFunc{
	"GET": {
		"/_ping":                          GetCatchAll,
		"/events":                         GetCatchAll,
		"/info":                           getInfo,
		"/version":                        GetCatchAll,
		"/images/json":                    GetCatchAll,
		"/images/viz":                     GetCatchAll,
		"/images/search":                  GetCatchAll,
		"/images/{name:.*}/get":           GetCatchAll,
		"/images/{name:.*}/history":       GetCatchAll,
		"/images/{name:.*}/json":          GetCatchAll,
		"/containers/ps":                  getContainersJSON,
		"/containers/json":                getContainersJSON,
		"/containers/{name:.*}/export":    GetCatchAll,
		"/containers/{name:.*}/changes":   GetCatchAll,
		"/containers/{name:.*}/json":      GetCatchAll,
		"/containers/{name:.*}/top":       GetCatchAll,
		"/containers/{name:.*}/logs":      GetCatchAll,
		"/containers/{name:.*}/attach/ws": GetCatchAll,
		"/ships/json":                     GetShipsJSON,
		"/plans/json":                     GetPlansJSON,
	},
	"POST": {
		"/auth":                         GetCatchAll,
		"/commit":                       GetCatchAll,
		"/build":                        GetCatchAll,
		"/images/create":                GetCatchAll,
		"/images/load":                  GetCatchAll,
		"/images/{name:.*}/push":        GetCatchAll,
		"/images/{name:.*}/tag":         GetCatchAll,
		"/containers/create":            PostContainersCreate,
		"/ships/create":                 PostShipsCreate,
		"/containers/{name:.*}/kill":    GetCatchAll,
		"/containers/{name:.*}/pause":   GetCatchAll,
		"/containers/{name:.*}/unpause": GetCatchAll,
		"/containers/{name:.*}/restart": GetCatchAll,
		"/containers/{name:.*}/start":   postContainersStart,
		"/containers/{name:.*}/stop":    postContainersStop,
		"/containers/{name:.*}/wait":    GetCatchAll,
		"/containers/{name:.*}/resize":  GetCatchAll,
		"/containers/{name:.*}/attach":  postContainersAttach,
		"/containers/{name:.*}/copy":    GetCatchAll,
	},
	"DELETE": {
		"/containers/{name:.*}": GetCatchAll,
		"/images/{name:.*}":     GetCatchAll,
		"/ships/{name:.*}":      DeleteShips,
	},
	"OPTIONS": {
		"": GetCatchAll,
	},
}

func GetCatchAll(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	fmt.Printf("%#v", r)
	return nil
}

func PostContainersCreate(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return nil
	}
	var (
		out          engine.Env
		job          = eng.Job("create", r.Form.Get("name"), r.Form.Get("ship"))
		outWarnings  []string
		stdoutBuffer = bytes.NewBuffer(nil)
		warnings     = bytes.NewBuffer(nil)
	)

	if err := checkForJson(r); err != nil {
		return err
	}

	if err := job.DecodeEnv(r.Body); err != nil {
		return err
	}
	// Read container ID from the first line of stdout
	job.Stdout.Add(stdoutBuffer)
	// Read warnings from stderr
	job.Stderr.Add(warnings)
	if err := job.Run(); err != nil {
		return err
	}
	// Parse warnings from stderr
	scanner := bufio.NewScanner(warnings)
	for scanner.Scan() {
		outWarnings = append(outWarnings, scanner.Text())
	}
	out.Set("Id", engine.Tail(stdoutBuffer, 1))
	out.SetList("Warnings", outWarnings)

	return writeJSON(w, http.StatusCreated, out)
}

func DeleteShips(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}

	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}

	job := eng.Job("decomission", r.Form.Get("name"))

	job.Setenv("time", r.Form.Get("time"))

	if err := job.Run(); err != nil {
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}
func GetPlansJSON(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}

	var (
		err error
		job = eng.Job("plans", r.Form.Get("name"))
	)

	streamJSON(job, w, false)

	if err = job.Run(); err != nil {
		return err
	}
	return nil
}

func GetShipsJSON(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
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

func PostShipsCreate(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
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
