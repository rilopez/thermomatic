// Package client provides handlers and channels to process client device connections
package client

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/spin-org/thermomatic/internal/imei"

	"github.com/spin-org/thermomatic/internal/common"
)

// Client is used to handle a client connection
type Client struct {
	IMEI             uint64
	Conn             net.Conn
	LastReading      *Reading
	LastReadingEpoch int64
	outbound         chan<- common.Command
	register         chan<- *Client
	deregister       chan<- *Client
}

// NewClient allocates a Client
func NewClient(conn net.Conn, o chan<- common.Command, r chan<- *Client, d chan<- *Client) *Client {
	return &Client{
		Conn:       conn,
		outbound:   o,
		register:   r,
		deregister: d,
	}
}

//TODO create a test for client string
func (c *Client) String() string {
	return fmt.Sprintf("%d,%d,%f,%f,%f,%f,%f",
		c.LastReadingEpoch,
		c.IMEI, c.LastReading.Temperature,
		c.LastReading.Altitude,
		c.LastReading.Latitude,
		c.LastReading.Longitude,
		c.LastReading.BatteryLevel)
}

//TODO code polish: split login and reading reads in separate testeable functions
func (c *Client) Read() error {
	log.Println("starting read client connection handler")
	if c.IMEI == 0 {
		log.Println("attempting to read IMEI ")
		var loginMsg [15]byte
		_, err := c.Conn.Read(loginMsg[:])
		//TODO  verify bytes read should be 15
		if err != nil {
			log.Printf("ERR trying to read IMEI, %v ", err)
			return err
		}

		imei, err := imei.Decode(loginMsg[:])
		if err != nil {
			log.Printf("ERR decoding IMEI bytes %v ", err)
			return err
		}
		c.IMEI = imei
		c.register <- c
	}
	//TODO verify if this is better than using make to allocate this
	var payload [40]byte
	for {
		_, err := c.Conn.Read(payload[:])
		//TODO  verify bytes read should be 40, explain why?
		if err != nil {
			log.Printf("ERR reading payload bytes %v ", err)
			if err == io.EOF {
				// deregister client when Connection is closed
				c.deregister <- c
				return nil
			}
			return err
		}

		c.handleReading(payload)
		//TODO Drops client connections which fail to send at least a _Reading_ every _2 seconds_. #5
	}
}

func (c *Client) handleReading(payload [40]byte) {
	log.Printf("DEBUG reading payload bytes\n%v", hex.Dump(payload[:]))
	//TODO Gathers & Output `Reading` messages #4 , this is untested code
	c.outbound <- common.Command{
		ID:     common.READING,
		Sender: c.IMEI,
		Body:   payload[:],
	}

}
