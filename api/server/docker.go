package server

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"net/http"

	dockerApi "github.com/docker/docker/api"
	dockerEngine "github.com/docker/docker/engine"
	dockerPkgStdCopy "github.com/docker/docker/pkg/stdcopy"
	dockerPkgVersion "github.com/docker/docker/pkg/version"
)

func getInfo(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	w.Header().Set("Content-Type", "application/json")
	eng.ServeHTTP(w, r)
	return nil
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

func writeJSON(w http.ResponseWriter, code int, v dockerEngine.Env) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return v.Encode(w)
}

func postContainersAttach(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
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
		errStream = dockerPkgStdCopy.NewStdWriter(outStream, dockerPkgStdCopy.Stderr)
		outStream = dockerPkgStdCopy.NewStdWriter(outStream, dockerPkgStdCopy.Stdout)
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

func postContainersStart(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
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

func postContainersStop(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
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

func postContainersCreate(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return nil
	}
	var (
		out          dockerEngine.Env
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
	out.Set("Id", dockerEngine.Tail(stdoutBuffer, 1))
	out.SetList("Warnings", outWarnings)

	return writeJSON(w, http.StatusCreated, out)
}

func getContainersJSON(eng *dockerEngine.Engine, version dockerPkgVersion.Version, w http.ResponseWriter, r *http.Request, vars map[string]string) error {
	if err := parseForm(r); err != nil {
		return err
	}
	var (
		err  error
		outs *dockerEngine.Table
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
			ports := dockerEngine.NewTable("", 0)
			ports.ReadListFrom([]byte(out.Get("Ports")))
			out.Set("Ports", dockerApi.DisplayablePorts(ports))
		}
		w.Header().Set("Content-Type", "application/json")
		if _, err = outs.WriteListTo(w); err != nil {
			return err
		}
	}
	return nil
}
