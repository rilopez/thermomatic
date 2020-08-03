package main

import (
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"time"

	"github.com/spin-org/thermomatic/internal/client"
	"github.com/spin-org/thermomatic/internal/server"

	"github.com/spin-org/thermomatic/internal/common"
)

func main() {
	initCommandLineInterface(
		serverCommandHandler,
		clientCommandHandler,
	)
}

const (
	defaultPort = 1337
)

func initLog(fileName string) error {
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	return err
}

func initCommandLineInterface(handleServerCmd func(uint), handleClientCmd func(clientServerAddress *string, clientImei *string)) {
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCmd.Uint("port", defaultPort, "port")

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientServerAddress := clientCmd.String("server-address", "", "server-address should have this format  host:port")
	clientImei := clientCmd.String("imei", "", "imei")

	if len(os.Args) < 2 {
		fmt.Println("server or client subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverCmd.Parse(os.Args[2:])
		handleServerCmd(*serverPort)
	case "client":
		clientCmd.Parse(os.Args[2:])
		if *clientServerAddress == "" {
			panic("Please use -server-address=host:port to connect ")
		}

		if *clientImei == "" {
			panic("-imei is required")
		}
		handleClientCmd(clientServerAddress, clientImei)

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func serverCommandHandler(port uint) {
	_ = initLog("server.log")
	address := fmt.Sprintf(":%d", port)
	ln, err := net.Listen("tcp", address)
	log.Printf("Server started, using %s as address", address)
	if err != nil {
		log.Fatalf("%v", err)
	}

	core := server.NewCore()

	go core.Run()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("%v", err)
		}
		log.Printf("client connection from %v", conn.RemoteAddr())
		//if the device fail to send the login message within 1 second the server will drop the client connection.
		conn.SetReadDeadline(time.Now().Add(time.Second))

		c := client.NewClient(
			conn,
			core.Commands,
			core.Registrations,
			core.Deregistrations,
		)
		go c.Read()
	}

}

func clientCommandHandler(clientServerAddress *string, clientImei *string) {
	_ = initLog("client.log")
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
		randomReading := client.CreateRandReading()
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
