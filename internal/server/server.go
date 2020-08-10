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
	if err != nil {
		log.Fatalf("ERR Failed to start tcp listener at %s,  %v", address, err)
	}

	log.Printf("Server started, using %s as address", address)
	core := newCore(time.Now)
	httpd := newHttpd(core, httpPort)
	go core.run()
	go httpd.run()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		numActiveClients := len(core.clients)
		if uint(numActiveClients) >= serverMaxClients {
			// Limit the number of active clients to prevent resource exhaustion
			log.Printf("ERR reached serverMaxClients:%d, there are already %d connected clients", serverMaxClients, numActiveClients)
			conn.Close()
		} else {
			log.Printf("client connection from %v", conn.RemoteAddr())
			//if the device fail to send the login message within 1 second the server will drop the client connection.
			err = conn.SetReadDeadline(time.Now().Add(time.Second))
			if err != nil {
				log.Fatalf("ERR trying to set read timeout of 1 sec for the login message %v", err)
			}

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
