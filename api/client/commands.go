package client

import (
	"encoding/json"
	"fmt"
	"github.com/docker/docker/api"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/opts"
	"github.com/docker/docker/pkg/log"
	"github.com/docker/docker/pkg/promise"
	"github.com/docker/docker/pkg/signal"
	"github.com/docker/docker/pkg/units"
	"github.com/docker/docker/runconfig"
	"github.com/docker/docker/utils"
	"github.com/krane-io/krane/types"
	"io"
	"net/url"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"
)

type CreateShip struct {
	Fqdn string
	Plan string
}

func (cli *KraneCli) CmdHelp(args ...string) error {
	if len(args) > 0 {
		method, exists := cli.getMethod(args[0])
		if !exists {
			fmt.Fprintf(cli.err, "Error: Command not found: %s\n", args[0])
		} else {
			method("--help")
			return nil
		}
	}
	help := fmt.Sprintf("Usage: krane [OPTIONS] COMMAND [arg...]\n -H=[unix://%s]: tcp://host:port to bind/connect to or unix://path/to/socket to use\n\nA self-sufficient runtime for linux containers.\n\nCommands:\n", api.DEFAULTUNIXSOCKET)
	for _, command := range [][]string{
		{"attach", "Attach to a running container"},
		{"build", "Build an image from a Dockerfile"},
		{"commit", "Create a new image from a container's changes"},
		{"cp", "Copy files/folders from a container's filesystem to the host path"},
		{"diff", "Inspect changes on a container's filesystem"},
		{"events", "Get real time events from the server"},
		{"export", "Stream the contents of a container as a tar archive"},
		{"history", "Show the history of an image"},
		{"images", "List images"},
		{"import", "Create a new filesystem image from the contents of a tarball"},
		{"info", "Display system-wide information"},
		{"inspect", "Return low-level information on a container"},
		{"kill", "Kill a running container"},
		{"load", "Load an image from a tar archive"},
		{"login", "Register or log in to a Docker registry server"},
		{"logout", "Log out from a Docker registry server"},
		{"logs", "Fetch the logs of a container"},
		{"port", "Lookup the public-facing port that is NAT-ed to PRIVATE_PORT"},
		{"pause", "Pause all processes within a container"},
		{"ps", "List containers"},
		{"pull", "Pull an image or a repository from a Docker registry server"},
		{"push", "Push an image or a repository to a Docker registry server"},
		{"restart", "Restart a running container"},
		{"rm", "Remove one or more containers"},
		{"rmi", "Remove one or more images"},
		{"run", "Run a command in a new container"},
		{"save", "Save an image to a tar archive"},
		{"search", "Search for an image on the Docker Hub"},
		{"start", "Start a stopped container"},
		{"stop", "Stop a running container"},
		{"tag", "Tag an image into a repository"},
		{"top", "Lookup the running processes of a container"},
		{"unpause", "Unpause a paused container"},
		{"version", "Show the Docker version information"},
		{"wait", "Block until a container stops, then print its exit code"},
		{"ships", "List the number of ships"},
		{"plans", "Search the for a specific cloud plan"},
		{"commission", "Commision a ship"},
		{"decomission", "Decomission a ship"},
	} {
		help += fmt.Sprintf("    %-15.15s%s\n", command[0], command[1])
	}
	fmt.Fprintf(cli.err, "%s\n", help)
	return nil
}

