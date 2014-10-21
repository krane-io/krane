package main

import (
	"log"
	"os"
	"os/user"

	"github.com/docker/docker/api"
	"github.com/docker/docker/opts"
	flag "github.com/docker/docker/pkg/mflag"
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
