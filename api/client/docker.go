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
	"net/http/httputil"
	"net/url"
	"os"
	"reflect"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	flag "github.com/docker/docker/pkg/mflag"
	gosignal "os/signal"

	"github.com/docker/docker/api"
	"github.com/docker/docker/api/client"
	"github.com/docker/docker/dockerversion"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/pkg/log"

	"github.com/docker/docker/pkg/parsers"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/term"
	"github.com/docker/docker/registry"
	"github.com/docker/docker/runconfig"
	"github.com/docker/docker/utils"
	"github.com/docker/libtrust"
)

var (
	ErrConnectionRefused = errors.New("Cannot connect to the Krane daemon. Is 'krane -d' running on this host?")
)

func (cli *KraneCli) dial() (net.Conn, error) {
	if cli.tlsConfig != nil && cli.proto != "unix" {
		return tls.Dial(cli.proto, cli.addr, cli.tlsConfig)
	}
	return net.Dial(cli.proto, cli.addr)
}

type cidFile struct {
	path    string
	file    *os.File
	written bool
}

func (cid *cidFile) Close() error {
	cid.file.Close()

	if !cid.written {
		if err := os.Remove(cid.path); err != nil {
			return fmt.Errorf("failed to remove the CID file '%s': %s \n", cid.path, err)
		}
	}

	return nil
}

func (cid *cidFile) Write(id string) error {
	if _, err := cid.file.Write([]byte(id)); err != nil {
		return fmt.Errorf("Failed to write the container ID to the file: %s", err)
	}
	cid.written = true
	return nil
}

func newCIDFile(path string) (*cidFile, error) {
	if _, err := os.Stat(path); err == nil {
		return nil, fmt.Errorf("Container ID file found, make sure the other container isn't running or delete %s", path)
	}
	f, err := os.Create(path)
	if err != nil {
		return nil, fmt.Errorf("Failed to create the container ID file: %s", err)
	}

	return &cidFile{path: path, file: f}, nil
}

func (cli *KraneCli) stream(method, path string, in io.Reader, out io.Writer, headers map[string][]string) error {
	return cli.streamHelper(method, path, true, in, out, nil, headers)
}

func (cli *KraneCli) forwardAllSignals(cid string) chan os.Signal {
	sigc := make(chan os.Signal, 128)
	signal.CatchAll(sigc)
	go func() {
		for s := range sigc {
			if s == syscall.SIGCHLD {
				continue
			}
			var sig string
			for sigStr, sigN := range signal.SignalMap {
				if sigN == s {
					sig = sigStr
					break
				}
			}
			if sig == "" {
				log.Errorf("Unsupported signal: %d. Discarding.", s)
			}
			if _, _, err := readBody(cli.call("POST", fmt.Sprintf("/containers/%s/kill?signal=%s", cid, sig), nil, false)); err != nil {
				log.Debugf("Error sending signal: %s", err)
			}
		}
	}()
	return sigc
}

func (cli *KraneCli) streamHelper(method, path string, setRawTerminal bool, in io.Reader, stdout, stderr io.Writer, headers map[string][]string) error {
	if (method == "POST" || method == "PUT") && in == nil {
		in = bytes.NewReader([]byte{})
	}

	req, err := http.NewRequest(method, fmt.Sprintf("/v%s%s", api.APIVERSION, path), in)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Docker-Client/"+dockerversion.VERSION)
	req.URL.Host = cli.addr
	req.URL.Scheme = cli.scheme
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

	if api.MatchesContentType(resp.Header.Get("Content-Type"), "application/json") {
		return utils.DisplayJSONMessagesStream(resp.Body, stdout, cli.outFd, cli.isTerminalOut)
	}
	if stdout != nil || stderr != nil {
		// When TTY is ON, use regular copy
		if setRawTerminal {
			_, err = io.Copy(stdout, resp.Body)
		} else {
			_, err = stdcopy.StdCopy(stdout, stderr, resp.Body)
		}
		log.Debugf("[stream] End of stdout")
		return err
	}
	return nil
}

