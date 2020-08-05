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

	now func() time.Time
}

// NewCore allocates a Core struct
func newCore(now func() time.Time) *core {
	return &core{
		Logins:   make(chan *client.Client),
		Logouts:  make(chan *client.Client),
		clients:  make(map[uint64]*client.Client),
		commands: make(chan common.Command),
		now:      now,
	}
}

// Run handles channels inbound communications from connected clients
func (c *core) run() {
	for {
		var err error
		select {
		case client := <-c.Logins:
			err = c.register(client)
		case client := <-c.Logouts:
			err = c.deregister(client)
		case cmd := <-c.commands:
			switch cmd.ID {
			case common.READING:
				err = c.handleReading(cmd.Sender, cmd.Body)
			default:
				err = fmt.Errorf("Unknown Command %d", cmd.ID)
			}
		}
		log.Print(err)
	}
}

func (c *core) handleReading(imei uint64, payload []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ERR recovering from hadleReading panic %v", r)
		}
	}()
	device, exists := c.clients[imei]
	if !exists {
		return fmt.Errorf("Client with IMEI %d does not exists", imei)
	}
	reading := &client.Reading{}
	if !reading.Decode(payload) {
		return fmt.Errorf("ERR decoding payload from device with IMEI %d", imei)
	}

	device.LastReadingEpoch = c.now().UnixNano()
	device.LastReading = reading
	return nil
}

func (c *core) register(device *client.Client) error {
	if _, exists := c.clients[device.IMEI]; exists {
		device.Conn.Close()
		return fmt.Errorf("ERR imei %d already logged in", device.IMEI)
	}
	c.clients[device.IMEI] = device
	log.Printf("device with IMEI %d connected succesfuly", device.IMEI)

	return nil
}

func (c *core) deregister(device *client.Client) error {
	if _, exists := c.clients[device.IMEI]; !exists {
		return fmt.Errorf("ERR imei %d is not logged in", device.IMEI)
	}
	log.Printf("device with IMEI %d desconnected succesfuly", device.IMEI)
	delete(c.clients, device.IMEI)
	return nil
}
