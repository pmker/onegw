package core

import (
	"context"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/pmkoo"
	"github.com/pmker/onegw/pmkoo/deposit"
	"github.com/pmker/onegw/pmkoo/models"
	"github.com/pmker/onegw/pmkoo/withdraw"
	"os"
	"path/filepath"
	"time"

	"github.com/albrow/stringset"
	ethrpc "github.com/ethereum/go-ethereum/rpc"
	p2pcrypto "github.com/libp2p/go-libp2p-crypto"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	ma "github.com/multiformats/go-multiaddr"
	"github.com/nats-io/nats.go"
	"github.com/pmker/onegw/ethereum"
	"github.com/pmker/onegw/keys"
	"github.com/pmker/onegw/oneplus/db"
	"github.com/pmker/onegw/oneplus/meshdb"
	bw "github.com/pmker/onegw/oneplus/watch"
	"github.com/pmker/onegw/p2p"
	log "github.com/sirupsen/logrus"
)

const (
	blockWatcherRetentionLimit = 20
	ethereumRPCRequestTimeout  = 30 * time.Second
	ethWatcherPollingInterval  = 1 * time.Minute
	peerConnectTimeout         = 60 * time.Second
	checkNewAddrInterval       = 20 * time.Second
	expirationPollingInterval  = 50 * time.Millisecond
)

// Config is a set of configuration options for 0x Mesh.
type Config struct {
	// Verbosity is the logging verbosity: 0=panic, 1=fatal, 2=error, 3=warn, 4=info, 5=debug 6=trace
	Verbosity int `envvar:"VERBOSITY" default:"4"`
	// DataDir is the directory to use for persisting all data, including the
	// database and private key files.
	DataDir string `envvar:"DATA_DIR" default:"0vedb"`
	// P2PListenPort is the port on which to listen for new peer connections. By
	// default, 0x Mesh will let the OS select a randomly available port.
	P2PListenPort int `envvar:"P2P_LISTEN_PORT" default:"0"`
	// EthereumRPCURL is the URL of an Etheruem node which supports the JSON RPC
	// API.
	EthereumRPCURL string `envvar:"ETHEREUM_RPC_URL" default:"main"`
	// EthereumNetworkID is the network ID to use when communicating with
	// Ethereum.
	EthereumNetworkID int `envvar:"ETHEREUM_NETWORK_ID" default:"1"`
	// UseBootstrapList is whether to use the predetermined list of peers to
	// bootstrap the DHT and peer discovery.
	UseBootstrapList bool `envvar:"USE_BOOTSTRAP_LIST" default:"false"`
	// OrderExpirationBuffer is the amount of time before the order's stipulated expiration time
	// that you'd want it pruned from the Mesh node.
	OrderExpirationBuffer time.Duration `envvar:"ORDER_EXPIRATION_BUFFER" default:"10s"`
	// BlockPollingInterval is the polling interval to wait before checking for a new Ethereum gateway
	// that might contain transactions that impact the fillability of orders stored by Mesh. Different
	// networks have different gateway producing intervals: POW networks are typically slower (e.g., Mainnet)
	// and POA networks faster (e.g., Kovan) so one should adjust the polling interval accordingly.
	BlockPollingInterval time.Duration `envvar:"BLOCK_POLLING_INTERVAL" default:"5s"`
	// EthereumRPCMaxContentLength is the maximum request Content-Length accepted by the backing Ethereum RPC
	// endpoint used by Mesh. Geth & Infura both limit a request's content length to 1024 * 512 Bytes. Parity
	// and Alchemy have much higher limits. When batch validating 0x orders, we will fit as many orders into a
	// request without crossing the max content length. The default value is appropriate for operators using Geth
	// or Infura. If using Alchemy or Parity, feel free to double the default max in order to reduce the
	// number of RPC calls made by Mesh.
	EthereumRPCMaxContentLength int `envvar:"ETHEREUM_RPC_MAX_CONTENT_LENGTH" default:"524288"`

	NatsConnectUrl string
	Action         string `envvar:"ActionDB" default:"main"`
	ActionDB       string `envvar:"ActionDB" default:"main"`
}

type snapshotInfo struct {
	Snapshot            *db.Snapshot
	ExpirationTimestamp time.Time
}

type App struct {
	config         Config
	db             *meshdb.MeshDB
	node           *p2p.Node
	networkID      int
	blockWatcher   *bw.Watcher
	gatewayWatcher *deposit.Watcher
	ethWatcher     *ethereum.ETHWatcher
	nat            *nats.Conn
	ec             *nats.EncodedConn
	nats           *pmkoo.Vgw
	deposit        *deposit.Action
	withdraw       *withdraw.Watcher
	action         string
}

