package server

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"expvar"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/docker/docker/api"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/pkg/listenbuffer"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/pkg/systemd"
	"github.com/docker/docker/pkg/version"
	"github.com/docker/docker/utils"
	"github.com/docker/libcontainer/user"
	"github.com/gorilla/mux"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/pprof"
	"os"
	"strconv"
	"strings"
	"syscall"
)

type HttpApiFunc func(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error

func getInfo(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.Header().Set("Content-Type", "application/json")
	eng.ServeHTTP(w, r)
	return nil
}

func streamJSON(job *engine.Job, w http.ResponseWriter, flush bool) {
	w.Header().Set("Content-Type", "application/json")
	if job.GetenvBool("lineDelim") {
		w.Header().Set("Content-Type", "application/x-json-stream")
	}

	if flush {
		job.Stdout.Add(utils.NewWriteFlusher(w))
	} else {
		job.Stdout.Add(w)
	}
}

func hijackServer(w http.ResponseWriter) (io.ReadCloser, io.Writer, error) {
	conn, _, err := w.(http.Hijacker).Hijack()
	if err != nil {
		return nil, nil, err
	}
	// Flush the options to make sure the client sets the raw mode
	conn.Write([]byte{})
	return conn, conn, nil
}

func writeJSON(w http.ResponseWriter, code int, v engine.Env) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return v.Encode(w)
}

func postContainersAttach(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}

	var (
		job    = eng.Job("container_inspect", vars["name"])
		c, err = job.Stdout.AddEnv()
	)
	if err != nil {
		return err
	}
	if err = job.Run(); err != nil {
		return err
	}

	inStream, outStream, err := hijackServer(w)
	if err != nil {
		return err
	}
	defer func() {
		if tcpc, ok := inStream.(*net.TCPConn); ok {
			tcpc.CloseWrite()
		} else {
			inStream.Close()
		}
	}()
	defer func() {
		if tcpc, ok := outStream.(*net.TCPConn); ok {
			tcpc.CloseWrite()
		} else if closer, ok := outStream.(io.Closer); ok {
			closer.Close()
		}
	}()

	var errStream io.Writer

	fmt.Fprintf(outStream, "HTTP/1.1 200 OK\r\nContent-Type: application/vnd.docker.raw-stream\r\n\r\n")

	if c.GetSubEnv("Config") != nil && !c.GetSubEnv("Config").GetBool("Tty") && version.GreaterThanOrEqualTo("1.6") {
		errStream = stdcopy.NewStdWriter(outStream, stdcopy.Stderr)
		outStream = stdcopy.NewStdWriter(outStream, stdcopy.Stdout)
	} else {
		errStream = outStream
	}

	job = eng.Job("attach", vars["name"])
	job.Setenv("logs", r.Form.Get("logs"))
	job.Setenv("stream", r.Form.Get("stream"))
	job.Setenv("stdin", r.Form.Get("stdin"))
	job.Setenv("stdout", r.Form.Get("stdout"))
	job.Setenv("stderr", r.Form.Get("stderr"))
	job.Stdin.Add(inStream)
	job.Stdout.Add(outStream)
	job.Stderr.Set(errStream)
	if err := job.Run(); err != nil {
		fmt.Fprintf(outStream, "Error attaching: %s\n", err)

	}
	return nil
}

