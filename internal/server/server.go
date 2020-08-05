package server

import (
	"fmt"
	"log"
	"net"
	"time"

	"github.com/spin-org/thermomatic/internal/client"
)

// Start creates a tcp connection listener to accept connections at `port`
func Start(port uint, httpPort uint, serverMaxClients uint) {
	address := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", address)
	log.Printf("Server started, using %s as address", address)
	if err != nil {
		log.Fatalf("%v", err)
	}

	core := newCore()
	httpd := newHttpd(core, httpPort)
	go core.run()
	go httpd.run()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("%v", err)
		}
		numActiveClients := len(core.clients)
		if uint(numActiveClients) >= serverMaxClients {
			// Limit the number of active clients to prevent resource exhaustion
			log.Printf("ERR reached serverMaxClients:%d, there are already %d connected clients", serverMaxClients, numActiveClients)
			conn.Close()
		} else {
			log.Printf("client connection from %v", conn.RemoteAddr())
			//if the device fail to send the login message within 1 second the server will drop the client connection.
			conn.SetReadDeadline(time.Now().Add(time.Second))

			c := client.NewClient(
				conn,
				core.commands,
				core.Logins,
				core.Logouts,
			)
			go c.Read()
		}

	}

}
