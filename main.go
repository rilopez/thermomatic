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

func initCommandLineInterface(handleServerCmd func(uint), handleClientCmd func(clientServerAddress *string, clientImei *string, clientType *string)) {
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCmd.Uint("port", defaultPort, "port")

	clientCmd := flag.NewFlagSet("client", flag.ExitOnError)
	clientServerAddress := clientCmd.String("server-address", "", "server-address should have this format  host:port")
	clientImei := clientCmd.String("imei", "", "imei")
	clientType := clientCmd.String("type", "random", "type")

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
		handleClientCmd(clientServerAddress, clientImei, clientType)

	default:
		flag.PrintDefaults()
		os.Exit(1)
	}
}

func serverCommandHandler(port uint) {
	_ = initLog("server.log")
	server.Start(port)
}

func clientCommandHandler(clientServerAddress *string, clientImei *string, clientType *string) {
	_ = initLog("client.log")
	switch *clientType {
	case "random":
		client.Randomatic(clientServerAddress, clientImei)
	case "slow":
		client.Slowmatic(clientServerAddress, clientImei)
	case "too-slow":
		client.TooSlowToPlayWithGrownups(clientServerAddress, clientImei)
	default:
		panic(fmt.Sprintf("unknown clientType %s", *clientType))
	}

}