func postContainersStart(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	var (
		name = vars["name"]
		job  = eng.Job("start", name)
	)

	// If contentLength is -1, we can assumed chunked encoding
	// or more technically that the length is unknown
	// http://golang.org/src/pkg/net/http/request.go#L139
	// net/http otherwise seems to swallow any headers related to chunked encoding
	// including r.TransferEncoding
	// allow a nil body for backwards compatibility
	if r.Body != nil && (r.ContentLength > 0 || r.ContentLength == -1) {
		if err := checkForJson(r); err != nil {
			return err
		}

		if err := job.DecodeEnv(r.Body); err != nil {
			return err
		}
	}

	if err := job.Run(); err != nil {
		if err.Error() == "Container already started" {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func postContainersStop(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	if vars == nil {
		return fmt.Errorf("Missing parameter")
	}
	job := eng.Job("stop", vars["name"])
	job.Setenv("t", r.Form.Get("t"))
	if err := job.Run(); err != nil {
		if err.Error() == "Container already stopped" {
			w.WriteHeader(http.StatusNotModified)
			return nil
		}
		return err
	}
	w.WriteHeader(http.StatusNoContent)
	return nil
}

func postContainersCreate(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return nil
	}
	var (
		out          engine.Env
		job          = eng.Job("create", r.Form.Get("name"))
		outWarnings  []string
		stdoutBuffer = bytes.NewBuffer(nil)
		warnings     = bytes.NewBuffer(nil)
	)

	if err := checkForJson(r); err != nil {
		return err
	}

	if err := job.DecodeEnv(r.Body); err != nil {
		return err
	}
	// Read container ID from the first line of stdout
	job.Stdout.Add(stdoutBuffer)
	// Read warnings from stderr
	job.Stderr.Add(warnings)
	if err := job.Run(); err != nil {
		return err
	}
	// Parse warnings from stderr
	scanner := bufio.NewScanner(warnings)
	for scanner.Scan() {
		outWarnings = append(outWarnings, scanner.Text())
	}
	out.Set("Id", engine.Tail(stdoutBuffer, 1))
	out.SetList("Warnings", outWarnings)

	return writeJSON(w, http.StatusCreated, out)
}

func getContainersJSON(eng *engine.Engine, version version.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	var (
		err  error
		outs *engine.Table
		job  = eng.Job("containers")
	)

	job.Setenv("all", r.Form.Get("all"))
	job.Setenv("size", r.Form.Get("size"))
	job.Setenv("since", r.Form.Get("since"))
	job.Setenv("before", r.Form.Get("before"))
	job.Setenv("limit", r.Form.Get("limit"))
	job.Setenv("filters", r.Form.Get("filters"))

	if version.GreaterThanOrEqualTo("1.5") {
		streamJSON(job, w, false)
	} else if outs, err = job.Stdout.AddTable(); err != nil {
		return err
	}
	if err = job.Run(); err != nil {
		return err
	}
	if version.LessThan("1.5") { // Convert to legacy format
		for _, out := range outs.Data {
			ports := engine.NewTable("", 0)
			ports.ReadListFrom([]byte(out.Get("Ports")))
			out.Set("Ports", api.DisplayablePorts(ports))
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err = outs.WriteListTo(w); err != nil {
			return err
		}
	}
	return nil
}

func AttachProfiler(router *mux.Router) {
	router.HandleFunc("/debug/vars", expvarHandler)
	router.HandleFunc("/debug/pprof/", pprof.Index)
	router.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	router.HandleFunc("/debug/pprof/profile", pprof.Profile)
	router.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	router.HandleFunc("/debug/pprof/heap", pprof.Handler("heap").ServeHTTP)
	router.HandleFunc("/debug/pprof/goroutine", pprof.Handler("goroutine").ServeHTTP)
	router.HandleFunc("/debug/pprof/threadcreate", pprof.Handler("threadcreate").ServeHTTP)
}

func makeHttpHandler(eng *engine.Engine, logging bool, localMethod string, localRoute string, handlerFunc HttpApiFunc, enableCors bool, dockerVersion version.Version) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// log the request
		log.Debugf("Calling %s %s", localMethod, localRoute)

		if logging {
			log.Println(r.Method, r.RequestURI)
		}

		if strings.Contains(r.Header.Get("User-Agent"), "Docker-Client/") {
			userAgent := strings.Split(r.Header.Get("User-Agent"), "/")
			if len(userAgent) == 2 && !dockerVersion.Equal(version.Version(userAgent[1])) {
				log.Debugf("Warning: client and server don't have the same version (client: %s, server: %s)", userAgent[1], dockerVersion)
			}
		}
		version := version.Version(mux.Vars(r)["version"])
		if version == "" {
			version = api.APIVERSION
		}
		if enableCors {
			writeCorsHeaders(w, r)
		}

		if version.GreaterThan(api.APIVERSION) {
			http.Error(w, fmt.Errorf("client and server don't have same version (client : %s, server: %s)", version, api.APIVERSION).Error(), http.StatusNotFound)
			return
		}

		if err := handlerFunc(eng, version, w, r, mux.Vars(r)); err != nil {
			log.Errorf("Error making handler: %s", err)
			httpError(w, err)
		}
	}
}

// ServeFD creates an http.Server and sets it up to serve given a socket activated
// argument.
func ServeFd(addr string, handle http.Handler) error {
	ls, e := systemd.ListenFD(addr)
	if e != nil {
		return e
	}

	chErrors := make(chan error, len(ls))

	// We don't want to start serving on these sockets until the
	// "initserver" job has completed. Otherwise required handlers
	// won't be ready.
	<-activationLock

	// Since ListenFD will return one or more sockets we have
	// to create a go func to spawn off multiple serves
	for i := range ls {
		listener := ls[i]
		go func() {
			httpSrv := http.Server{Handler: handle}
			chErrors <- httpSrv.Serve(listener)
		}()
	}

	for i := 0; i < len(ls); i += 1 {
		err := <-chErrors
		if err != nil {
			return err
		}
	}

	return nil
}

func lookupGidByName(nameOrGid string) (int, error) {
	groups, err := user.ParseGroupFilter(func(g *user.Group) bool {
		return g.Name == nameOrGid || strconv.Itoa(g.Gid) == nameOrGid
	})
	if err != nil {
		return -1, err
	}
	if groups != nil && len(groups) > 0 {
		return groups[0].Gid, nil
	}
	return -1, fmt.Errorf("Group %s not found", nameOrGid)
}

func changeGroup(addr string, nameOrGid string) error {
	gid, err := lookupGidByName(nameOrGid)
	if err != nil {
		return err
	}

	log.Debugf("%s group found. gid: %d", nameOrGid, gid)
	return os.Chown(addr, 0, gid)
}

//If we don't do this, POST method without Content-type (even with empty body) will fail
func parseForm(r *http.Request) error {
	if r == nil {
		return nil
	}
	if err := r.ParseForm(); err != nil && !strings.HasPrefix(err.Error(), "mime:") {
		return err
	}
	return nil
}

// Check to make sure request's Content-Type is application/json
func checkForJson(r *http.Request) error {
	ct := r.Header.Get("Content-Type")

	// No Content-Type header is ok as long as there's no Body
	if ct == "" {
		if r.Body == nil || r.ContentLength == 0 {
			return nil
		}
	}

	// Otherwise it better be json
	if api.MatchesContentType(ct, "application/json") {
		return nil
	}
	return fmt.Errorf("Content-Type specified (%s) must be 'application/json'", ct)
}

var (
	activationLock chan struct{}
)

func httpError(w http.ResponseWriter, err error) {
	statusCode := http.StatusInternalServerError
	// FIXME: this is brittle and should not be necessary.
	// If we need to differentiate between different possible error types, we should
	// create appropriate error types with clearly defined meaning.
	if strings.Contains(err.Error(), "No such") {
		statusCode = http.StatusNotFound
	} else if strings.Contains(err.Error(), "Bad parameter") {
		statusCode = http.StatusBadRequest
	} else if strings.Contains(err.Error(), "Conflict") {
		statusCode = http.StatusConflict
	} else if strings.Contains(err.Error(), "Impossible") {
		statusCode = http.StatusNotAcceptable
	} else if strings.Contains(err.Error(), "Wrong login/password") {
		statusCode = http.StatusUnauthorized
	} else if strings.Contains(err.Error(), "hasn't been activated") {
		statusCode = http.StatusForbidden
	}

	if err != nil {
		log.Errorf("HTTP Error: statusCode=%d %s", statusCode, err.Error())
		http.Error(w, err.Error(), statusCode)
	}
}

func writeCorsHeaders(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept")
	w.Header().Add("Access-Control-Allow-Methods", "GET, POST, DELETE, PUT, OPTIONS")
}

// Replicated from expvar.go as not public.
func expvarHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	fmt.Fprintf(w, "{\n")
	first := true
	expvar.Do(func(kv expvar.KeyValue) {
		if !first {
			fmt.Fprintf(w, ",\n")
		}
		first = false
		fmt.Fprintf(w, "%q: %s", kv.Key, kv.Value)
	})
	fmt.Fprintf(w, "\n}\n")
}