const maxOrderSizeInBytes = 8192

// maxOrderExpirationDuration is the maximum duration between the current time and the expiration
// set on an order that will be accepted by Mesh.
const maxOrderExpirationDuration = 9 * 30 * 24 * time.Hour // 9 months

func New(config Config) (*App, error) {
	// Configure logger
	// TODO(albrow): Don't use global variables for log settings.
	log.SetLevel(log.Level(config.Verbosity))
	log.WithFields(map[string]interface{}{
		"config":  config,
		"version": "development",
	}).Info("Initializing new core.App")

	if config.EthereumRPCMaxContentLength < maxOrderSizeInBytes {
		return nil, fmt.Errorf("Cannot set `EthereumRPCMaxContentLength` to be less then maxOrderSizeInBytes: %d", maxOrderSizeInBytes)
	}

	bowPath := filepath.Join(config.DataDir, config.ActionDB)

	opt := bgdb.DefaultOptions
	opt.Dir = bowPath
	opt.ValueDir = opt.Dir
	//opt.Logger = emptyLogger{}
	bowdb, err := bgdb.Open(opt)
	if err != nil {
		log.WithField("error", err).Error("bowdb error open")

	}
	models.Connect(os.Getenv("ONE_DATABASE_URL"))

	databasePath := filepath.Join(config.DataDir, "db")
	meshDB, err := meshdb.NewMeshDB(databasePath)
	if err != nil {
		return nil, err
	}
	nc, err := nats.Connect(config.NatsConnectUrl)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"nats1": config.NatsConnectUrl,
			"err":   err.Error(),
		}).Error("started p2p node")
	}
	//defer nc.Close()

	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"nats2": config.NatsConnectUrl,
			"err":   err.Error(),
		}).Error("started p2p node")
	}
	nats, err := pmkoo.Nats(bowdb, ec)
	if err != nil {
		return nil, err
	}
	app := &App{}
	if config.Action == "DEPOSIT" {
		// Initialize gateway watcher (but don't start it yet).
		blockWatcherClient, err := bw.NewRpcClient(config.EthereumRPCURL, ethereumRPCRequestTimeout)
		if err != nil {
			return nil, err
		}

		//topics := orderwatch.GetRelevantTopics()
		blockWatcherConfig := bw.Config{
			MeshDB:              meshDB,
			PollingInterval:     config.BlockPollingInterval,
			StartBlockDepth:     ethrpc.LatestBlockNumber,
			BlockRetentionLimit: blockWatcherRetentionLimit,
			WithLogs:            true,
			Client:              blockWatcherClient,
			RpcURL:              config.EthereumRPCURL,
		}
		blockWatcher := bw.New(blockWatcherConfig)
		go func() {
			for {
				err, isOpen := <-blockWatcher.Errors
				if isOpen {
					log.WithField("error", err).Error("BlockWatcher error encountered")
				} else {
					return // Exit when the error channel is closed
				}
			}
		}()
		gatewayWatcher, err := deposit.Deposit(blockWatcher, bowdb, config.EthereumNetworkID, ec)
		if err != nil {
			return nil, err
		}

		deposit, err := deposit.New(bowdb, ec, blockWatcher)

		if err != nil {
			return nil, err
		}

		app = &App{
			config:         config,
			db:             meshDB,
			networkID:      config.EthereumNetworkID,
			blockWatcher:   blockWatcher,
			gatewayWatcher: gatewayWatcher,

			nat:     nc,
			ec:      ec,
			nats:    nats,
			deposit: deposit,
		}
	}

	if config.Action == "WITHDRAW" {
		ethClient, err := ethclient.Dial(config.EthereumRPCURL)
		if err != nil {
			return nil, err
		}

		ethWatcher, err := ethereum.NewETHWatcher(ethWatcherPollingInterval, ethClient, config.EthereumNetworkID)
		if err != nil {
			return nil, err
		}

		withdraw, err := withdraw.New(bowdb, ec)

		if err != nil {
			return nil, err
		}

		// Initialize the ETH balance watcher (but don't start it yet).

		app = &App{
			config:     config,
			db:         meshDB,
			networkID:  config.EthereumNetworkID,
			ethWatcher: ethWatcher,
			nat:        nc,
			ec:         ec,
			nats:       nats,
			withdraw:   withdraw,
		}
	}
	app.action = config.Action
	//defer ec.Close()

	// Initialize order watcher (but don't start it yet).

	// Initialize the p2p node.
	//privateKeyPath := filepath.Join(config.DataDir, "keys", "privkey")
	//privKey, err := initPrivateKey(privateKeyPath)
	//if err != nil {
	//	return nil, err
	//}
	//nodeConfig := p2p.Config{
	//	Topic:            getPubSubTopic(config.EthereumNetworkID),
	//	ListenPort:       config.P2PListenPort,
	//	Insecure:         false,
	//	PrivateKey:       privKey,
	//	MessageHandler:   app,
	//	RendezvousString: getRendezvous(config.EthereumNetworkID),
	//	UseBootstrapList: config.UseBootstrapList,
	//}
	//node, err := p2p.New(nodeConfig)
	//if err != nil {
	//	return nil, err
	//}
	//app.node = node

	return app, nil
}

