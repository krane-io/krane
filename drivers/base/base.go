package main

import (
	"fmt"

	flag "github.com/docker/docker/pkg/mflag"
	"io"
	"os"
	"reflect"
	"strings"
)

type Driver interface {
	CmdList(args ...string) error
	CmdCreate(args ...string) error
	CmdStart(args ...string) error
	CmdPlans(args ...string) error
	CmdStop(args ...string) error
	CmdRestart(args ...string) error
	CmdConf(args ...string) error
	CmdRm(args ...string) error
}

type Base struct {
	name string
	in   io.ReadCloser
	out  io.Writer
	err  io.Writer
}

func (driver *Base) CmdList(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdCreate(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdStart(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdPlans(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdStop(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdRestart(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdConf(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdRm(args ...string) error {
	fmt.Printf("{ error: \"Not Implemented please update driver %s\" }\n", driver.name)
	return nil
}

func (driver *Base) CmdHelp(args ...string) error {
	if len(args) > 0 {
		method, exists := driver.getMethod(args[0])
		if !exists {
			fmt.Fprintf(driver.err, "Error: Command not found: %s\n", args[0])
		} else {
			method("--help")
			return nil
		}
	}
	help := fmt.Sprintf("Usage: %s [OPTIONS] COMMAND [arg...]\n\nCommands:\n", driver.name)
	for _, command := range [][]string{
		{"list", "List ships"},
		{"create", "Creates a Ship"},
		{"rm", "Remove a ship"},
		{"start", "Starts a ship"},
		{"plans", "Starts a ship"},
		{"stop", "Stops a ship"},
		{"restart", "Restarts a ship"},
		{"conf", "Displays Configuration"},
	} {
		help += fmt.Sprintf("    %-10.10s%s\n", command[0], command[1])
	}
	fmt.Fprintf(driver.err, "%s\n", help)
	return nil
}

func (driver *Base) getMethod(name string) (func(...string) error, bool) {
	if len(name) == 0 {
		return nil, false
	}
	methodName := "Cmd" + strings.ToUpper(name[:1]) + strings.ToLower(name[1:])
	method := reflect.ValueOf(driver).MethodByName(methodName)
	if !method.IsValid() {
		return nil, false
	}
	return method.Interface().(func(...string) error), true
}

func (driver *Base) Subcmd(name, signature, description string) *flag.FlagSet {
	flags := flag.NewFlagSet(name, flag.ContinueOnError)
	flags.Usage = func() {
		options := ""
		fmt.Fprintf(driver.err, "\nUsage: docker %s %s%s\n\n%s\n\n", name, options, signature, description)
		flags.PrintDefaults()
		os.Exit(2)
	}
	return flags
}

func (driver *Base) Cmd(args ...string) error {
	if len(args) > 0 {
		method, exists := driver.getMethod(args[0])
		if !exists {
			fmt.Println("Error: Command not found:", args[0])
			return driver.CmdHelp(args[1:]...)
		}
		return method(args[1:]...)
	}
	return driver.CmdHelp(args...)
}

func NewDriver() *Base {
	return &Base{name: "base", in: os.Stdin, out: os.Stdout, err: os.Stderr}
}

func main() {
	flag.Parse()
	driver := NewDriver()

	driver.Cmd(flag.Args()...)

}
