package core

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/pmker/onegw/oneplus/common/bgdb"
	"github.com/pmker/onegw/oneplus/openssl"
	"github.com/pmker/onegw/pmkoo"
	"github.com/pmker/onegw/pmkoo/deposit"
	"github.com/pmker/onegw/pmkoo/models"
	"github.com/pmker/onegw/pmkoo/types"
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

type snapshotInfo struct {
	Snapshot            *db.Snapshot
	ExpirationTimestamp time.Time
}

type App struct {
	config         types.Config
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

func LoadConfigFromDb(bowdb *bgdb.Store) (*types.Config, error) {
	config := &types.Config{}

	if exists := bowdb.Cache().Has(types.OPEN_SYS_CONFIG); exists {
		value, err := bowdb.Cache().Get(types.OPEN_SYS_CONFIG)

		data, err := base64.StdEncoding.DecodeString(value)
		if err != nil {

			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warn("json.Unmarshal")
		}
		key, err := bowdb.Cache().Get(types.OPEN_SYS_CONFIG_KEY)

		dst, err := openssl.AesECBDecrypt(data, []byte(key), openssl.PKCS7_PADDING)
		///fmt.Print(string(dst)) // 123456
		if err != nil {
			fmt.Printf(err.Error())

		}

		err = json.Unmarshal(dst, config)
		if err != nil {
			log.WithFields(log.Fields{
				"error": err.Error(),
			}).Warn("json.Unmarshal")
		}
		//fmt.Printf("adfadf config%+v", config)

	}
	return config, nil

}

func New(config types.Config) (*App, error) {
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

	bowPath := filepath.Join(config.DataDir, "bgdb")

	opt := bgdb.DefaultOptions
	opt.Dir = bowPath
	opt.ValueDir = opt.Dir
	//opt.Logger = emptyLogger{}
	bowdb, err := bgdb.Open(opt)

	if err != nil {
		log.WithField("error", err).Error("bowdb error open")

	}
	pconfig,_:=LoadConfigFromDb(bowdb)





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

	 fmt.Printf("fuck config %+v \n",pconfig)

	nats, err := pmkoo.Nats(bowdb, ec,pconfig)
	if err != nil {
		return nil, err
	}

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

	wallet,err:=pmkoo.UnlockViteWallet(pconfig)
	if err != nil {
		return nil, err
	}
	deposit, err := deposit.New(bowdb, ec, blockWatcher,wallet,pconfig)

	if err != nil {
		return nil, err
	}

	ethClient, err := ethclient.Dial(config.EthereumRPCURL)
	if err != nil {
		return nil, err
	}

	ethWatcher, err := ethereum.NewETHWatcher(ethWatcherPollingInterval, ethClient, config.EthereumNetworkID)
	if err != nil {
		return nil, err
	}

	withdraw, err := withdraw.New(bowdb, ec,wallet,pconfig)

	if err != nil {
		return nil, err
	}

	// Initialize the ETH balance watcher (but don't start it yet).

	app := &App{
		config:         config,
		db:             meshDB,
		networkID:      config.EthereumNetworkID,
		ethWatcher:     ethWatcher,
		nat:            nc,
		ec:             ec,
		nats:           nats,
		withdraw:       withdraw,
		blockWatcher:   blockWatcher,
		gatewayWatcher: gatewayWatcher,
		deposit:        deposit,
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

	if err := app.blockWatcher.Start(); err != nil {
		return err
	}

	if err := app.gatewayWatcher.Start(); err != nil {
		return err
	}

	if err := app.deposit.Start();
		err != nil {
		return err
	}

	if err := app.withdraw.Start(); err != nil {
		return err
	}
	if err := app.ethWatcher.Start(); err != nil {
		return err
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
