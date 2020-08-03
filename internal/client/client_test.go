package client

import (
	"net"
	"testing"
	"time"

	"github.com/spin-org/thermomatic/internal/common"
)

func TestString(t *testing.T) {

	c := &Client{
		IMEI:             490154203237518,
		LastReadingEpoch: 1257894000000000000,
		LastReading: &Reading{
			Temperature:  67.77,
			Altitude:     2.63555,
			Latitude:     33.41,
			Longitude:    44.4,
			BatteryLevel: 0.25666,
		},
	}

	expectedRecord := "1257894000000000000,490154203237518,67.770000,2.635550,33.410000,44.400000,0.256660"
	actualRecord := c.String()
	//TODO follow up email to Rob about field string format requirements
	if expectedRecord != actualRecord {
		t.Errorf("expected %s got %s", expectedRecord, actualRecord)
	}
}

func TestReceiveReadings(t *testing.T) {
	t.Error(common.ErrNotImplemented)
}

func TestReceiveLoginMessage(t *testing.T) {
	timeout := time.After(1 * time.Second)
	logins := make(chan *Client)
	device := &Client{
		register: logins,
	}

	expectedIMEI := uint64(490154203237518)
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Errorf("ERR while trying to start testing server, %v", err)
	}
	go func() {
		defer ln.Close()
		device.Conn, err = ln.Accept()
		err = device.receiveLoginMessage()
		if err != nil {
			t.Errorf("ERR while receiving Login Message, %v", err)
		}

		if device.IMEI != expectedIMEI {
			t.Errorf("Expected client.IMEI %d but got %d", expectedIMEI, device.IMEI)
		}
	}()

	conn, err := net.Dial("tcp", ln.Addr().String())
	expectedIMEIbytes := []byte{4, 9, 0, 1, 5, 4, 2, 0, 3, 2, 3, 7, 5, 1, 8}

	conn.Write(expectedIMEIbytes[:])
	select {
	case <-timeout:
		t.Fatal("Timeout")
	case clientToLogin := <-logins:
		if clientToLogin != device {
			t.Errorf("device was not sent to login channel")
		}
	}

}

func TestRead(t *testing.T) {
	t.Error(common.ErrNotImplemented)
}
