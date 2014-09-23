package server

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"

	"strings"
	"time"

	dockerApi "github.com/docker/docker/api"
	dockerEngine "github.com/docker/docker/engine"
	"github.com/krane-io/krane/types"
)

func ErrConnectionRefused(host string) error {
	return errors.New(fmt.Sprintf("Cannot connect to the Docker daemon. Is 'docker -d' running on %s", host))
}

func Call(method string, ship types.Ship, path string, data interface{}, passAuthInfo bool) ([]byte, int, error) {
	return httpRequest(method, ship.Schema, ship.Fqdn, ship.Port, path, data, passAuthInfo)
}

func httpRequest(method, schema string, host string, port int, path string, data interface{}, passAuthInfo bool) ([]byte, int, error) {
	// Timer Function
	tr := &http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 300*time.Millisecond)
		},
	}

	client := &http.Client{Transport: tr}

	fmt.Printf("We are going to execute the following url: (%s)%s\n", method, fmt.Sprintf("%s://%s:%d%s", schema, host, port, path))

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

	fmt.Printf("\n\n%s\n\n", params)

	req, err := http.NewRequest(method, fmt.Sprintf("%s://%s:%d%s", schema, host, port, path), params)
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
	req.Header.Set("User-Agent", fmt.Sprintf("Docker-Client/%s", dockerApi.APIVERSION))
	if data != nil {
		req.Header.Set("Content-Type", "application/json")
	} else if method == "POST" {
		req.Header.Set("Content-Type", "plain/text")
	}

	resp, err := client.Do(req)

	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, -1, ErrConnectionRefused(host)
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

	if resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return nil, resp.StatusCode, err
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, err
	}

	return body, resp.StatusCode, nil
}

func readBody(stream io.ReadCloser, statusCode int, err error) ([]byte, int, error) {
	if stream != nil {
		defer stream.Close()
	}
	if err != nil {
		return nil, statusCode, err
	}
	body, err := ioutil.ReadAll(stream)
	if err != nil {
		return nil, -1, err
	}
	return body, statusCode, nil
}
