package hacks

import(	
	"encoding/json"
	"github.com/krane-io/krane/types"
	dockerEngine "github.com/docker/docker/engine"
)

func DockerSetGlobalConfig(job *dockerEngine.Job, configuration types.KraneConfiguration) error {
	value, err := json.Marshal(configuration)
	if err != nil {
		return err
	}
	job.Eng.Hack_SetGlobalVar("KraneConfiguration", value)
	return nil
}

func DockerGetGlobalConfig(job *dockerEngine.Job) types.KraneConfiguration {
	sval := job.Eng.Hack_GetGlobalVar("KraneConfiguration").([]byte)

	var configuration types.KraneConfiguration
	json.Unmarshal(sval, &configuration)
	return configuration
}