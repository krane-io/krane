package ssh

import (
	"code.google.com/p/go.crypto/ssh"
	"fmt"
	"io"
	"io/ioutil"

	"net"
	"os"

	dockerEngine "github.com/docker/docker/engine"

	"github.com/krane-io/krane/hacks"
)

func getPrivateKeys() ssh.Signer {
	privateBytes, err := ioutil.ReadFile(os.Getenv("HOME") + "/.ssh/id_rsa")

	if err != nil {
		panic("Failed to load private key")
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		panic("Failed to parse private key")
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
		job.Errorf("net.Listen failed: %v", err)
	}

	config := &ssh.ClientConfig{
		User: "root",
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(getPrivateKeys()),
		},
	}

	// Dial your ssh server.
	conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", remoteHost), config)
	if err != nil {
		job.Errorf("unable to connect: %s", err)
	} else {
		job.Logf("Establish ssh tunnel with %s:22", remoteHost)
	}

	defer conn.Close()

	for {
		// Setup localConn (type net.Conn)
		localConnection, err := localListener.Accept()
		if err != nil {
			job.Errorf("listen.Accept failed: %v", err)
		}
		defer localConnection.Close()

		remoteConnection, err := conn.Dial("tcp", "127.0.0.1:2375")
		if err != nil {
			job.Errorf("unable to register tcp forward: %v", err)
		}
		defer remoteConnection.Close()

		go ioProxy(localConnection, remoteConnection)
	}
}

func Tunnel(job *dockerEngine.Job) dockerEngine.Status {

	configuration := hacks.DockerGetGlobalConfig(job)

	for index := 0; index < len(configuration.Production.Fleet); index++ {
		configuration.Production.Fleet[index].LocalPort = 8000 + index
		ship := configuration.Production.Fleet[index]
		go portMapping(job, ship.Fqdn, ship.LocalPort, ship.Port)
	}

	hacks.DockerSetGlobalConfig(job, configuration)

	return dockerEngine.StatusOK
}
