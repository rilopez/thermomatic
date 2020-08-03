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

func (c *Client) String() string {
	return fmt.Sprintf("%d,%d,%f,%f,%f,%f,%f",
		c.LastReadingEpoch,
		c.IMEI, c.LastReading.Temperature,
		c.LastReading.Altitude,
		c.LastReading.Latitude,
		c.LastReading.Longitude,
		c.LastReading.BatteryLevel)
}

func (c *Client) receiveLoginMessage() error {
	log.Println("attempting to read IMEI ")
	var loginMsg [15]byte
	n, err := c.Conn.Read(loginMsg[:])
	if err != nil || n < 15 {
		log.Printf("ERR trying to read IMEI, bytes read: %d, err: %v", n, err)
		return err
	}

	imei, err := imei.Decode(loginMsg[:])
	if err != nil {
		log.Printf("ERR decoding IMEI bytes %v ", err)
		return err
	}
	c.IMEI = imei
	//c.register <- c
	return nil
}

func (c *Client) receiveReadings() error {
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

		log.Printf("DEBUG reading payload bytes\n%v", hex.Dump(payload[:]))

		c.outbound <- common.Command{
			ID:     common.READING,
			Sender: c.IMEI,
			Body:   payload[:],
		}

		//TODO Drops client connections which fail to send at least a _Reading_ every _2 seconds_. #5
	}
}

func (c *Client) Read() error {
	log.Println("starting read client connection handler")
	if c.IMEI == 0 {
		err := c.receiveLoginMessage()
		if err != nil {
			return err
		}
	}
	return c.receiveReadings()
}
