package server

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/engine"
	"github.com/krane-io/krane/api/server/client"
	"github.com/krane-io/krane/runconfig"
	"github.com/krane-io/krane/types"
	"net/url"
	"strconv"
)

func Containers(job *engine.Job) engine.Status {
	ships := listContainers(job)
	jsonShip, _ := json.Marshal(ships)
	job.Stdout.Write(jsonShip)
	return engine.StatusOK
}

func ContainerStart(job *engine.Job) engine.Status {
	hostConfig := runconfig.ContainerHostConfigFromJob(job)
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s CONTAINER\n", job.Name)
	}
	var (
		name = job.Args[0]
	)

	v := url.Values{}

	if job.EnvExists("t") {
		v.Set("t", strconv.Itoa(job.GetenvInt("t")))
	}

	job.SetenvBool("all", true)

	ship := shipWithContainerId(job, name)

	cli := client.NewKraneClientApi(*ship, false, job)

	if ship != nil {
		body, statusCode, err := cli.Call("POST", "/containers/"+name+"/start?"+v.Encode(), hostConfig, false)
		if statusCode == 304 {
			return job.Errorf("Container already started")
		} else if statusCode == 404 {
			return job.Errorf("No such container: %s\n", name)
		} else if statusCode == 404 {
			return job.Errorf("Cannot start container %s: %s\n", name, err)
		} else if (statusCode >= 200) && (statusCode < 300) {
			fmt.Printf("%s", body)
		} else {
			return job.Errorf("Cannot start container %s: %s\n", name, err)
		}
	} else {
		return job.Errorf("No such container: %s\n", name)
	}

	return engine.StatusOK
}

func ContainerStop(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s CONTAINER\n", job.Name)
	}
	var (
		name = job.Args[0]
	)
	v := url.Values{}

	if job.EnvExists("t") {
		v.Set("t", strconv.Itoa(job.GetenvInt("t")))
	}

	ship := shipWithContainerId(job, name)

	cli := client.NewKraneClientApi(*ship, false, job)

	if ship != nil {
		body, statusCode, err := readBody(cli.Call("POST", "/containers/"+name+"/stop?"+v.Encode(), nil, false))
		if statusCode == 304 {
			return job.Errorf("Container already stopped")
		} else if statusCode == 404 {
			return job.Errorf("No such container: %s\n", name)
		} else if statusCode == 404 {
			return job.Errorf("Cannot stop container %s: %s\n", name, err)
		} else if (statusCode >= 200) && (statusCode < 300) {
			fmt.Printf("%s", body)
		} else {
			return job.Errorf("Cannot stop container %s: %s\n", name, err)
		}
	} else {
		return job.Errorf("No such container: %s\n", name)
	}
	return engine.StatusOK
}

func ContainerAttach(job *engine.Job) engine.Status {
	// if len(job.Args) == 1 {
	// 	containerValues.Set("name", job.Args[0])
	// }

	// var (
	// 	name   = job.Args[0]
	// 	logs   = job.GetenvBool("logs")
	// 	stream = job.GetenvBool("stream")
	// 	stdin  = job.GetenvBool("stdin")
	// 	stdout = job.GetenvBool("stdout")
	// 	stderr = job.GetenvBool("stderr")
	// )
	return engine.StatusOK
}

func ContainerCreate(job *engine.Job) engine.Status {
	configuration := job.Eng.Hack_GetGlobalVar("configuration").(types.KraneConfiguration)
	config := runconfig.ContainerConfigFromJob(job)

	ship := configuration.GetShip(config.Ship)

	if len(job.Args) != 1 || ship.Fqdn == "" {
		return job.Errorf("Usage: %s CONTAINER\n", job.Name)
	}

	containerValues := url.Values{}
	containerValues.Set("name", job.Args[0])

	cli := client.NewKraneClientApi(ship, false, job)

	stream, statusCode, err := cli.Call("POST", "/containers/create?"+containerValues.Encode(), config, false)

	if statusCode == 404 {
		job.Printf("Unable to find image '%s' in %s://%s:%d\n", config.Image, ship.Schema, ship.Fqdn, ship.Port)

		if err = pullImage(job, config.Image, ship); err != nil {
			return job.Errorf("Cannot pull image %s: %s\n", config.Image, err)
		}
		if stream, _, err = cli.Call("POST", "/containers/create?"+containerValues.Encode(), config, false); err != nil {
			return job.Errorf("Cannot create container: %s\n", err)
		}
	} else if err != nil {
		return job.Errorf("Cannot create container: %s\n", err)
	}

	var runResult engine.Env

	if err := runResult.Decode(stream); err != nil {
		return job.Errorf("Error with container: %s\n", err)
	}

	for _, warning := range runResult.GetList("Warnings") {
		job.Stdout.Write([]byte(fmt.Sprintf("WARNING: %s\n", warning)))
	}

	job.Stdout.Write([]byte(runResult.Get("Id")))

	return engine.StatusOK
}

