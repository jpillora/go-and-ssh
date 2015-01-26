// A simple SSH server providing bash sessions
package main

import (
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

func main() {
	// Simple server config with hard-coded key
	config := &ssh.ServerConfig{NoClientAuth: true}
	private, _ := ssh.ParsePrivateKey([]byte(key))
	config.AddHostKey(private)

	// Once a ServerConfig has been configured, connections can be accepted.
	listener, err := net.Listen("tcp", "0.0.0.0:2200")
	if err != nil {
		log.Fatal("Failed to listen on 2200")
	}

	// Accept all connections
	log.Print("Listening on 2200...")
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
		go ssh.DiscardRequests(reqs)
		// Accept all channels
		go handleChannels(chans)
	}
}

func handleChannels(chans <-chan ssh.NewChannel) {
	for newChannel := range chans {
		go handleChannel(newChannel)
	}
}

func handleChannel(newChannel ssh.NewChannel) {

	channel, requests, err := newChannel.Accept()
	if err != nil {
		log.Printf("could not accept channel (%s)", err)
		return
	}

	chanType := newChannel.ChannelType()
	extraData := newChannel.ExtraData()

	log.Printf("open channel [%s] '%s'", chanType, extraData)

	//requests must be serviced
	go ssh.DiscardRequests(requests)

	//channel
	buff := make([]byte, 256)
	for {
		n, err := channel.Read(buff)
		if err != nil {
			break
		}
		b := buff[:n]
		log.Printf("[%s] %s", chanType, string(b))
	}
}

//dont do this IRL :)
const key = `
-----BEGIN RSA PRIVATE KEY-----
MIIEowIBAAKCAQEAzNO5vZPpP7WgXA3Ck5NeCq85i1v2JCB5vM0udK+oWrCQpMdy
oKZlxC8z8n/mSsylm+2xEm+kAFxyvB9ae/Pr8Lh0czePw473Qx2v78E/HdouXn3w
xEHG12IoDUdC7Rt4faxNdfsebd/wWybHEV6vOEDDkxmppJ1y6Cbgx6a59X0wqW54
bTKy5D98iLMzSvWi6AUS3I/hP53f7mNK7cTPqHTdVOwICgCGHOI1hcDKwMafj590
+3H/F5ACYRl9Keuij09zsk+QkI+7HJN5HUtq9mjJ9Mw4vo9LzqIWTOWncEvX5b2f
99GOOlsBNh91L3PNwQdf1M++CM6F0HTv5p8ioQIDAQABAoIBACwruJF2dUWE8IkJ
ep2CmTQqp3kzIriVvEsH4G3Pd7ne+8JdNI4KdEXDfCteg5Y73bbrolT8eFyPkzqY
dFXou0fVL1+tarZcfVwe6dMFVIwmgftko2hfWvcVttduN7OUSf6oCqhXuC8vrNCr
YyCOz7CM3uA5F4llXuNLhwvnG5EhxHk/AVN0SUbJbfKD5DEpqFM33PuITAuIPuSi
Td2qa84WitZ12hBJqtZGngujE/bMZNaY0Lk6EM4L2p47+//z3raScQT2B+eF/LnR
Jn32YaI7np7Y4D7RbW6QZBB/sOkrvtX51tIHIQEYdn4zlfT8+tNeVo9jn0QM77Ky
FcY4a8ECgYEA5vF+P5MeSa+QsUVgK3HY4MuNNRKw4daIFJr/keYLyUwfPYQsdu5V
ZXfJPkQ/y1Xlgek6E/eiiaiJN91hZEkoF6fkXcORCCmjr19FfssC++arTKk/UPxT
y946yFscsZXosssCON7CskGLCiPMn7YwdwQiJ9uvKIxwB2ChfJ/trSkCgYEA4wzY
rp5Pz3lbXg6P7xqYibnIH847PW9GVMGNl6pXfhUkP3NqFD+Oc41S/wD/vv1SVSZ7
2ih56E7vctxtxc9b5wWcZfzRUbBWrSKwWO1ImqsBdFapxtoOynDL0uHnXaDrQCvW
UsI44d92gmO+MMYst9//I/sLRTrwYrrIvJOVALkCgYEAg0uqVeSDJKtOnKnveeOY
xHyVBCZjL5Hy/Zv9Tmo2KzQ+0o9xZBAttqk6XU8Z4bUs7QW2giGYY6DQmlUfCI/a
3lASMgh8TOK3b32/mc07HhFPNB9IovdBgLcQPlYmYwPyLqvh0Ik8sXE35gTiUa6X
sSJFdNmdpHTrQBZ82MhnrLkCgYB8wG06HKALhkmOd3/cR4eyfNKZry3bho1lOmf7
AkxKaYFeH6MUdwtlMCx/EmRy4ytev+NjLcQ1wVFNkhH6kwGTAQE7BFtagAJP5PRy
GAZBfV4yNv/X0642yx0ixJ7kUeuQecWr+S1Z5fdukzFICUs+yKOeeGxr4IN+K9Tp
0EkZeQKBgF58RcI6PZD7mayf0Z58gd+zb2WXL1rTGErYsbVgxkc/TFRaZYK0cb+n
V6WZNy6k5Amx54pv59U34sEiGqFb8xo9Q0o+jcdrirTJKvuJuGh5Hm/4jjRvu4O3
1Qr6yBnUTsDcXkDy8G0oenhDMceZEbIz+WOqmxKx7eGl0OxE0CNt
-----END RSA PRIVATE KEY-----
`
