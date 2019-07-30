package cmd

import (
	"fmt"
	"github.com/joho/godotenv"
	"os"
	"strconv"

	"github.com/pmker/onegw/core"
	"github.com/pmker/onegw/cmd/rpc"
	"github.com/plaid/go-envvar/envvar"
	log "github.com/sirupsen/logrus"
)

// standaloneConfig contains configuration options specific to running 0x Mesh
// in standalone mode (i.e. not in a browser).

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	// Configure logger to output JSON
	// TODO(albrow): Don't use global settings for logger.
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.SetFormatter(&log.TextFormatter{ForceColors: true})


	// Parse env vars
	var coreConfig core.Config
	if err := envvar.Parse(&coreConfig); err != nil {
		log.WithField("error", err.Error()).Fatal("could not parse environment variables")
	}
	coreConfig.EthereumRPCURL = os.Getenv("ETHEREUM_RPC_URL")
	coreConfig.NatsConnectUrl = os.Getenv("NATS_CONNECT_URL")

	coreConfig.Action = "WITHDRAW"
	coreConfig.ActionDB = os.Getenv("WITHDRAW_DATA_DIR")

	networkdId, err := strconv.ParseInt(os.Getenv("ETHEREUM_NETWORK_ID"), 10, 32)

	if err != nil {
		fmt.Print("err networkid")

	}
	coreConfig.EthereumNetworkID = int(networkdId)

	var config rpc.StandaloneConfig
	port := os.Getenv("RPC_PORT")

	p, err := strconv.ParseInt(port, 10, 32)
	if err != nil {
		fmt.Print("rpc networkid")

	}
	config.RPCPort = int(p)

	if err := envvar.Parse(&config); err != nil {
		log.WithField("error", err.Error()).Fatal("could not parse environment variables")
	}

	// Start core.App.
	app, err := core.New(coreConfig)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("could not initialize app")
	}
	if err := app.Start(); err != nil {
		log.WithField("error", err.Error()).Fatal("fatal error while starting app")
	}
	defer app.Close()



	// Block forever or until the app is closed.
	select {}
}