func ContainerRestart(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s CONTAINER\n", job.Name)
	}
	var (
		name = job.Args[0]
	)
	v := url.Values{}

	if job.EnvExists("t") {
		v.Set("t", strconv.Itoa(job.GetenvInt("t")))
	}

	ship := shipWithContainerId(job, name)

	if ship != nil {
		_, statusCode, err := httpRequest("POST", "http", ship.Fqdn, ship.Port, "/containers/"+name+"/restart?"+v.Encode(), nil, false)
		if statusCode == 404 {
			return job.Errorf("No such container: %s\n", name)
		} else if statusCode == 500 {
			return job.Errorf("Server error trying to restart %s: %s\n", name, err)
		} else if (statusCode >= 200) && (statusCode < 300) {
			fmt.Printf("Container %s restarted\n", name)
		} else {
			return job.Errorf("Cannot restart container %s: %s\n", name, err)
		}
	} else {
		return job.Errorf("No such container: %s\n", name)
	}
	return engine.StatusOK
}

func ContainerDelete(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Not enough arguments. Usage: %s CONTAINER\n", job.Name)
	}

	name := job.Args[0]
	removeVolume := job.GetenvBool("removeVolume")
	removeLink := job.GetenvBool("removeLink")
	stop := job.GetenvBool("stop")
	kill := job.GetenvBool("kill")

	val := url.Values{}
	if removeVolume {
		val.Set("v", "1")
	}
	if removeLink {
		val.Set("link", "1")
	}
	if stop {
		val.Set("stop", "1")
	}
	if kill {
		val.Set("kill", "1")
	}

	ship := shipWithContainerId(job, name)

	if ship != nil {
		_, statusCode, err := httpRequest("DELETE", "http", ship.Fqdn, ship.Port, "/containers/"+name+"?"+val.Encode(), nil, false)
		if statusCode == 404 {
			return job.Errorf("No such container: %s\n", name)
		} else if statusCode == 400 {
			return job.Errorf("Bad parameter : %s\n", err)
		} else if (statusCode >= 200) && (statusCode < 300) {
			fmt.Printf("Container %s deleted\n", name)
		} else {
			return job.Errorf("Could not remove container %s: %s\n", name, err)
		}
	} else {
		return job.Errorf("No such container: %s\n", name)
	}
	return engine.StatusOK
}

func ContainerPause(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s CONTAINER", job.Name)
	}

	name := job.Args[0]
	ship := shipWithContainerId(job, name)

	if ship != nil {
		_, statusCode, err := httpRequest("POST", "http", ship.Fqdn, ship.Port, "/containers/"+name+"/pause", nil, false)
		if statusCode == 404 {
			return job.Errorf("No such container: %s\n", name)
		} else if statusCode == 500 {
			return job.Errorf("Could not pause %s. Server error : %s\n", name, err)
		} else if (statusCode >= 200) && (statusCode < 300) {
			fmt.Printf("Container %s paused\n", name)
		} else {
			return job.Errorf("Could not pause container %s: %s\n", name, err)
		}
	} else {
		return job.Errorf("No such container: %s\n", name)
	}
	return engine.StatusOK
}

func ContainerUnpause(job *engine.Job) engine.Status {
	if len(job.Args) != 1 {
		return job.Errorf("Usage: %s CONTAINER", job.Name)
	}

	name := job.Args[0]
	ship := shipWithContainerId(job, name)

	if ship != nil {
		_, statusCode, err := httpRequest("POST", "http", ship.Fqdn, ship.Port, "/containers/"+name+"/unpause", nil, false)
		if statusCode == 404 {
			return job.Errorf("No such container: %s\n", name)
		} else if statusCode == 500 {
			return job.Errorf("Could not unpause %s. Server error : %s\n", name, err)
		} else if (statusCode >= 200) && (statusCode < 300) {
			fmt.Printf("Container %s unpaused\n", name)
		} else {
			return job.Errorf("Could not unpause container %s: %s\n", name, err)
		}
	} else {
		return job.Errorf("No such container: %s\n", name)
	}
	return engine.StatusOK
}
