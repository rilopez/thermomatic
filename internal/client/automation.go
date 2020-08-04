package client

import (
	"log"
	"net"
	"time"

	"github.com/spin-org/thermomatic/internal/common"
)

// Randomatic Implements a simple TCP client that sends `n` random readings after login
func Randomatic(clientServerAddress *string, clientImei *string) {
	log.Printf("Connecting to %s", *clientServerAddress)
	conn, err := net.Dial("tcp", *clientServerAddress)
	defer func() {
		conn.Close()
	}()

	if err != nil {
		log.Fatal(err)
	}
	log.Printf("DEBUG: converting imei %s to bytes", *clientImei)
	imeiBytes, err := common.ImeiStringToBytes(clientImei)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("DEBUG: sending login imei %v to server", imeiBytes)
	n, err := conn.Write(imeiBytes[:])
	log.Printf("DEBUG: %d bytes sent", n)
	if err != nil {
		log.Fatalf("Error trying to send IMEI %v", err)
	}

	//TODO add a flag for amount of random reading messages
	totalReadings := 5
	for i := 0; i < totalReadings; i++ {
		randomReading := CreateRandReading()
		log.Printf("DEBUG: [%d] sending reading %v to server", i, randomReading)
		n, err := conn.Write(randomReading[:])
		log.Printf("DEBUG: [%d] %d bytes sent", i, n)
		if err != nil {
			log.Fatalf("Error trying to send reading %v", err)
		}
		time.Sleep(25 * time.Millisecond)
	}
	//TODO add a flag to reading frequency , default 25ms
	//TODO nice to have read from std in csv file of readings
}
