package types

import(
	"net/http"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/pkg/version"
)

type HttpApiFunc func(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error