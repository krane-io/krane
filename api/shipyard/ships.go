package shipyard

import (
	"encoding/json"
	"github.com/docker/docker/engine"
	"github.com/krane-io/krane/types"
	"net/url"
)

func Plan(job *engine.Job) engine.Status {
	parameters := url.Values{}
	configuration := job.Eng.Hack_GetGlobalVar("configuration").(types.KraneConfiguration)

	if len(job.Args) > 0 {
		parameters.Set("name", job.Args[0])
	}

	plans, err := configuration.Driver.Plan(parameters)
	if err != nil {
		job.Errorf("unable to get list of ships from %s", configuration.Driver.Name())
	}

	jsonShip, err := json.Marshal(plans)

	if err != nil {
		job.Errorf("unable to marshal to json ship response")
	}

	job.Stdout.Write(jsonShip)
	return engine.StatusOK
}

func List(job *engine.Job) engine.Status {
	parameters := url.Values{}
	configuration := job.Eng.Hack_GetGlobalVar("configuration").(types.KraneConfiguration)

	ships, err := configuration.Driver.List(parameters)
	if err != nil {
		job.Errorf("unable to get list of ships from %s", configuration.Driver.Name())
	}

	jsonShip, err := json.Marshal(ships)
	if err != nil {
		job.Errorf("unable to marshal to json ship response")
	}

	job.Stdout.Write(jsonShip)
	return engine.StatusOK
}

func Decomission(job *engine.Job) engine.Status {
	var name string
	if len(job.Args) == 1 {
		name = job.Args[0]
	} else if len(job.Args) > 1 {
		return job.Errorf("Usage: %s", job.Name)
	}

	fqdn := job.Getenv("Fqdn")
	plan := job.Getenv("Plan")

	parameters := url.Values{}
	parameters.Set("name", name)
	parameters.Set("fqdn", fqdn)
	parameters.Set("plan", plan)

	configuration := job.Eng.Hack_GetGlobalVar("configuration").(types.KraneConfiguration)
	id, err := configuration.Driver.Create(parameters)

	if err != nil {
		job.Errorf("unable to create ship %s", fqdn)
	}

	job.Stdout.Write([]byte(id))

	newjob := job.Eng.Job("ssh_tunnel", fqdn, "true")
	go newjob.Run()

	return engine.StatusOK
}

func Commission(job *engine.Job) engine.Status {
	var name string
	if len(job.Args) == 1 {
		name = job.Args[0]
	} else if len(job.Args) > 1 {
		return job.Errorf("Usage: %s", job.Name)
	}

	fqdn := job.Getenv("Fqdn")
	plan := job.Getenv("Plan")

	parameters := url.Values{}
	parameters.Set("name", name)
	parameters.Set("fqdn", fqdn)
	parameters.Set("plan", plan)

	configuration := job.Eng.Hack_GetGlobalVar("configuration").(types.KraneConfiguration)
	id, err := configuration.Driver.Create(parameters)

	if err != nil {
		job.Errorf("unable to create ship %s", fqdn)
	}

	job.Stdout.Write([]byte(id))

	newjob := job.Eng.Job("ssh_tunnel", fqdn, "true")
	go newjob.Run()

	return engine.StatusOK
}
