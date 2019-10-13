package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	log "github.com/sirupsen/logrus"
)

type ChannelInfo struct {
	Name         string           `json:"name"`
	MsgCount     int64            `json:"msgs"`
	LastSequence int64            `json:"last_seq"`
	Subscriber   []SubscriberInfo `json:"subscriptions"`
}

type SubscriberInfo struct {
	ClientID     string `json:"client_id"`
	QueueName    string `json:"queue_name"`
	Inbox        string `json:"inbox"`
	AckInbox     string `json:"ack_inbox"`
	IsDurable    bool   `json:"is_durable"`
	IsOffline    bool   `json:"is_offline"`
	MaxInflight  int    `json:"max_inflight"`
	LastSent     int64  `json:"last_sent"`
	PendingCount int    `json:"pending_count"`
	IsStalled    bool   `json:"is_stalled"`
}

func init() {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true})

	// Output to stdout instead of the default stderr, could also be a file.
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)

}

var DurableName = "ImDurable"
var QueueGroup = ":grp1"

func main() {
	var channelInfo = &ChannelInfo{}

	for {
		resp, err := http.Get("http://10.152.183.67:8222/streaming/channelsz?channel=Test&subs=1")

		if err != nil {
			log.Error("unable to access the broker")
			panic(err)
		}

		defer resp.Body.Close()

		json.NewDecoder(resp.Body).Decode(&channelInfo)

		log.Debugf("Max Messages %s", getTotalMessages(*channelInfo))
		log.Debugf("Message Lag %s", getMaxMsgLag(*channelInfo))
		time.Sleep(5 * time.Second)
	}
}

func getTotalMessages(channelInfo ChannelInfo) int64 {
	return channelInfo.MsgCount
}

func getMaxMsgLag(channelInfo ChannelInfo) int64 {
	var maxValue int64
	maxValue = 0
	for _, subs := range channelInfo.Subscriber {
		if subs.LastSent > maxValue && subs.QueueName == (DurableName+QueueGroup) {
			maxValue = subs.LastSent
		}
	}

	return channelInfo.MsgCount - maxValue
}