func (cli *KraneCli) pullImageCustomOut(image string, out io.Writer) error {
	v := url.Values{}
	repos, tag := parsers.ParseRepositoryTag(image)
	// pull only the image tagged 'latest' if no tag was specified
	if tag == "" {
		tag = "latest"
	}
	v.Set("fromImage", repos)
	v.Set("tag", tag)

	// Resolve the Repository name from fqn to hostname + name
	hostname, _, err := registry.ResolveRepositoryName(repos)
	if err != nil {
		return err
	}

	// Load the auth config file, to be able to pull the image
	cli.LoadConfigFile()

	// Resolve the Auth config relevant for this server
	authConfig := cli.configFile.ResolveAuthConfig(hostname)
	buf, err := json.Marshal(authConfig)
	if err != nil {
		return err
	}

	registryAuthHeader := []string{
		base64.URLEncoding.EncodeToString(buf),
	}
	if err = cli.stream("POST", "/images/create?"+v.Encode(), nil, out, map[string][]string{"X-Registry-Auth": registryAuthHeader}); err != nil {
		return err
	}
	return nil
}

func (cli *KraneCli) createContainer(config *runconfig.Config, hostConfig *runconfig.HostConfig, cidfile, name string) (engine.Env, error) {
	containerValues := url.Values{}
	if name != "" {
		containerValues.Set("name", name)
	}

	mergedConfig := runconfig.MergeConfigs(config, hostConfig)

	var containerIDFile *cidFile
	if cidfile != "" {
		var err error
		if containerIDFile, err = newCIDFile(cidfile); err != nil {
			return nil, err
		}
		defer containerIDFile.Close()
	}

	//create the container
	stream, statusCode, err := cli.call("POST", "/containers/create?"+containerValues.Encode(), mergedConfig, false)
	//if image not found try to pull it
	if statusCode == 404 {
		fmt.Fprintf(cli.err, "Unable to find image '%s' locally\n", config.Image)

		// we don't want to write to stdout anything apart from container.ID
		if err = cli.pullImageCustomOut(config.Image, cli.err); err != nil {
			return nil, err
		}
		// Retry
		if stream, _, err = cli.call("POST", "/containers/create?"+containerValues.Encode(), mergedConfig, false); err != nil {
			return nil, err
		}
	} else if err != nil {
		return nil, err
	}

	var result engine.Env
	if err := result.Decode(stream); err != nil {
		return nil, err
	}

	for _, warning := range result.GetList("Warnings") {
		fmt.Fprintf(cli.err, "WARNING: %s\n", warning)
	}

	if containerIDFile != nil {
		if err = containerIDFile.Write(result.Get("Id")); err != nil {
			return nil, err
		}
	}

	return result, nil

}

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

