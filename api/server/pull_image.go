package server

import (
	"net/url"

	dockerEngine "github.com/docker/docker/engine"
	dockerPkgParser "github.com/docker/docker/pkg/parsers"

	"github.com/krane-io/krane/api/server/client"
	"github.com/krane-io/krane/types"
)

func pullImage(job *dockerEngine.Job, image string, ship types.Ship) error {
	v := url.Values{}
	repos, tag := dockerPkgParser.ParseRepositoryTag(image)
	// pull only the image tagged 'latest' if no tag was specified
	if tag == "" {
		tag = "latest"
	}
	v.Set("fromImage", repos)
	v.Set("tag", tag)

	// // Resolve the Repository name from fqn to hostname + name
	// hostname, _, err := registry.ResolveRepositoryName(repos)
	// if err != nil {
	// 	return err
	// }

	// // Load the auth config file, to be able to pull the image
	// cli.LoadConfigFile()

	// // Resolve the Auth config relevant for this server
	// authConfig := cli.configFile.ResolveAuthConfig(hostname)
	// buf, err := json.Marshal(authConfig)
	// if err != nil {
	// 	return err
	// }

	// registryAuthHeader := []string{
	// 	base64.URLEncoding.EncodeToString(buf),
	// }
	//

	cli := client.NewKraneClientApi(ship, false, job)

	if err := cli.Stream("POST", "/images/create?"+v.Encode(), nil, job.Stdout, nil); err != nil {
		return err
	}
	return nil

}