func (cli *KraneCli) CmdRun(args ...string) error {
	// FIXME: just use runconfig.Parse already
	cmd := cli.Subcmd("run", "IMAGE [COMMAND] [ARG...]", "Run a command in a new container")

	// These are flags not stored in Config/HostConfig
	var (
		flAutoRemove = cmd.Bool([]string{"#rm", "-rm"}, false, "Automatically remove the container when it exits (incompatible with -d)")
		flDetach     = cmd.Bool([]string{"d", "-detach"}, false, "Detached mode: run the container in the background and print the new container ID")
		flSigProxy   = cmd.Bool([]string{"#sig-proxy", "-sig-proxy"}, true, "Proxy received signals to the process (even in non-TTY mode). SIGCHLD, SIGSTOP, and SIGKILL are not proxied.")
		flName       = cmd.String([]string{"#name", "-name"}, "", "Assign a name to the container")
		flShip       = cmd.String([]string{"ship", "-ship"}, "", "Ship name to deploy at")
		flAttach     *opts.ListOpts

		ErrConflictAttachDetach               = fmt.Errorf("Conflicting options: -a and -d")
		ErrConflictRestartPolicyAndAutoRemove = fmt.Errorf("Conflicting options: --restart and --rm")
		ErrConflictDetachAutoRemove           = fmt.Errorf("Conflicting options: --rm and -d")
	)

	config, hostConfig, cmd, err := runconfig.Parse(cmd, args, nil)
	if err != nil {
		return err
	}
	if config.Image == "" || *flShip == "" {
		cmd.Usage()
		return nil
	}

	if *flDetach {
		if fl := cmd.Lookup("attach"); fl != nil {
			flAttach = fl.Value.(*opts.ListOpts)
			if flAttach.Len() != 0 {
				return ErrConflictAttachDetach
			}
		}
		if *flAutoRemove {
			return ErrConflictDetachAutoRemove
		}

		config.AttachStdin = false
		config.AttachStdout = false
		config.AttachStderr = false
		config.StdinOnce = false
	}

	// Disable flSigProxy when in TTY mode
	sigProxy := *flSigProxy
	if config.Tty {
		sigProxy = false
	}

	runResult, err := cli.createContainer(config, hostConfig, hostConfig.ContainerIDFile, *flName)
	if err != nil {
		return err
	}

	if sigProxy {
		sigc := cli.forwardAllSignals(runResult.Get("Id"))
		defer signal.StopCatch(sigc)
	}

	var (
		waitDisplayId chan struct{}
		errCh         chan error
	)

	if !config.AttachStdout && !config.AttachStderr {
		// Make this asynchronous to allow the client to write to stdin before having to read the ID
		waitDisplayId = make(chan struct{})
		go func() {
			defer close(waitDisplayId)
			fmt.Fprintf(cli.out, "%s\n", runResult.Get("Id"))
		}()
	}

	if *flAutoRemove && (hostConfig.RestartPolicy.Name == "always" || hostConfig.RestartPolicy.Name == "on-failure") {
		return ErrConflictRestartPolicyAndAutoRemove
	}

	// We need to instantiate the chan because the select needs it. It can
	// be closed but can't be uninitialized.
	hijacked := make(chan io.Closer)

	// Block the return until the chan gets closed
	defer func() {
		log.Debugf("End of CmdRun(), Waiting for hijack to finish.")
		if _, ok := <-hijacked; ok {
			log.Errorf("Hijack did not finish (chan still open)")
		}
	}()

	if config.AttachStdin || config.AttachStdout || config.AttachStderr {
		var (
			out, stderr io.Writer
			in          io.ReadCloser
			v           = url.Values{}
		)
		v.Set("stream", "1")

		if config.AttachStdin {
			v.Set("stdin", "1")
			in = cli.in
		}
		if config.AttachStdout {
			v.Set("stdout", "1")
			out = cli.out
		}
		if config.AttachStderr {
			v.Set("stderr", "1")
			if config.Tty {
				stderr = cli.out
			} else {
				stderr = cli.err
			}
		}

		errCh = promise.Go(func() error {
			return cli.hijack("POST", "/containers/"+runResult.Get("Id")+"/attach?"+v.Encode(), config.Tty, in, out, stderr, hijacked, nil)
		})
	} else {
		close(hijacked)
	}

	// Acknowledge the hijack before starting
	select {
	case closer := <-hijacked:
		// Make sure that the hijack gets closed when returning (results
		// in closing the hijack chan and freeing server's goroutines)
		if closer != nil {
			defer closer.Close()
		}
	case err := <-errCh:
		if err != nil {
			log.Debugf("Error hijack: %s", err)
			return err
		}
	}

	//start the container
	if _, _, err = readBody(cli.call("POST", "/containers/"+runResult.Get("Id")+"/start", hostConfig, false)); err != nil {
		return err
	}

	if (config.AttachStdin || config.AttachStdout || config.AttachStderr) && config.Tty && cli.isTerminalOut {
		if err := cli.monitorTtySize(runResult.Get("Id"), false); err != nil {
			log.Errorf("Error monitoring TTY size: %s", err)
		}
	}

	if errCh != nil {
		if err := <-errCh; err != nil {
			log.Debugf("Error hijack: %s", err)
			return err
		}
	}

	// Detached mode: wait for the id to be displayed and return.
	if !config.AttachStdout && !config.AttachStderr {
		// Detached mode
		<-waitDisplayId
		return nil
	}

	var status int

	// Attached mode
	if *flAutoRemove {
		// Autoremove: wait for the container to finish, retrieve
		// the exit code and remove the container
		if _, _, err := readBody(cli.call("POST", "/containers/"+runResult.Get("Id")+"/wait", nil, false)); err != nil {
			return err
		}
		if _, status, err = getExitCode(cli, runResult.Get("Id")); err != nil {
			return err
		}
		if _, _, err := readBody(cli.call("DELETE", "/containers/"+runResult.Get("Id")+"?v=1", nil, false)); err != nil {
			return err
		}
	} else {
		// No Autoremove: Simply retrieve the exit code
		if !config.Tty {
			// In non-TTY mode, we can't detach, so we must wait for container exit
			if status, err = waitForExit(cli, runResult.Get("Id")); err != nil {
				return err
			}
		} else {
			// In TTY mode, there is a race: if the process dies too slowly, the state could
			// be updated after the getExitCode call and result in the wrong exit code being reported
			if _, status, err = getExitCode(cli, runResult.Get("Id")); err != nil {
				return err
			}
		}
	}
	if status != 0 {
		return &utils.StatusError{StatusCode: status}
	}
	return nil
}

