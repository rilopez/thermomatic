package server

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/spin-org/thermomatic/internal/common"
	"github.com/spin-org/thermomatic/internal/device"
)

// core mantains a map of clients and communication channels
type core struct {
	devices          map[uint64]*connectedDevice
	commands         chan common.Command
	port             uint
	serverMaxClients uint
	now              func() time.Time
}

type connectedDevice struct {
	callbackChannel  chan common.Command
	lastReadingEpoch int64
	lastReading      *device.Reading
}

// NewCore allocates a Core struct
func newCore(now func() time.Time, port uint, serverMaxClients uint) *core {
	return &core{
		devices:          make(map[uint64]*connectedDevice),
		commands:         make(chan common.Command, 1),
		now:              now,
		port:             port,
		serverMaxClients: serverMaxClients,
	}
}

func (c *core) listenConnections(wg *sync.WaitGroup) {
	address := fmt.Sprintf(":%d", c.port)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalf("ERR Failed to start tcp listener at %s,  %v", address, err)
	}
	defer func() {
		ln.Close()
		wg.Done()
	}()

	log.Printf("Server started listening for connections at %s ", address)
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("Failed to accept connection: %v", err)
			continue
		}
		log.Print("DEBUG: new client connection")
		//TODO #20 accesing c.clients in this way is not thread safe, use a mutex
		numActiveClients := len(c.devices)
		if uint(numActiveClients) >= c.serverMaxClients {
			// Limit the number of active clients to prevent resource exhaustion
			log.Printf("ERR reached serverMaxClients:%d, there are already %d connected clients", c.serverMaxClients, numActiveClients)
			conn.Close()

		} else {
			log.Printf("client connection from %v", conn.RemoteAddr())
			//if the device fail to send the login message within 1 second the server will drop the client connection.
			err := conn.SetReadDeadline(time.Now().Add(time.Second))
			if err != nil {
				log.Printf("trying to set read timeout of 1 sec for the login message %v", err)
				conn.Close()
				continue
			}
			client, err := device.NewClient(
				conn,
				c.commands,
				c.now,
			)
			if err != nil {
				conn.Close()
				log.Printf("ERR trying to create a client worker for the connection, %v", err)
				continue
			}
			wg.Add(1)
			go client.Read(wg)
		}

	}

}

// Run handles channels inbound communications from connected clients
func (c *core) run(wg *sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()

	go c.listenConnections(wg)

	for {
		var err error
		select {
		case cmd := <-c.commands:
			switch cmd.ID {
			case common.LOGIN:
				err = c.register(cmd.Sender, cmd.CallbackChannel)
			case common.LOGOUT:
				err = c.deregister(cmd.Sender)
			case common.READING:
				err = c.handleReading(cmd.Sender, cmd.Body)
			default:
				err = fmt.Errorf("Unknown Command %d", cmd.ID)
			}
		}
		if err != nil {
			log.Printf("ERR %v", err)
		}

	}

}

func (c *core) handleReading(imei uint64, payload []byte) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("ERR recovering from hadleReading panic %v", r)
		}
	}()
	dev, exists := c.devices[imei]
	if !exists {
		return fmt.Errorf("Client with IMEI %d does not exists", imei)
	}
	reading := &device.Reading{}
	if !reading.Decode(payload) {
		return fmt.Errorf("ERR decoding payload from device with IMEI %d", imei)
	}

	dev.lastReadingEpoch = c.now().UnixNano()
	dev.lastReading = reading

	fmt.Println(formatReadingOutput(imei, dev.lastReadingEpoch, dev.lastReading))

	return nil
}

func formatReadingOutput(imei uint64, lastReadingEpoch int64, lastReading *device.Reading) string {
	return fmt.Sprintf("%d,%d,%f,%f,%f,%f,%f",
		lastReadingEpoch,
		imei,
		lastReading.Temperature,
		lastReading.Altitude,
		lastReading.Latitude,
		lastReading.Longitude,
		lastReading.BatteryLevel)

}

func (c *core) register(imei uint64, callbackChannel chan common.Command) error {
	if _, exists := c.devices[imei]; exists {
		log.Printf("DEBUG trying to kill connected dup device %v", imei)
		callbackChannel <- common.Command{ID: common.KILL}
		log.Printf("DEBUG KILL cmd sent  %v", imei)
		return fmt.Errorf("imei %d already logged in", imei)
	}
	c.devices[imei] = &connectedDevice{
		callbackChannel: callbackChannel,
	}
	callbackChannel <- common.Command{ID: common.WELCOME}
	log.Printf("device with IMEI %d connected succesfuly", imei)

	return nil
}

func (c *core) deregister(imei uint64) error {
	log.Printf("DEBUG trying to deregister device with IMEI %d ", imei)
	_, exists := c.devices[imei]
	if !exists {
		return fmt.Errorf("ERR imei %d is not logged in", imei)
	}

	delete(c.devices, imei)
	log.Printf("device with IMEI %d desconnected succesfuly", imei)
	return nil
}