func NewKraneCli(in io.ReadCloser, out, err io.Writer, key libtrust.PrivateKey, proto, addr string, tlsConfig *tls.Config) *KraneCli {
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
			isTerminalIn = term.IsTerminal(inFd)
		}
	}

	if out != nil {
		if file, ok := out.(*os.File); ok {
			outFd = file.Fd()
			isTerminalOut = term.IsTerminal(outFd)
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
		*client.NewDockerCli(in, out, err, key, proto, addr, tlsConfig),
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
	client.DockerCli
	proto      string
	addr       string
	configFile *registry.ConfigFile
	in         io.ReadCloser
	out        io.Writer
	err        io.Writer
	key        libtrust.PrivateKey
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

// getExitCode perform an inspect on the container. It returns
// the running state and the exit code.
func getExitCode(cli *KraneCli, containerId string) (bool, int, error) {
	steam, _, err := cli.call("GET", "/containers/"+containerId+"/json", nil, false)
	if err != nil {
		// If we can't connect, then the daemon probably died.
		if err != ErrConnectionRefused {
			return false, -1, err
		}
		return false, -1, nil
	}

	var result engine.Env
	if err := result.Decode(steam); err != nil {
		return false, -1, err
	}

	state := result.GetSubEnv("State")
	return state.GetBool("Running"), state.GetInt("ExitCode"), nil
}

func waitForExit(cli *KraneCli, containerId string) (int, error) {
	stream, _, err := cli.call("POST", "/containers/"+containerId+"/wait", nil, false)
	if err != nil {
		return -1, err
	}

	var out engine.Env
	if err := out.Decode(stream); err != nil {
		return -1, err
	}
	return out.GetInt("StatusCode"), nil
}

func (cli *KraneCli) monitorTtySize(id string, isExec bool) error {
	cli.resizeTty(id, isExec)

	sigchan := make(chan os.Signal, 1)
	gosignal.Notify(sigchan, syscall.SIGWINCH)
	go func() {
		for _ = range sigchan {
			cli.resizeTty(id, isExec)
		}
	}()
	return nil
}

func (cli *KraneCli) hijack(method, path string, setRawTerminal bool, in io.ReadCloser, stdout, stderr io.Writer, started chan io.Closer, data interface{}) error {
	defer func() {
		if started != nil {
			close(started)
		}
	}()

	params, err := cli.encodeData(data)
	if err != nil {
		return err
	}
	req, err := http.NewRequest(method, fmt.Sprintf("/v%s%s", api.APIVERSION, path), params)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Docker-Client/"+dockerversion.VERSION)
	req.Header.Set("Content-Type", "plain/text")
	req.Host = cli.addr

	dial, err := cli.dial()
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return fmt.Errorf("Cannot connect to the Docker daemon. Is 'docker -d' running on this host?")
		}
		return err
	}
	clientconn := httputil.NewClientConn(dial, nil)
	defer clientconn.Close()

	// Server hijacks the connection, error 'connection closed' expected
	clientconn.Do(req)

	rwc, br := clientconn.Hijack()
	defer rwc.Close()

	if started != nil {
		started <- rwc
	}

	var receiveStdout chan error

	var oldState *term.State

	if in != nil && setRawTerminal && cli.isTerminalIn && os.Getenv("NORAW") == "" {
		oldState, err = term.SetRawTerminal(cli.inFd)
		if err != nil {
			return err
		}
		defer term.RestoreTerminal(cli.inFd, oldState)
	}

	if stdout != nil || stderr != nil {
		receiveStdout = promise.Go(func() (err error) {
			defer func() {
				if in != nil {
					if setRawTerminal && cli.isTerminalIn {
						term.RestoreTerminal(cli.inFd, oldState)
					}
					// For some reason this Close call blocks on darwin..
					// As the client exists right after, simply discard the close
					// until we find a better solution.
					if runtime.GOOS != "darwin" {
						in.Close()
					}
				}
			}()

			// When TTY is ON, use regular copy
			if setRawTerminal && stdout != nil {
				_, err = io.Copy(stdout, br)
			} else {
				_, err = stdcopy.StdCopy(stdout, stderr, br)
			}
			log.Debugf("[hijack] End of stdout")
			return err
		})
	}

	sendStdin := promise.Go(func() error {
		if in != nil {
			io.Copy(rwc, in)
			log.Debugf("[hijack] End of stdin")
		}
		if tcpc, ok := rwc.(*net.TCPConn); ok {
			if err := tcpc.CloseWrite(); err != nil {
				log.Debugf("Couldn't send EOF: %s", err)
			}
		} else if unixc, ok := rwc.(*net.UnixConn); ok {
			if err := unixc.CloseWrite(); err != nil {
				log.Debugf("Couldn't send EOF: %s", err)
			}
		}
		// Discard errors due to pipe interruption
		return nil
	})

	if stdout != nil || stderr != nil {
		if err := <-receiveStdout; err != nil {
			log.Debugf("Error receiveStdout: %s", err)
			return err
		}
	}

	if !cli.isTerminalIn {
		if err := <-sendStdin; err != nil {
			log.Debugf("Error sendStdin: %s", err)
			return err
		}
	}
	return nil
}

func (cli *KraneCli) resizeTty(id string, isExec bool) {
	height, width := cli.getTtySize()
	if height == 0 && width == 0 {
		return
	}
	v := url.Values{}
	v.Set("h", strconv.Itoa(height))
	v.Set("w", strconv.Itoa(width))

	path := ""
	if !isExec {
		path = "/containers/" + id + "/resize?"
	} else {
		path = "/exec/" + id + "/resize?"
	}

	if _, _, err := readBody(cli.call("POST", path+v.Encode(), nil, false)); err != nil {
		log.Debugf("Error resize: %s", err)
	}
}

func (cli *KraneCli) getTtySize() (int, int) {
	if !cli.isTerminalOut {
		return 0, 0
	}
	ws, err := term.GetWinsize(cli.outFd)
	if err != nil {
		log.Debugf("Error getting size: %s", err)
		if ws == nil {
			return 0, 0
		}
	}
	return int(ws.Height), int(ws.Width)
}

func (cli *KraneCli) encodeData(data interface{}) (*bytes.Buffer, error) {
	params := bytes.NewBuffer(nil)
	if data != nil {
		if env, ok := data.(engine.Env); ok {
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
	req, err := http.NewRequest(method, fmt.Sprintf("/v%s%s", api.APIVERSION, path), params)
	if err != nil {
		return nil, -1, err
	}
	if passAuthInfo {
		cli.LoadConfigFile()
		// Resolve the Auth config relevant for this server
		authConfig := cli.configFile.ResolveAuthConfig(registry.IndexServerAddress())
		getHeaders := func(authConfig registry.AuthConfig) (map[string][]string, error) {
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
	req.Header.Set("User-Agent", "Docker-Client/"+dockerversion.VERSION)
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