func (cli *KraneCli) CmdDecomission(args ...string) error {
	cmd := cli.Subcmd("decomission", "[OPTIONS] SHIP [SHIP]", "Decomission a ship")
	time := cmd.String([]string{"t", "-time"}, "10", "Number of seconds to wait for the decomissioning of ship before killing it.")

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if len(args) == 0 {
		cmd.Usage()
		return nil
	}

	v := url.Values{}
	v.Set("name", args[0])
	v.Set("time", *time)

	body, _, err := readBody(cli.call("DELETE", "/ships/decomission?"+v.Encode(), nil, false))
	if err != nil {
		return err
	}

	fmt.Printf("%s", string(body))

	return nil
}

func (cli *KraneCli) CmdCommission(args ...string) error {
	cmd := cli.Subcmd("commission", "[OPTIONS]", "Commision a ship")
	name := cmd.String([]string{"n", "-name"}, "", "Name of Ship")
	fqdn := cmd.String([]string{"f", "-fqdn"}, "", "Full qualify domain name of Ship")
	plan := cmd.String([]string{"p", "-plan"}, "53f0f10fd8a5975a1c000395", "Cloud Plan to use for the commissioning of the Ship")

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if *fqdn == "" || *name == "" {
		cmd.Usage()
		return nil
	}

	parameters := CreateShip{*fqdn, *plan}

	v := url.Values{}

	v.Set("name", *name)

	body, _, err := readBody(cli.call("POST", "/ships/create?"+v.Encode(), parameters, false))
	if err != nil {
		return err
	}

	fmt.Printf("%s", string(body))

	return nil
}

func (cli *KraneCli) CmdPlans(args ...string) error {
	cmd := cli.Subcmd("plans", "NAME", "Search the for a specific cloud plan of the availiable ")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")
	all := cmd.Bool([]string{"a", "-all"}, false, "Show all the cloud plans")

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	if len(args) == 0 {
		cmd.Usage()
		return nil
	}

	v := url.Values{}
	if *all {
		v.Set("name", "all")
	} else {
		v.Set("name", args[0])
	}

	body, _, err := readBody(cli.call("GET", "/plans/json?"+v.Encode(), nil, false))
	if err != nil {
		return err
	}

	var plans []types.Plan
	json.Unmarshal(body, &plans)

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tPROVIDER\tCONTINENT\tREGION\tPLAN\n")

	for _, plan := range plans {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n", plan.Id, plan.Provider, plan.Continent, plan.Region, plan.Plan)
	}

	if !*quiet {
		w.Flush()
	}

	return nil
}

func (cli *KraneCli) CmdShips(args ...string) error {
	cmd := cli.Subcmd("ships", "[OPTIONS]", "List the number of ships")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	v := url.Values{}

	body, _, err := readBody(cli.call("GET", "/ships/json?"+v.Encode(), nil, false))
	if err != nil {
		return err
	}

	var ships []types.Ship
	json.Unmarshal(body, &ships)

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	fmt.Fprint(w, "ID\tNAME\tFQDN\tIP\tSTATE\tOS\tPLAN\n")

	for _, ship := range ships {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n", ship.Id, ship.Name, ship.Fqdn, ship.Ip, ship.State, ship.Os, ship.Plan)
	}

	if !*quiet {
		w.Flush()
	}

	return nil
}

