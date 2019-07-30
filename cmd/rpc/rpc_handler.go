// +build !js

package rpc

import (


	"fmt"

	"time"

	"github.com/pmker/onegw/constants"
	"github.com/pmker/onegw/core"
	"github.com/pmker/onegw/rpc"

	peerstore "github.com/libp2p/go-libp2p-peerstore"
	log "github.com/sirupsen/logrus"

)
type StandaloneConfig struct {
	// RPCPort is the port to use for the JSON RPC API over WebSockets. By
	// default, 0x Mesh will let the OS select a randomly available port.
	RPCPort int `envvar:"RPC_PORT" default:"60557"`
}
type rpcHandler struct {
	app *core.App
}
//
//func (handler *rpcHandler) AddOrders(signedOrdersRaw []*json.RawMessage) (*zeroex.ValidationResults, error) {
//	panic("implement me")
//}
//
//func (handler *rpcHandler) GetOrders(page, perPage int, snapshotID string) (*rpc.GetOrdersResponse, error) {
//	panic("implement me")
//}
//
//func (handler *rpcHandler) SubscribeToOrders(ctx context.Context) (*ethRpc.Subscription, error) {
//	panic("implement me")
//}

// listenRPC starts the RPC server and listens on config.RPCPort. It blocks
// until there is an error or the RPC server is closed.
func ListenRPC(app *core.App, config  StandaloneConfig) error {
	// Initialize the JSON RPC WebSocket server (but don't start it yet).
	rpcAddr := fmt.Sprintf(":%d", config.RPCPort)
	rpcHandler := &rpcHandler{
		app: app,
	}
	rpcServer, err := rpc.NewServer(rpcAddr, rpcHandler)
	if err != nil {
		return nil
	}
	go func() {
		// Wait for the server to start listening and select an address.
		for rpcServer.Addr() == nil {
			time.Sleep(10 * time.Millisecond)
		}
		log.WithField("address", rpcServer.Addr().String()).Info("started RPC server")
	}()
	return rpcServer.Listen()
}

// AddPeer is called when an RPC client calls AddPeer,
func (handler *rpcHandler) AddPeer(peerInfo peerstore.PeerInfo) error {
	log.Debug("received AddPeer request via RPC")
	if err := handler.app.AddPeer(peerInfo); err != nil {
		log.WithField("error", err.Error()).Error("internal error in AddPeer RPC call")
		return constants.ErrInternal
	}
	return nil
}

