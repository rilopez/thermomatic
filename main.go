package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/spin-org/thermomatic/internal/client"
	"github.com/spin-org/thermomatic/internal/server"
)

func main() {
	initCommandLineInterface(
		serverCommandHandler,
		clientCommandHandler,
	)
}

func initLog(fileName string) error {
	logFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}
	log.SetOutput(logFile)
	return err
}

type serverHandler func(port uint, httpPort uint, serverMaxClients uint)
type clientHandler func(clientServerAddress *string, clientImei *string, clientType *string, numReadings *uint, readingRateInMilliSeconds *uint)

func initCommandLineInterface(handleServerCmd serverHandler, handleClientCmd clientHandler) {
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCmd.Uint("port", 1337, "port number to listen for TCP connections of clients implementing the  thermomatic protocol")
	serverHTTPPort := serverCmd.Uint("http-port", 80, "port number to listen for HTTP connections used mainly for healthchecks")
	serverMaxClients := serverCmd.Uint("max-clients", 1000, "maximun number of active client connections  using the thermomatic protocol")

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientServerAddress := clientCmd.String("server-address", "localhost:1337", "Address (host:port) of the Thermomatic server")
	clientImei := clientCmd.String("imei", "", "device IMEI number")
	clientType := clientCmd.String("type", "random", "Automated simulated client type, it could be random, slow, too slow ")
	clientNumReadings := clientCmd.Uint("readings", 5, "Number of automatic readings the automated client will send,  if equals 0  it sends an infite number of readings")
	clientReadingRate := clientCmd.Uint("reading-rate", 25, "Number of milliseconds between each reading ")

	if len(os.Args) < 2 {
		fmt.Println("server or client subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		err := serverCmd.Parse(os.Args[2:])
		if err != nil {

			serverCmd.Usage()
		}
		handleServerCmd(*serverPort, *serverHTTPPort, *serverMaxClients)
	case "client":
		clientCmd.Parse(os.Args[2:])
		if *clientServerAddress == "" {
			panic("Please use -server-address=host:port to connect ")
		}

		if *clientImei == "" {
			panic("-imei is required")
		}

		if *clientType == "" {
			panic("-type is required, it could be random, slow or too-slow")
		}
		handleClientCmd(clientServerAddress, clientImei, clientType, clientNumReadings, clientReadingRate)

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func serverCommandHandler(port uint, httpPort uint, serverMaxClients uint) {
	_ = initLog("server.log")
	server.Start(port, httpPort, serverMaxClients)
}

func clientCommandHandler(clientServerAddress *string, clientImei *string, clientType *string, numReadings *uint, readingRateInMilliSeconds *uint) {
	_ = initLog("client.log")
	switch *clientType {
	case "random":
		client.Randomatic(clientServerAddress, clientImei, numReadings, readingRateInMilliSeconds)
	case "slow":
		client.Slowmatic(clientServerAddress, clientImei, numReadings)
	case "too-slow":
		client.TooSlowToPlayWithGrownups(clientServerAddress, clientImei, numReadings)
	default:
		panic(fmt.Sprintf("unknown clientType %s", *clientType))
	}

}
