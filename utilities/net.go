package utilities

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	dockerVersion "github.com/docker/docker/dockerversion"
	dockerEngine "github.com/docker/docker/engine"
)

var (
	ErrConnectionRefused = errors.New("Cannot connect to the Docker daemon. Is 'docker -d' running on this host?")
)

func HTTPClient() *http.Client {
	tr := &http.Transport{
		// TLSClientConfig: cli.tlsConfig,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 10*time.Second)
		},
	}
	return &http.Client{Transport: tr}
}

func Call(method, path string, data interface{}, passAuthInfo bool) (io.ReadCloser, int, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if env, ok := data.(dockerEngine.Env); ok {
			if err := env.Encode(params); err != nil {
				return nil, -1, err
			}
		} else {
			buf, err := json.Marshal(data)
			if err != nil {
				return nil, -1, err
			}
			if _, err := params.Write(buf); err != nil {
				return nil, -1, err
			}
		}
	}

	var req *http.Request
	var err error

	req, err = http.NewRequest(method, path, params)

	log.Printf("(%s) %s -> %s \n", method, req.URL.String(), path)

	if err != nil {
		return nil, -1, err
	}
	// if passAuthInfo {
	// 	cli.LoadConfigFile()
	// 	// Resolve the Auth config relevant for this server
	// 	authConfig := cli.configFile.ResolveAuthConfig(registry.IndexServerAddress())
	// 	getHeaders := func(authConfig registry.AuthConfig) (map[string][]string, error) {
	// 		buf, err := json.Marshal(authConfig)
	// 		if err != nil {
	// 			return nil, err
	// 		}
	// 		registryAuthHeader := []string{
	// 			base64.URLEncoding.EncodeToString(buf),
	// 		}
	// 		return map[string][]string{"X-Registry-Auth": registryAuthHeader}, nil
	// 	}
	// 	if headers, err := getHeaders(authConfig); err == nil && headers != nil {
	// 		for k, v := range headers {
	// 			req.Header[k] = v
	// 		}
	// 	}
	// }
	req.Header.Set("User-Agent", "Docker-Client/"+dockerVersion.VERSION)
	// req.URL.Host = cli.ship.Fqdn
	// req.URL.Scheme = cli.ship.Schema
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	} else if method == "POST" {
		req.Header.Set("Content-Type", "plain/text")
	}
	resp, err := HTTPClient().Do(req)

	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, -1, ErrConnectionRefused
		}
		return nil, -1, err
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, -1, err
		}
		if len(body) == 0 {
			return nil, resp.StatusCode, fmt.Errorf("Error: request returned %s for API route and version %s, check if the server supports the requested API version", http.StatusText(resp.StatusCode), req.URL)
		}
		return nil, resp.StatusCode, fmt.Errorf("Error response from daemon: %s", bytes.TrimSpace(body))
	}
	return resp.Body, resp.StatusCode, nil
}