func getPubSubTopic(networkID int) string {
	return fmt.Sprintf("/vgw-orders/network/%d/version/0.0.1", networkID)
}

func getRendezvous(networkID int) string {
	return fmt.Sprintf("/vgw-mesh/network/%d/version/0.0.1", networkID)
}

func initPrivateKey(path string) (p2pcrypto.PrivKey, error) {
	privKey, err := keys.GetPrivateKeyFromPath(path)
	if err == nil {
		return privKey, nil
	} else if os.IsNotExist(err) {
		// If the private key doesn't exist, generate one.
		log.Info("No private key found. Generating a new one.")
		return keys.GenerateAndSavePrivateKey(path)
	}

	// For any other type of error, return it.
	return nil, err
}

func (app *App) Start() error {
	//go func() {
	//	err := app.node.Start()
	//	if err != nil {
	//		log.WithField("error", err.Error()).Error("p2p node returned error")
	//		app.Close()
	//	}
	//}()
	//addrs := app.node.Multiaddrs()
	//go app.periodicallyCheckForNewAddrs(addrs)
	//log.WithFields(map[string]interface{}{
	//	"addresses": addrs,
	//	"peerID":    app.node.ID().String(),
	//}).Info("started p2p node")

	if app.action == "DEPOSIT" {
		if err := app.gatewayWatcher.Start(); err != nil {
			return err
		}

		if err := app.deposit.Start();
		err != nil{
			return err
		}

		log.Info("started order watcher")
		if err := app.blockWatcher.Start(); err != nil {
			return err
		}
	}
	if app.action == "WITHDRAW" {
		if err := app.withdraw.Start(); err != nil {
			return err
		}
		if err := app.ethWatcher.Start(); err != nil {
			return err
		}
	}
	if err := app.nats.Start(); err != nil {
		return err
	}

	return nil
}

func (app *App) periodicallyCheckForNewAddrs(startingAddrs []ma.Multiaddr) {
	seenAddrs := stringset.New()
	for _, addr := range startingAddrs {
		seenAddrs.Add(addr.String())
	}
	// TODO: There might be a more efficient way to do this if we have access to
	// an event bus. See: https://github.com/libp2p/go-libp2p/issues/467
	for {
		time.Sleep(checkNewAddrInterval)
		newAddrs := app.node.Multiaddrs()
		for _, addr := range newAddrs {
			if !seenAddrs.Contains(addr.String()) {
				log.WithFields(map[string]interface{}{
					"address": addr,
				}).Info("found new listen address")
				seenAddrs.Add(addr.String())
			}
		}
	}
}

// ErrSnapshotNotFound is the error returned when a snapshot not found with a particular id
type ErrSnapshotNotFound struct {
	id string
}

func (e ErrSnapshotNotFound) Error() string {
	return fmt.Sprintf("No snapshot found with id: %s. To create a new snapshot, send a request with an empty snapshotID", e.id)
}

// AddPeer can be used to manually connect to a new peer.
func (app *App) AddPeer(peerInfo peerstore.PeerInfo) error {
	ctx, cancel := context.WithTimeout(context.Background(), peerConnectTimeout)
	defer cancel()
	return app.node.Connect(ctx, peerInfo)
}

// Close closes the app
func (app *App) Close() {
	if err := app.node.Close(); err != nil {
		log.WithField("error", err.Error()).Error("error while closing node")
	}
	app.ethWatcher.Stop()

	app.blockWatcher.Stop()
	app.db.Close()
	app.nat.Close()
	app.ec.Close()
}
