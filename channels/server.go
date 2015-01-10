// A simple SSH server providing bash sessions
package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os/exec"
	"sync"

	"github.com/kr/pty"
	"golang.org/x/crypto/ssh"
)

func main() {
	config := &ssh.ServerConfig{
		NoClientAuth: true,
	}

	// You can generate a keypair with 'ssh-keygen -t rsa'
	privateBytes, err := ioutil.ReadFile("id_rsa")
	if err != nil {
		log.Fatal("Failed to load private key (./id_rsa)")
	}

	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatal("Failed to parse private key")
	}

	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2022")
	if err != nil {
		log.Fatal("Failed to listen on 2022")
	}

	// Accept all connections
	log.Print("Listening on 2022...")
	for {
		tcpConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection (%s)", err)
			continue
		}
		// Before use, a handshake must be performed on the incoming net.Conn.
		sshConn, chans, reqs, err := ssh.NewServerConn(tcpConn, config)
		if err != nil {
			log.Printf("Failed to handshake (%s)", err)
			continue
		}

		log.Printf("New SSH connection from %s (%s)", sshConn.RemoteAddr(), sshConn.ClientVersion())
		// Discard all global out-of-band Requests
		go handleRequests(reqs)
		// Accept all channels
		go handleChannels(chans)
	}
}

func handleRequests(reqs <-chan *ssh.Request) {
	for req := range reqs {
		log.Printf("Recieved global request: %+v", req)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	// Service the incoming Channel channel.
	for newChannel := range chans {
		// Since we're handling the execution of a shell, we expect a
		// channel type of "session". However, there are also: "x11", "direct-tcpip"
		// and "forwarded-tcpip" channel types.
		if t := newChannel.ChannelType(); t != "session" {
			newChannel.Reject(ssh.UnknownChannelType, fmt.Sprintf("unknown channel type: %s", t))
			continue
		}

		// At this point, we have the opportunity to reject the client's
		// request for another logical connection
		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("could not accept channel (%s)", err)
			continue
		}
		// fire up bash for this session
		bash := exec.Command("bash")
		// allocate a terminal for this channel
		log.Print("creating pty...")
		bashf, err := pty.Start(bash)
		if err != nil {
			log.Printf("could not start pty (%s)", err)
			continue
		}

		//teardown session
		var once sync.Once
		close := func() {
			channel.Close()
			_, err := bash.Process.Wait()
			if err != nil {
				log.Printf("failed to exit bash (%s)", err)
			}
			log.Printf("session closed")
		}

		//pipe session to bash and visa-versa
		go func() {
			io.Copy(channel, bashf)
			once.Do(close)
		}()
		go func() {
			io.Copy(bashf, channel)
			once.Do(close)
		}()

		// Sessions have out-of-band requests such as "shell", "pty-req" and "env"
		go func(in <-chan *ssh.Request) {
			for req := range in {
				ok := false
				switch req.Type {
				case "shell":
					// We don't accept any commands (Payload),
					// only the default shell.
					if len(req.Payload) == 0 {
						ok = true
					}
				case "pty-req":
					// Responding 'ok' here will let the client
					// know we have a pty ready for input
					ok = true
					// Parse body...
					termLen := req.Payload[3]
					termEnv := string(req.Payload[4 : termLen+4])
					w, h := parseDims(req.Payload[termLen+4:])
					SetWinsize(bashf.Fd(), w, h)
					log.Printf("pty-req '%s'", termEnv)
				case "window-change":
					w, h := parseDims(req.Payload)
					SetWinsize(bashf.Fd(), w, h)
					continue //no response
				}
				if !ok {
					log.Printf("declining %s request...", req.Type)
				}
				req.Reply(ok, nil)
			}
		}(requests)
	}
}
