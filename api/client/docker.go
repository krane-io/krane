// Code generated because docker has made most of the methods of docker api
// client not publicly available.
package client

import (
	"bytes"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"

	dockerApi "github.com/docker/docker/api"
	dockerApiClient "github.com/docker/docker/api/client"
	dockerVersion "github.com/docker/docker/dockerversion"
	dockerEngine "github.com/docker/docker/engine"
	flag "github.com/docker/docker/pkg/mflag"
	dockerPkgTerm "github.com/docker/docker/pkg/term"
	dockerRegistry "github.com/docker/docker/registry"
	dockerLibtrust "github.com/docker/libtrust"
)

var (
	ErrConnectionRefused = errors.New("Cannot connect to the Krane daemon. Is 'krane -d' running on this host?")
)

func (cli *KraneCli) Subcmd(name, signature, description string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintf(cli.err, "\nUsage: krane %s %s\n\n%s\n\n", name, signature, description)
		flags.PrintDefaults()
		os.Exit(2)
	}
	return flags
}

func (cli *KraneCli) getMethod(name string) (func(...string) error, bool) {
	if len(name) == 0 {
		return nil, false
	}
	methodName := "Cmd" + strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	method := reflect.ValueOf(cli).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}

// Cmd executes the specified command
func (cli *KraneCli) Cmd(args ...string) error {
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Println("Error: Command not found:", args[0])
			return cli.CmdHelp(args[1:]...)
		}
		return method(args[1:]...)
	}
	return cli.CmdHelp(args...)
}

func NewKraneCli(in io.ReadCloser, out, err io.Writer, key dockerLibtrust.PrivateKey, proto, addr string, tlsConfig *tls.Config) *KraneCli {
	var (
		inFd          uintptr
		outFd         uintptr
		isTerminalIn  = false
		isTerminalOut = false
		scheme        = "http"
	)

	if tlsConfig != nil {
		scheme = "https"
	}

	if in != nil {
		if file, ok := in.(*os.File); ok {
			inFd = file.Fd()
			isTerminalIn = dockerPkgTerm.IsTerminal(inFd)
		}
	}

	if out != nil {
		if file, ok := out.(*os.File); ok {
			outFd = file.Fd()
			isTerminalOut = dockerPkgTerm.IsTerminal(outFd)
		}
	}

	if err == nil {
		err = out
	}

	// The transport is created here for reuse during the client session
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
		Dial: func(dial_network, dial_addr string) (net.Conn, error) {
			// Why 32? See issue 8035
			return net.DialTimeout(proto, addr, 32*time.Second)
		},
	}
	if proto == "unix" {
		// no need in compressing for local communications
		tr.DisableCompression = true
	}

	return &KraneCli{
		*dockerApiClient.NewDockerCli(in, out, err, key, proto, addr, tlsConfig),
		proto,
		addr,
		nil,
		in,
		out,
		err,
		key,
		tlsConfig,
		scheme,
		inFd,
		outFd,
		isTerminalIn,
		isTerminalOut,
		nil,
	}
}

type KraneCli struct {
	dockerApiClient.DockerCli
	proto      string
	addr       string
	configFile *dockerRegistry.ConfigFile
	in         io.ReadCloser
	out        io.Writer
	err        io.Writer
	key        dockerLibtrust.PrivateKey
	tlsConfig  *tls.Config
	scheme     string
	// inFd holds file descriptor of the client's STDIN, if it's a valid file
	inFd uintptr
	// outFd holds file descriptor of the client's STDOUT, if it's a valid file
	outFd uintptr
	// isTerminalIn describes if client's STDIN is a TTY
	isTerminalIn bool
	// isTerminalOut describes if client's STDOUT is a TTY
	isTerminalOut bool
	transport     *http.Transport
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

func (cli *KraneCli) encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if env, ok := data.(dockerEngine.Env); ok {
			if err := env.Encode(params); err != nil {
				return nil, err
			}
		} else {
			buf, err := json.Marshal(data)
			if err != nil {
				return nil, err
			}
			if _, err := params.Write(buf); err != nil {
				return nil, err
			}
		}
	}
	return params, nil
}

func (cli *KraneCli) call(method, path string, data interface{}, passAuthInfo bool) (io.ReadCloser, int, error) {
	params, err := cli.encodeData(data)
	if err != nil {
		return nil, -1, err
	}
	req, err := http.NewRequest(method, fmt.Sprintf("/v%s%s", dockerApi.APIVERSION, path), params)
	if err != nil {
		return nil, -1, err
	}
	if passAuthInfo {
		cli.LoadConfigFile()
		// Resolve the Auth config relevant for this server
		authConfig := cli.configFile.ResolveAuthConfig(dockerRegistry.IndexServerAddress())
		getHeaders := func(authConfig dockerRegistry.AuthConfig) (map[string][]string, error) {
			buf, err := json.Marshal(authConfig)
			if err != nil {
				return nil, err
			}
			registryAuthHeader := []string{
				base64.URLEncoding.EncodeToString(buf),
			}
			return map[string][]string{"X-Registry-Auth": registryAuthHeader}, nil
		}
		if headers, err := getHeaders(authConfig); err == nil && headers != nil {
			for k, v := range headers {
				req.Header[k] = v
			}
		}
	}
	req.Header.Set("User-Agent", "Docker-Client/"+dockerVersion.VERSION)
	req.URL.Host = cli.addr
	req.URL.Scheme = cli.scheme
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
