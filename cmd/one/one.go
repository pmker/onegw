package one

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/pmker/onegw/pmkoo"
	"github.com/pmker/onegw/pmkoo/core"
	"github.com/pmker/onegw/pmkoo/types"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"
)

var defaultNodeConfigFileName = "config.json"


// standaloneConfig contains configuration options specific to running 0x Mesh
// in standalone mode (i.e. not in a browser).
func DataDir() string {
	home := pmkoo.HomeDir()
	if home != "" {
		onegwData:=filepath.Join(home, "onegw", "data")
		if _, err := os.Stat( onegwData); os.IsNotExist(err) {
			if err := os.MkdirAll(onegwData, 0755); err != nil {
				fmt.Printf(err.Error())
			}
		}
		//fmt.Printf(onegwData)
		return onegwData
	}
	return ""
}

func NewOneApp() error{
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
		return err
	}
	// Configure logger to output JSON
	// TODO(albrow): Don't use global settings for logger.
	//log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	log.SetFormatter(&log.TextFormatter{ForceColors: true})


	// Parse env vars
	var coreConfig types.Config
	//if err := envvar.Parse(&coreConfig); err != nil {
	//	log.WithField("error", err.Error()).Fatal("could not parse environment variables")
	//}
	//
	if jsonConf, err := ioutil.ReadFile(defaultNodeConfigFileName); err == nil {
		//json.NewDecoder(bytes.NewBuffer([]byte(in))).Decode(&msg)
		//err = json.Unmarshal(jsonConf, &coreConfig)
		//if err != nil {
		//
		//}

		if err := json.NewDecoder(bytes.NewBuffer([]byte(jsonConf))).Decode(&coreConfig); err != nil {
			log.Error("Cannot unmarshal the default config file content", "error", err)

		}

	}

	coreConfig.EthereumRPCURL = os.Getenv("ETHEREUM_RPC_URL")
	coreConfig.NatsConnectUrl = os.Getenv("NATS_CONNECT_URL")

	networkdId, err := strconv.ParseInt(os.Getenv("ETHEREUM_NETWORK_ID"), 10, 32)

	if err != nil {
		fmt.Print("err networkid")
		return err
	}
	coreConfig.EthereumNetworkID = int(networkdId)
	coreConfig.DataDir=DataDir()
	//fmt.Printf("config:%+v",coreConfig)

	//var config rpc.StandaloneConfig
	//port := os.Getenv("RPC_PORT")

	//p, err := strconv.ParseInt(port, 10, 32)
	//if err != nil {
	//	fmt.Print("rpc networkid")
	//
	//}
	//config.RPCPort = int(p)
	//
	//fmt.Print(config.RPCPort)
	//if err := envvar.Parse(&config); err != nil {
	//	log.WithField("error", err.Error()).Fatal("could not parse environment variables")
	//}

	// Start core.App.
	app, err := core.New(coreConfig)
	if err != nil {
		log.WithField("error", err.Error()).Fatal("could not initialize app")
		return err
	}
	if err := app.Start(); err != nil {
		log.WithField("error", err.Error()).Fatal("fatal error while starting app")
		return err
	}
	defer app.Close()


	// Block forever or until the app is closed.
	select {}
}
