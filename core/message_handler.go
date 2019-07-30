// +build !js

package core

import (
	"github.com/pmker/onegw/p2p"
)

// Ensure that App implements p2p.MessageHandler.
var _ p2p.MessageHandler = &App{}

func (app *App) GetMessagesToShare(max int) ([][]byte, error) {

	messageData := make([][]byte, 1)

	return messageData, nil
}

func (app *App) HandleMessages(messages []*p2p.Message) error {

	return nil
}
