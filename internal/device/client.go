package device

import (
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/spin-org/thermomatic/internal/common"
)

// Client is used to handle a client connection
type Client struct {
	imei     uint64
	conn     net.Conn
	outbound chan<- common.Command
	inbound  chan common.Command
	now      func() time.Time
}

// NewClient allocates a Client
func NewClient(conn net.Conn, outbound chan<- common.Command, now func() time.Time) (*Client, error) {

	client := &Client{
		conn:     conn,
		outbound: outbound,
		now:      now,
	}
	return client, nil
}

// Close terminates a connection to core.
func (c *Client) logout() error {

	c.outbound <- common.Command{
		ID:     common.LOGOUT,
		Sender: c.imei,
	}
	return nil
}

func (c *Client) receiveLoginMessage() error {
	log.Println("DEBUG: receiveLoginMessage start")
	var loginMsg [15]byte
	n, err := c.conn.Read(loginMsg[:])
	if err != nil || n < 15 {
		return fmt.Errorf("ERR trying to read IMEI, bytes read: %d, err: %v", n, err)
	}

	imei, err := decodeIMEI(loginMsg[:])
	if err != nil {
		return fmt.Errorf("ERR decoding IMEI bytes %v ", err)
	}
	c.imei = imei
	c.inbound = make(chan common.Command)

	c.outbound <- common.Command{
		ID:              common.LOGIN,
		Sender:          c.imei,
		CallbackChannel: c.inbound,
	}

	cmd := <-c.inbound
	switch cmd.ID {
	case common.WELCOME:
		log.Printf("Server accepted client connection")
	case common.KILL:
		return fmt.Errorf("Server sent KILL cmd to connected device %d", c.imei)
	}

	log.Println("DEBUG: receiveLoginMessage END")
	return nil
}

func (c *Client) receiveReadingsLoop() {
	var payload [40]byte
	log.Print("DEBUG starting receiveReadingsLoop")
	for {

		select {
		case cmd := <-c.inbound:
			if cmd.ID == common.KILL {
				log.Printf("Server sent KILL cmd to connected device %d", c.imei)
				break
			}
		default:
			//Continue receiveReadings loop
		}
		err := c.nextReading(payload[:])
		if err != nil {
			log.Printf("ERR during reading %v", err)
			c.logout()
			break
		}

		c.outbound <- common.Command{
			ID:     common.READING,
			Sender: c.imei,
			Body:   payload[:], //TODO #20 pass a copy instead of a reference
		}

	}
	log.Println("DEBUG receiveReadingsLoop exit")

}

func (c *Client) nextReading(payload []byte) error {
	err := c.conn.SetReadDeadline(c.now().Add(time.Second * 2))
	if err != nil {
		return err

	}
	n, err := c.conn.Read(payload[:])
	if err != nil {
		return err

	}
	//TODO is possible to read less bytes , we should use manual buffer or an buf reader
	if n != 40 {
		err := fmt.Errorf("WARN read only %d bytes, expected 40 for reading payloads", n)
		return err
	}

	return nil
}

func (c *Client) Read(wg *sync.WaitGroup) {
	log.Println("DEBUG starting client Read")
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Printf("ERR trying to close the connection %v", err)
		}
		log.Println("DEBUG client connection closed")
		wg.Done()
	}()

	if c.imei == 0 {
		err := c.receiveLoginMessage()
		if err != nil {
			log.Printf("%v", err)
			return
		}

	}
	c.receiveReadingsLoop()
	log.Print("DEBUG client Read exit")

}
