package rtmnats

import (
	"encoding/json"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/zs2619/kissrtm/common"
)

var NatsConn *nats.Conn

func InitNats(natsURI string) error {
	var err error
	NatsConn, err = nats.Connect(natsURI)
	if err != nil {
		return err
	}
	logrus.WithFields(logrus.Fields{
		"URI": natsURI,
	}).Info("InitNats")

	return nil
}

func FiniNats() error {
	NatsConn.Close()
	logrus.WithFields(logrus.Fields{}).Info("FiniNats")
	return nil
}

func PublishNatsNotifyEvent(userIDList []string, EventType int32) error {
	msg, _ := json.Marshal(common.NotifyEvent{UserIDList: userIDList, EventType: EventType})
	err := NatsConn.Publish(common.GlobalMessageChannel, msg)
	return err
}