func createRouter(eng *engine.Engine, logging, enableCors bool, dockerVersion string) (*mux.Router, error) {
	r := mux.NewRouter()
	if os.Getenv("DEBUG") != "" {
		AttachProfiler(r)
	}

	for method, routes := range ServerRoutes {
		for route, fct := range routes {
			log.Debugf("Registering %s, %s", method, route)
			// NOTE: scope issue, make sure the variables are local and won't be changed
			localRoute := route
			localFct := fct
			localMethod := method

			// build the handler function
			f := makeHttpHandler(eng, logging, localMethod, localRoute, localFct, enableCors, version.Version(dockerVersion))

			// add the new route
			if localRoute == "" {
				r.Methods(localMethod).HandlerFunc(f)
			} else {
				r.Path("/v{version:[0-9.]+}" + localRoute).Methods(localMethod).HandlerFunc(f)
				r.Path(localRoute).Methods(localMethod).HandlerFunc(f)
			}
		}
	}

	return r, nil
}

// ListenAndServe sets up the required http.Server and gets it listening for
// each addr passed in and does protocol specific checking.
func ListenAndServe(proto, addr string, job *engine.Job) error {
	log.Debugf("Logging: %t", job.GetenvBool("Logging"))
	log.Debugf("EnableCors: %t", job.GetenvBool("EnableCors"))
	log.Debugf("Version: %t", job.GetenvBool("Version"))

	AttachJobs(job.Eng)

	var l net.Listener
	r, err := createRouter(job.Eng, job.GetenvBool("Logging"), job.GetenvBool("EnableCors"), job.Getenv("Version"))
	if err != nil {
		return err
	}

	if proto == "fd" {
		return ServeFd(addr, r)
	}

	if proto == "unix" {
		if err := syscall.Unlink(addr); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	var oldmask int
	if proto == "unix" {
		oldmask = syscall.Umask(0777)
	}

	if job.GetenvBool("BufferRequests") {
		l, err = listenbuffer.NewListenBuffer(proto, addr, activationLock)
	} else {
		l, err = net.Listen(proto, addr)
	}

	if proto == "unix" {
		syscall.Umask(oldmask)
	}
	if err != nil {
		return err
	}

	if proto != "unix" && (job.GetenvBool("Tls") || job.GetenvBool("TlsVerify")) {
		tlsCert := job.Getenv("TlsCert")
		tlsKey := job.Getenv("TlsKey")
		cert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
		if err != nil {
			return fmt.Errorf("Couldn't load X509 key pair (%s, %s): %s. Key encrypted?",
				tlsCert, tlsKey, err)
		}
		tlsConfig := &tls.Config{
			NextProtos:   []string{"http/1.1"},
			Certificates: []tls.Certificate{cert},
		}
		if job.GetenvBool("TlsVerify") {
			certPool := x509.NewCertPool()
			file, err := ioutil.ReadFile(job.Getenv("TlsCa"))
			if err != nil {
				return fmt.Errorf("Couldn't read CA certificate: %s", err)
			}
			certPool.AppendCertsFromPEM(file)

			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			tlsConfig.ClientCAs = certPool
		}
		l = tls.NewListener(l, tlsConfig)
	}

	// Basic error and sanity checking
	switch proto {
	case "tcp":
		if !strings.HasPrefix(addr, "127.0.0.1") && !job.GetenvBool("TlsVerify") {
			log.Println("/!\\ DON'T BIND ON ANOTHER IP ADDRESS THAN 127.0.0.1 IF YOU DON'T KNOW WHAT YOU'RE DOING /!\\")
		}
	case "unix":
		socketGroup := job.Getenv("SocketGroup")
		if socketGroup != "" {
			if err := changeGroup(addr, socketGroup); err != nil {
				if socketGroup == "docker" {
					// if the user hasn't explicitly specified the group ownership, don't fail on errors.
					log.Debugf("Warning: could not chgrp %s to docker: %s", addr, err.Error())
				} else {
					return err
				}
			}
		}
		if err := os.Chmod(addr, 0660); err != nil {
			return err
		}
	default:
		return fmt.Errorf("Invalid protocol format.")
	}

	httpSrv := http.Server{Addr: addr, Handler: r}
	return httpSrv.Serve(l)
}

// ServeApi loops through all of the protocols sent in to docker and spawns
// off a go routine to setup a serving http.Server for each.
func ServeApi(job *engine.Job) engine.Status {
	if len(job.Args) == 0 {
		return job.Errorf("usage: %s PROTO://ADDR [PROTO://ADDR ...]", job.Name)
	}
	var (
		protoAddrs = job.Args
		chErrors   = make(chan error, len(protoAddrs))
	)
	activationLock = make(chan struct{})

	for _, protoAddr := range protoAddrs {
		protoAddrParts := strings.SplitN(protoAddr, "://", 2)
		if len(protoAddrParts) != 2 {
			return job.Errorf("usage: %s PROTO://ADDR [PROTO://ADDR ...]", job.Name)
		}
		go func() {
			log.Printf("Listening for HTTP on %s (%s)\n", protoAddrParts[0], protoAddrParts[1])
			chErrors <- ListenAndServe(protoAddrParts[0], protoAddrParts[1], job)
		}()
	}

	for i := 0; i < len(protoAddrs); i += 1 {
		err := <-chErrors
		if err != nil {
			return job.Error(err)
		}
	}

	return engine.StatusOK
}
