package client

import (
	"log"
	"net"
	"time"

	"github.com/spin-org/thermomatic/internal/common"
)

// Randomatic implements a simple TCP client that sends `n` random readings after login
func Randomatic(clientServerAddress *string, clientImei *string) {
	baseClient(clientServerAddress, clientImei, 25*time.Millisecond, time.Nanosecond)
}

// Slowmatic implements a client that will be send be disconected by the server  because it takes more than 2 seconds between msgs
func Slowmatic(clientServerAddress *string, clientImei *string) {
	baseClient(clientServerAddress, clientImei, 3*time.Second, time.Nanosecond)
}

// TooSlowToPlayWithGrownups implements a client that is too slow to send the initial login message, so the server will disconnect the connection
func TooSlowToPlayWithGrownups(clientServerAddress *string, clientImei *string) {
	baseClient(clientServerAddress, clientImei, time.Second, 2*time.Second)
}

func baseClient(clientServerAddress *string, clientImei *string, readingRate time.Duration, sleepBeforeLogin time.Duration) {
	//TODO  detect and log desconections
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
	time.Sleep(sleepBeforeLogin)
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
		time.Sleep(readingRate)
	}
	//TODO add a flag to reading frequency , default 25ms
	//TODO nice to have read from std in csv file of readings
}
