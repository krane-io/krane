package client

import (
	"crypto/tls"
	"fmt"

	dockerApiClient "github.com/docker/docker/api/client"
	flag "github.com/docker/docker/pkg/mflag"
	"io"
	"os"
	"reflect"
	"strings"
)

func (cli *KraneCli) Subcmd(name, signature, description string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		fmt.Fprintf(cli.Err(), "\nUsage: krane %s %s\n\n%s\n\n", name, signature, description)
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

func NewKraneCli(in io.ReadCloser, out, err io.Writer, proto, addr string, tlsConfig *tls.Config) *KraneCli {
	return &KraneCli{*dockerApiClient.NewDockerCli(in, out, err, proto, addr, tlsConfig)}
}

type KraneCli struct {
	dockerApiClient.DockerCli
}
