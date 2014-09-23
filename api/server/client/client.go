package client

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
	dockerVersion "github.com/docker/docker/dockerversion"
	dockerEngine "github.com/docker/docker/engine"
	dockerUtils "github.com/docker/docker/utils"

	"github.com/krane-io/krane/types"
)

type KraneClientApi struct {
	ship                    types.Ship
	strictAPIversionControl bool
	job                     *dockerEngine.Job
	isTerminal              bool
	terminalFd              uintptr
}

func NewKraneClientApi(ship types.Ship, strict bool, job *dockerEngine.Job) *KraneClientApi {
	return &KraneClientApi{
		ship: ship,
		strictAPIversionControl: strict,
		job: job,
	}
}

var (
	ErrConnectionRefused = errors.New("Cannot connect to the Docker daemon. Is 'docker -d' running on this host?")
)

func (cli *KraneClientApi) HTTPClient() *http.Client {
	tr := &http.Transport{
		// TLSClientConfig: cli.tlsConfig,
		Dial: func(network, addr string) (net.Conn, error) {
			return net.DialTimeout(network, addr, 10*time.Second)
		},
	}
	return &http.Client{Transport: tr}
}

func (cli *KraneClientApi) Call(method, path string, data interface{}, passAuthInfo bool) (io.ReadCloser, int, error) {
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

	if cli.strictAPIversionControl {
		req, err = http.NewRequest(method, fmt.Sprintf("%s://%s:%d/v%s%s", cli.ship.Port, dockerApi.APIVERSION, path), params)
	} else {
		req, err = http.NewRequest(method, fmt.Sprintf("%s://%s:%d%s", cli.ship.Schema, "127.0.0.1", cli.ship.LocalPort, path), params)
	}

	cli.job.Logf("(%s) %s -> %s \n", method, req.URL.String(), fmt.Sprintf("%s://%s:%d%s", cli.ship.Schema, cli.ship.Fqdn, cli.ship.Port, path))

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
	resp, err := cli.HTTPClient().Do(req)

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

func (cli *KraneClientApi) Stream(method, path string, in io.Reader, out io.Writer, headers map[string][]string) error {
	return cli.streamHelper(method, path, true, in, out, nil, headers)
}

func (cli *KraneClientApi) streamHelper(method, path string, setRawTerminal bool, in io.Reader, stdout, stderr io.Writer, headers map[string][]string) error {
	if (method == "POST" || method == "PUT") && in == nil {
		in = bytes.NewReader([]byte{})
	}

	var req *http.Request
	var err error

	if cli.strictAPIversionControl {
		req, err = http.NewRequest(method, fmt.Sprintf("%s://%s:%d/v%s%s", cli.ship.Port, dockerApi.APIVERSION, path), in)
	} else {
		req, err = http.NewRequest(method, fmt.Sprintf("%s://%s:%d%s", cli.ship.Schema, "127.0.0.1", cli.ship.LocalPort, path), in)
	}

	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Docker-Client/"+dockerVersion.VERSION)

	if method == "POST" {
		req.Header.Set("Content-Type", "plain/text")
	}

	if headers != nil {
		for k, v := range headers {
			req.Header[k] = v
		}
	}
	resp, err := cli.HTTPClient().Do(req)
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("Cannot connect to the Docker daemon. Is 'docker -d' running on this host?")
		}
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		if len(body) == 0 {
			return fmt.Errorf("Error :%s", http.StatusText(resp.StatusCode))
		}
		return fmt.Errorf("Error: %s", bytes.TrimSpace(body))
	}

	if dockerApi.MatchesContentType(resp.Header.Get("Content-Type"), "application/json") {
		return dockerUtils.DisplayJSONMessagesStream(resp.Body, stdout, cli.terminalFd, cli.isTerminal)
	}
	if stdout != nil || stderr != nil {
		// When TTY is ON, use regular copy
		if setRawTerminal {
			_, err = io.Copy(stdout, resp.Body)
		} else {
			_, err = dockerUtils.StdCopy(stdout, stderr, resp.Body)
		}
		fmt.Printf("[stream] End of stdout") // TODO: Look into this
		return err
	}
	return nil
}
