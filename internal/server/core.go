package server

import (
	"fmt"
	"log"
	"time"

	"github.com/spin-org/thermomatic/internal/client"
	"github.com/spin-org/thermomatic/internal/common"
)

// core mantains a map of clients and communication channels
type core struct {
	clients  map[uint64]*client.Client
	commands chan common.Command
	Logouts  chan *client.Client
	Logins   chan *client.Client
}

// NewCore allocates a Core struct
func newCore() *core {
	return &core{
		Logins:   make(chan *client.Client),
		Logouts:  make(chan *client.Client),
		clients:  make(map[uint64]*client.Client),
		commands: make(chan common.Command),
	}
}

// Run handles channels inbound communications from connected clients
func (c *core) run() {
	for {
		select {
		case client := <-c.Logins:
			c.register(client)
		case client := <-c.Logouts:
			c.deregister(client)
		case cmd := <-c.commands:
			switch cmd.ID {
			case common.READING:
				c.handleReading(cmd.Sender, cmd.Body)
			default:
				log.Printf("Unknown Command %d", cmd.ID)
			}
		}
	}
}

func (c *core) handleReading(imei uint64, payload []byte) {
	if device, exists := c.clients[imei]; exists {
		reading := &client.Reading{}
		reading.Decode(payload)
		device.LastReadingEpoch = time.Now().UnixNano()
		device.LastReading = reading
		fmt.Println(device)
	}

}

func (c *core) register(device *client.Client) {
	if _, exists := c.clients[device.IMEI]; exists {
		//TODO strategy for what should happen when a device attempts to login twice #11
		log.Printf("ERR imei %d already logged in", device.IMEI)
		device.Conn.Close()
	} else {
		c.clients[device.IMEI] = device
		log.Printf("device with IMEI %d connected succesfuly", device.IMEI)
	}
}

func (c *core) deregister(device *client.Client) {
	if _, exists := c.clients[device.IMEI]; exists {
		delete(c.clients, device.IMEI)
	}
}
