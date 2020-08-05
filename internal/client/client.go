package client

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

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
	login            chan<- *Client
	logout           chan<- *Client
}

// NewClient allocates a Client
func NewClient(conn net.Conn, outbound chan<- common.Command, login chan<- *Client, logout chan<- *Client) *Client {
	return &Client{
		Conn:     conn,
		outbound: outbound,
		login:    login,
		logout:   logout,
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
	c.login <- c
	return nil
}

func (c *Client) receiveReadings() error {
	//TODO verify if this is better than using make to allocate this
	var payload [40]byte
	for {
		n, err := c.Conn.Read(payload[:])
		if err != nil {
			if err == io.EOF {
				// deregister client when Connection is closed
				c.logout <- c
				return nil
			}
			return err
		}

		if n != 40 {
			log.Printf("WARN read only %d bytes, expected 40 for reading payloads", n)
		}
		c.outbound <- common.Command{
			ID:     common.READING,
			Sender: c.IMEI,
			Body:   payload[:],
		}

		c.Conn.SetReadDeadline(time.Now().Add(time.Second * 2))
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
