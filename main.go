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

type serverHandler func(uint, uint)
type clientHandler func(clientServerAddress *string, clientImei *string, clientType *string, numReadings *uint)

func initCommandLineInterface(handleServerCmd serverHandler, handleClientCmd clientHandler) {
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCmd.Uint("port", 1337, "port")
	serverHTTPPort := serverCmd.Uint("httport", 80, "port")

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientServerAddress := clientCmd.String("server-address", "localhost:1337", "server-address should have this format  host:port")
	clientImei := clientCmd.String("imei", "", "imei")
	clientType := clientCmd.String("type", "random", "type")
	clientNumReadings := clientCmd.Uint("readings", 5, "readings,  if equals 0 creates an infite readings loop")

	if len(os.Args) < 2 {
		fmt.Println("server or client subcommand is required")
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverCmd.Parse(os.Args[2:])
		handleServerCmd(*serverPort, *serverHTTPPort)
	case "client":
		clientCmd.Parse(os.Args[2:])
		if *clientServerAddress == "" {
			panic("Please use -server-address=host:port to connect ")
		}

		if *clientImei == "" {
			panic("-imei is required")
		}
		handleClientCmd(clientServerAddress, clientImei, clientType, clientNumReadings)

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func serverCommandHandler(port uint, httpPort uint) {
	_ = initLog("server.log")
	server.Start(port, httpPort)
}

func clientCommandHandler(clientServerAddress *string, clientImei *string, clientType *string, numReadings *uint) {
	_ = initLog("client.log")
	switch *clientType {
	case "random":
		client.Randomatic(clientServerAddress, clientImei, numReadings)
	case "slow":
		client.Slowmatic(clientServerAddress, clientImei, numReadings)
	case "too-slow":
		client.TooSlowToPlayWithGrownups(clientServerAddress, clientImei, numReadings)
	default:
		panic(fmt.Sprintf("unknown clientType %s", *clientType))
	}

}
