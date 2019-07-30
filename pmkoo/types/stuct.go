package types

import "time"
const(
	OPEN_SYS_CONFIG="OPEN_SYS_CONFIG"
	OPEN_SYS_CONFIG_KEY="OPEN_SYS_CONFIG_KEY"
)
// Config is a set of configuration options for 0x Mesh.
type Config struct {
	// Verbosity is the logging verbosity: 0=panic, 1=fatal, 2=error, 3=warn, 4=info, 5=debug 6=trace
	Verbosity int
	// DataDir is the directory to use for persisting all data, including the
	// database and private key files.
	DataDir string
	// P2PListenPort is the port on which to listen for new peer connections. By
	// default, 0x Mesh will let the OS select a randomly available port.
	P2PListenPort int
	// EthereumRPCURL is the URL of an Etheruem node which supports the JSON RPC
	// API.
	EthereumRPCURL string
	// EthereumNetworkID is the network ID to use when communicating with
	// Ethereum.
	EthereumNetworkID int
	// UseBootstrapList is whether to use the predetermined list of peers to
	// bootstrap the DHT and peer discovery.
	UseBootstrapList bool
	// OrderExpirationBuffer is the amount of time before the order's stipulated expiration time
	// that you'd want it pruned from the Mesh node.
	// BlockPollingInterval is the polling interval to wait before checking for a new Ethereum gateway
	// that might contain transactions that impact the fillability of orders stored by Mesh. Different
	// networks have different gateway producing intervals: POW networks are typically slower (e.g., Mainnet)
	// and POA networks faster (e.g., Kovan) so one should adjust the polling interval accordingly.
	BlockPollingInterval time.Duration `json:"BlockPollingInterval,timeunit=s"`
	// EthereumRPCMaxContentLength is the maximum request Content-Length accepted by the backing Ethereum RPC
	// endpoint used by Mesh. Geth & Infura both limit a request's content length to 1024 * 512 Bytes. Parity
	// and Alchemy have much higher limits. When batch validating 0x orders, we will fit as many orders into a
	// request without crossing the max content length. The default value is appropriate for operators using Geth
	// or Infura. If using Alchemy or Parity, feel free to double the default max in order to reduce the
	// number of RPC calls made by Mesh.
	EthereumRPCMaxContentLength int
	NatsConnectUrl              string
	Action                      string
	ActionDB                    string
	OneKey                      string
	Mnemonic                    string
	Epath                       string
	Vpath                       string
}

