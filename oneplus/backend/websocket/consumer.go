package websocket

import (
	"context"
	"encoding/json"
	"github.com/pmker/onegw/oneplus/backend/common"
	"github.com/pmker/onegw/oneplus/backend/utils"
)

// StartConsumer initializes a queue instance and ready events from it
func startConsumer(ctx context.Context, queue common.IQueue) {
	for {
		select {
		case <-ctx.Done():
			utils.Infof("Websocket Consumer Exit")
			return
		default:

			// This method should not block this go thread all the time to make it has chance to exit gracefully
			msg, err := queue.Pop()
			if err != nil {
				utils.Errorf("read message error %v", err)
				continue
			}

			utils.Debugf("rec msg: %s", string(msg))

			var wsMsg common.WebSocketMessage

			_ = json.Unmarshal(msg, &wsMsg)

			channel := findChannel(wsMsg.ChannelID)

			if channel == nil {
				channel = createChannelByID(wsMsg.ChannelID)
			}

			channel.AddMessage(&wsMsg)
		}
	}
}
