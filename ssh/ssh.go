package ssh

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"github.com/krane-io/krane/config"
	"io"
	"io/ioutil"
	"net"
	"os"
	"time"

	dockerEngine "github.com/docker/docker/engine"
)

func getPrivateKeys(job *dockerEngine.Job) ssh.Signer {
	privateBytes, err := ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa")

	if err != nil {
		job.Errorf("Failed to load private key")
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		job.Errorf("Failed to parse private key")
	}

	return private
}

func ioProxy(conn1 net.Conn, conn2 net.Conn) {
	go io.Copy(conn1, conn2)
	go io.Copy(conn2, conn1)
}

func portMapping(job *dockerEngine.Job, remoteHost string, localPort int, remotePort int) {
	localListener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", localPort))
	if err != nil {
		job.Errorf("\nnet.Listen failed: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(getPrivateKeys(job)),
		},
	}

	// Dial your ssh server.
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", remoteHost), config)
	if err != nil {
		job.Errorf("\nUnable to connect: %s with %s", err, remoteHost)
	} else {
		job.Logf("\nEstablish ssh tunnel with %s:22:%d", remoteHost, localPort)
	}

	defer conn.Close()

	for {
		// Setup localConn (type net.Conn)
		localConnection, err := localListener.Accept()
		if err != nil {
			job.Errorf("\nListen.Accept failed: %v", err)
		}
		defer localConnection.Close()

		remoteConnection, err := conn.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", remotePort))
		if err != nil {
			job.Errorf("\nUnable to register tcp forward: %v", err)
		}
		defer remoteConnection.Close()

		go ioProxy(localConnection, remoteConnection)
	}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("Recovered in f", r)
		}
	}()
}

func Tunnel(job *dockerEngine.Job) dockerEngine.Status {
	var fqdn string
	var delayed string
	if len(job.Args) == 2 {
		fqdn = job.Args[0]
		delayed = job.Args[1]
	} else if len(job.Args) > 2 {
		return job.Errorf("Usage: %s", job.Name)
	}

	configuration := job.Eng.Hack_GetGlobalVar("configuration").(config.KraneConfiguration)

	if delayed == "true" {
		job.Logf("\nWe are going to waiting 30 seconds to create ssh tunnel with %s", fqdn)
		time.Sleep(30 * time.Second)
		ships, err := configuration.Driver.List(nil)
		fmt.Printf("%#v", ships)
		if err != nil {
			job.Logf("\nUnable to get list of ships from %s", configuration.Driver.Name())
		}
		configuration.UpdateShips(ships)
		job.Eng.Hack_SetGlobalVar("configuration", configuration)
		fmt.Printf("%#v", configuration.Production.Fleet)
	}

	if configuration.Production.HighPort == 0 {
		configuration.Production.HighPort = 8000
		job.Eng.Hack_SetGlobalVar("configuration", configuration)
	}

	ship := configuration.GetShip(fqdn)

	if (ship.State == "operational") && (ship.LocalPort == 0) {
		job.Logf("\nCreating ssh tunnel for %s\n", fqdn)
		configuration.Production.HighPort = configuration.Production.HighPort + 1
		ship.LocalPort = configuration.Production.HighPort
		configuration.UpdateShip(ship)
		job.Eng.Hack_SetGlobalVar("configuration", configuration)
		fmt.Printf("%#v", configuration.Production.Fleet)
		go portMapping(job, ship.Fqdn, ship.LocalPort, ship.Port)
		return dockerEngine.StatusOK

	} else {
		job.Logf("\nGoing to queue job to create tunnel for %s\n", fqdn)
		newjob := job.Eng.Job("ssh_tunnel", fqdn, "true")
		if delayed == "true" {
			newjob.Run()
		} else {
			go newjob.Run()
		}
		return dockerEngine.StatusOK
	}
}
