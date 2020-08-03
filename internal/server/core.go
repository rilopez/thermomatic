package server

import (
	"fmt"
	"log"
	"time"

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
				c.handleReading(cmd.Sender, cmd.Body)
			default:
				log.Printf("Unknown Command %d", cmd.ID)
			}
		}
	}
}

func (c *Core) handleReading(imei uint64, payload []byte) {
	if device, exists := c.clients[imei]; exists {
		reading := &client.Reading{}
		reading.Decode(payload)
		device.LastReadingEpoch = time.Now().UnixNano()
		device.LastReading = reading
		fmt.Println(device)
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
