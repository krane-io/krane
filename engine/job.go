package engine

import (
	dockerEngine "github.com/docker/docker/engine"
	"github.com/krane-io/krane/api/server"
)

func AttachJobs(eng *dockerEngine.Engine) error {
	for name, method := range map[string]dockerEngine.Handler{
		"containers":        server.Containers,
		"create":            server.ContainerCreate,
		"stop":              server.ContainerStop,
		"container_inspect": server.ContainerAttach,
		"start":             server.ContainerStart,
	} {
		if err := eng.Register(name, method); err != nil {
			return err
		}
	}
	return nil
}