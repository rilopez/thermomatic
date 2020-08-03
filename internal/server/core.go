package server

import (
	"log"

	"github.com/spin-org/thermomatic/internal/client"
	"github.com/spin-org/thermomatic/internal/common"
)

// Core mantains a map of clients and communication channels
type Core struct {
	clients  map[uint64]*client.Client
	Commands chan common.Command
	//TODO rename to logouts
	Deregistrations chan *client.Client
	//TODO rename to logins
	Registrations chan *client.Client
}

// NewCore allocates a Core struct
func NewCore() *Core {
	return &Core{
		Registrations:   make(chan *client.Client),
		Deregistrations: make(chan *client.Client),
		clients:         make(map[uint64]*client.Client),
		Commands:        make(chan common.Command),
	}
}

// Run handles channels inbound communications from connected clients
func (c *Core) Run() {
	for {
		select {
		case client := <-c.Registrations:
			c.register(client)
		case client := <-c.Deregistrations:
			c.deregister(client)
		case cmd := <-c.Commands:
			switch cmd.ID {
			case common.READING:
				//TODO Gathers & Output `Reading` messages #4
				log.Printf("Reading not implemented yet")
			default:
				// Ka booomn?
			}
		}
	}
}

func (c *Core) register(device *client.Client) {
	if _, exists := c.clients[device.IMEI]; exists {
		//TODO strategy for what should happen when a device attempts to login twice #11
		log.Printf("ERR imei %d already logged in", device.IMEI)
		device.Conn.Close()
	} else {
		c.clients[device.IMEI] = device
		log.Printf("device with IMEI %d connected succesfuly", device.IMEI)
	}
}

func (c *Core) deregister(device *client.Client) {
	if _, exists := c.clients[device.IMEI]; exists {
		delete(c.clients, device.IMEI)
	}
}