func (cli *KraneCli) CmdPs(args ...string) error {
	cmd := cli.Subcmd("ps", "[OPTIONS]", "List containers")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")
	size := cmd.Bool([]string{"s", "-size"}, false, "Display sizes")
	all := cmd.Bool([]string{"a", "-all"}, false, "Show all containers. Only running containers are shown by default.")
	noTrunc := cmd.Bool([]string{"#notrunc", "-no-trunc"}, false, "Don't truncate output")
	nLatest := cmd.Bool([]string{"l", "-latest"}, false, "Show only the latest created container, include non-running ones.")
	since := cmd.String([]string{"#sinceId", "#-since-id", "-since"}, "", "Show only containers created since Id or Name, include non-running ones.")
	before := cmd.String([]string{"#beforeId", "#-before-id", "-before"}, "", "Show only container created before Id or Name, include non-running ones.")
	last := cmd.Int([]string{"n"}, -1, "Show n last created containers, include non-running ones.")

	if err := cmd.Parse(args); err != nil {
		return nil
	}
	v := url.Values{}
	if *last == -1 && *nLatest {
		*last = 1
	}
	if *all {
		v.Set("all", "1")
	}
	if *last != -1 {
		v.Set("limit", strconv.Itoa(*last))
	}
	if *since != "" {
		v.Set("since", *since)
	}
	if *before != "" {
		v.Set("before", *before)
	}
	if *size {
		v.Set("size", "1")
	}

	body, _, err := readBody(cli.call("GET", "/containers/json?"+v.Encode(), nil, false))
	if err != nil {
		return err
	}

	ships := engine.NewTable("Created", 0)
	if _, err := ships.ReadListFrom(body); err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.out, 20, 1, 3, ' ', 0)
	if !*quiet {
		fmt.Fprint(w, "CONTAINER ID\tIMAGE\tCOMMAND\tCREATED\tSTATUS\tPORTS\tNAMES\tSHIP")
		if *size {
			fmt.Fprintln(w, "\tSIZE")
		} else {
			fmt.Fprint(w, "\n")
		}
	}

	for _, outShip := range ships.Data {
		var (
			outShipFQDN   = outShip.Get("fqdn")
			outContainers = outShip.Get("Containers")
		)

		outs := engine.NewTable("Created", 0)

		if _, err := outs.ReadListFrom([]byte(outContainers)); err != nil {
			return err
		}

		for _, out := range outs.Data {
			var (
				outID    = out.Get("Id")
				outNames = out.GetList("Names")
			)

			if !*noTrunc {
				outID = utils.TruncateID(outID)
			}

			// Remove the leading / from the names
			for i := 0; i < len(outNames); i++ {
				outNames[i] = outNames[i][1:]
			}

			if !*quiet {
				var (
					outCommand = out.Get("Command")
					ports      = engine.NewTable("", 0)
				)
				outCommand = strconv.Quote(outCommand)
				if !*noTrunc {
					outCommand = utils.Trunc(outCommand, 20)
				}
				ports.ReadListFrom([]byte(out.Get("Ports")))
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s ago\t%s\t%s\t%s\t", outID, out.Get("Image"), outCommand, units.HumanDuration(time.Now().UTC().Sub(time.Unix(out.GetInt64("Created"), 0))), out.Get("Status"), api.DisplayablePorts(ports), strings.Join(outNames, ","), outShipFQDN)
				if *size {
					if out.GetInt("SizeRootFs") > 0 {
						fmt.Fprintf(w, "%s (virtual %s)\n", units.HumanSize(out.GetInt64("SizeRw")), units.HumanSize(out.GetInt64("SizeRootFs")))
					} else {
						fmt.Fprintf(w, "%s\n", units.HumanSize(out.GetInt64("SizeRw")))
					}
				} else {
					fmt.Fprint(w, "\n")
				}
			} else {
				fmt.Fprintln(w, outID)
			}
		}
	}

	if !*quiet {
		w.Flush()
	}
	return nil
}
