package main

import (
	"errors"
	"flag"
	"log"
	"time"

	"github.com/0xPolygon/polygon-edge/helper/common"
	"github.com/0xPolygon/polygon-edge/server"
	golog "github.com/ipfs/go-log/v2"
	"github.com/maticnetwork/avail-settlement/consensus/avail"
	"github.com/maticnetwork/avail-settlement/pkg/config"
)

const (
	AvailConsensus server.ConsensusType = "avail"
)

func main() {
	var availAddr, path string
	flag.StringVar(&availAddr, "avail-addr", "ws://127.0.0.1:9944/v1/json-rpc", "Avail JSON-RPC URL")
	flag.StringVar(&path, "config-file", "./configs/bootnode.yaml", "Path to the configuration file")

	flag.Parse()

	// Enable LibP2P logging
	golog.SetAllLoggers(golog.LevelError)

	config, err := config.NewServerConfig(path)
	if err != nil {
		log.Fatalf("failure to get node configuration: %s", err)
	}

	log.Printf("Server config: %+v", config)

	// Attach the concensus to the server
	err = server.RegisterConsensus(AvailConsensus, avail.Factory(avail.Config{AvailAddr: availAddr}))
	if err != nil {
		log.Fatalf("failure to register consensus: %s", err)
	}

	serverInstance, err := server.NewServer(config)
	if err != nil {
		log.Fatalf("failure to start node: %s", err)
	}

	log.Printf("Server instance %#v", serverInstance)

	if err := HandleSignals(serverInstance.Close); err != nil {
		log.Fatalf("handle signal error: %s", err)
	}
}

// HandleSignals is a helper method for handling signals sent to the console
// Like stop, error, etc.
func HandleSignals(closeFn func()) error {
	signalCh := common.GetTerminationSignalCh()
	sig := <-signalCh

	log.Printf("\n[SIGNAL] Caught signal: %v\n", sig)

	// Call the Minimal server close callback
	gracefulCh := make(chan struct{})

	go func() {
		if closeFn != nil {
			closeFn()
		}

		close(gracefulCh)
	}()

	select {
	case <-signalCh:
		return errors.New("shutdown by signal channel")
	case <-time.After(5 * time.Second):
		return errors.New("shutdown by timeout")
	case <-gracefulCh:
		return nil
	}
}
