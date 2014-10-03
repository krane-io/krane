package client

import (
	"fmt"
	// "io"
	"net/url"
	// "os"
	"encoding/json"
	dockerApi "github.com/docker/docker/api"
	"github.com/docker/docker/engine"
	"github.com/docker/docker/pkg/units"
	"github.com/docker/docker/utils"
	"github.com/krane-io/krane/types"
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
			fmt.Fprintf(cli.Err(), "Error: Command not found: %s\n", args[0])
		} else {
			method("--help")
			return nil
		}
	}
	help := fmt.Sprintf("Usage: krane [OPTIONS] COMMAND [arg...]\n -H=[unix://%s]: tcp://host:port to bind/connect to or unix://path/to/socket to use\n\nA self-sufficient runtime for linux containers.\n\nCommands:\n", dockerApi.DEFAULTUNIXSOCKET)
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
		{"commission", "Commision a ship"},
		{"decomission", "Decomission a ship"},
	} {
		help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
	}
	fmt.Fprintf(cli.Err(), "%s\n", help)
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

	body, _, err := cli.ReadBody(cli.Call("POST", "/ships/create?"+v.Encode(), parameters, false))
	if err != nil {
		return err
	}

	fmt.Printf("%s", string(body))

	return nil
}

func (cli *KraneCli) CmdShips(args ...string) error {
	cmd := cli.Subcmd("ships", "[OPTIONS]", "List the number of ships")
	quiet := cmd.Bool([]string{"q", "-quiet"}, false, "Only display numeric IDs")

	if err := cmd.Parse(args); err != nil {
		return nil
	}

	v := url.Values{}

	body, _, err := cli.ReadBody(cli.Call("GET", "/ships/json?"+v.Encode(), nil, false))
	if err != nil {
		return err
	}

	var ships []types.Ship
	json.Unmarshal(body, &ships)

	w := tabwriter.NewWriter(cli.Out(), 20, 1, 3, ' ', 0)
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

	body, _, err := cli.ReadBody(cli.Call("GET", "/containers/json?"+v.Encode(), nil, false))
	if err != nil {
		return err
	}

	ships := engine.NewTable("Created", 0)
	if _, err := ships.ReadListFrom(body); err != nil {
		return err
	}

	w := tabwriter.NewWriter(cli.Out(), 20, 1, 3, ' ', 0)
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
			outShipFQDN   = outShip.Get("Fqdn")
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
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s ago\t%s\t%s\t%s\t", outID, out.Get("Image"), outCommand, units.HumanDuration(time.Now().UTC().Sub(time.Unix(out.GetInt64("Created"), 0))), out.Get("Status"), dockerApi.DisplayablePorts(ports), strings.Join(outNames, ","), outShipFQDN)
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
