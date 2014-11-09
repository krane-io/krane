package main

import (
	"fmt"
	"github.com/docker/docker/api"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
	"github.com/docker/docker/utils"
	"github.com/krane-io/krane/api/client"
	"github.com/krane-io/krane/config"
	"github.com/krane-io/krane/drivers/aws"
	"github.com/krane-io/krane/drivers/concerto"
	"github.com/krane-io/krane/engine"
	"log"
	"os"
	"os/user"
	"strings"
	"unsafe"
)

const (
	defaultCaFile   = "ca.pem"
	defaultKeyFile  = "key.pem"
	defaultCertFile = "cert.pem"
)

var (
	kraneCertPath = os.Getenv("KRANE_CERT_PATH")
)

func init() {
	if kraneCertPath == "" {
		user, err := user.Current()
		if err != nil {
			log.Fatalln(err)
		}
		if user.Uid == "0" {
			kraneCertPath = "/etc/krane"
		} else {
			kraneCertPath = user.HomeDir + "/.krane"
		}
	}
}

var (
	flVersion            = flag.Bool([]string{"v", "-version"}, false, "Print version information and quit")
	flDaemon             = flag.Bool([]string{"d", "-daemon"}, false, "Enable daemon mode")
	flGraphOpts          opts.ListOpts
	flDebug              = flag.Bool([]string{"D", "-debug"}, false, "Enable debug mode")
	flAutoRestart        = flag.Bool([]string{"r", "-restart"}, true, "Restart previously running containers")
	bridgeName           = flag.String([]string{"b", "-bridge"}, "", "Attach containers to a pre-existing network bridge\nuse 'none' to disable container networking")
	bridgeIp             = flag.String([]string{"#bip", "-bip"}, "", "Use this CIDR notation address for the network bridge's IP, not compatible with -b")
	pidfile              = flag.String([]string{"p", "-pidfile"}, "/var/run/krane.pid", "Path to use for daemon PID file")
	flRoot               = flag.String([]string{"g", "-graph"}, "/var/lib/krane", "Path to use as the root of the Krane runtime")
	flSocketGroup        = flag.String([]string{"G", "-group"}, "krane", "Group to assign the unix socket specified by -H when running in daemon mode\nuse '' (the empty string) to disable setting of a group")
	flEnableCors         = flag.Bool([]string{"#api-enable-cors", "-api-enable-cors"}, false, "Enable CORS headers in the remote API")
	flDns                = opts.NewListOpts(opts.ValidateIPAddress)
	flDnsSearch          = opts.NewListOpts(opts.ValidateDnsSearch)
	flEnableIptables     = flag.Bool([]string{"#iptables", "-iptables"}, true, "Enable Krane's addition of iptables rules")
	flEnableIpForward    = flag.Bool([]string{"#ip-forward", "-ip-forward"}, true, "Enable net.ipv4.ip_forward")
	flDefaultIp          = flag.String([]string{"#ip", "-ip"}, "0.0.0.0", "Default IP address to use when binding container ports")
	flInterContainerComm = flag.Bool([]string{"#icc", "-icc"}, true, "Enable inter-container communication")
	flGraphDriver        = flag.String([]string{"s", "-storage-driver"}, "", "Force the Krane runtime to use a specific storage driver")
	flExecDriver         = flag.String([]string{"e", "-exec-driver"}, "native", "Force the Krane runtime to use a specific exec driver")
	flHosts              = opts.NewListOpts(api.ValidateHost)
	flMtu                = flag.Int([]string{"#mtu", "-mtu"}, 0, "Set the containers network MTU\nif no value is provided: default to the default route MTU or 1500 if no default route is available")
	flTls                = flag.Bool([]string{"-tls"}, false, "Use TLS; implied by tls-verify flags")
	flTlsVerify          = flag.Bool([]string{"-tlsverify"}, false, "Use TLS and verify the remote (daemon: verify client, client: verify daemon)")
	flSelinuxEnabled     = flag.Bool([]string{"-selinux-enabled"}, false, "Enable selinux support. SELinux does not presently support the BTRFS storage driver")

	// these are initialized in init() below since their default values depend on kraneCertPath which isn't fully initialized until init() runs
	flCa   *string
	flCert *string
	flKey  *string
)

func main() {
	configuration := config.LoadConfigurationFile()

	flag.Parse()

	if *flVersion {
		showVersion()
		return
	}

	if *flDebug {
		os.Setenv("DEBUG", "1")
	}

	if flHosts.Len() == 0 {
		defaultHost := os.Getenv("DOCKER_HOST")

		if os.Getenv("KRANE_HOST") != "" {
			defaultHost = os.Getenv("KRANE_HOST")
		}

		if unsafe.Sizeof(configuration) > 0 {
			defaultHost = fmt.Sprintf("%s:%d", configuration.Production.Server.Host.Fqdn, configuration.Production.Server.Host.Port)
		} else if defaultHost == "" || *flDaemon {
			// If we do not have a host, default to unix socket
			defaultHost = fmt.Sprintf("unix://%s", api.DEFAULTUNIXSOCKET)
		}
		if _, err := api.ValidateHost(defaultHost); err != nil {
			log.Fatal(err)
		}
		flHosts.Set(defaultHost)
	}

	if *flDaemon {

		if configuration.Production.Server.Driver == "concerto" {
			configuration.Driver = concerto.NewDriver()
		} else if configuration.Production.Server.Driver == "aws" {
			configuration.Driver = aws.NewDriver()
		}

		engine.InitializeDockerEngine(configuration)
		return
	}

	if flHosts.Len() > 1 {
		log.Fatal("Please specify only one -H")
	}

	protoAddrParts := strings.SplitN(flHosts.GetAll()[0], "://", 2)

	var (
		cli *client.KraneCli
	)

	cli = client.NewKraneCli(os.Stdin, os.Stdout, os.Stderr, nil, protoAddrParts[0], protoAddrParts[1], nil)

	if err := cli.Cmd(flag.Args()...); err != nil {
		if sterr, ok := err.(*utils.StatusError); ok {
			if sterr.Status != "" {
				log.Println(sterr.Status)
			}
			os.Exit(sterr.StatusCode)
		}
		log.Fatal(err)
	}

}

// TODO: We need to do version similar to https://github.com/docker/docker/blob/master/dockerversion/dockerversion.go
func showVersion() {
	fmt.Printf("Krane version .......\n")
}
