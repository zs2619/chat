package rtm

import (
	"encoding/json"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/sirupsen/logrus"
	"github.com/zs2619/kissrtm/common"
	"github.com/zs2619/kissrtm/rtmnats"
)

const subTimeOut = 30

func RunNatsSubGameEvent() {
	logrus.Info("start nats Subscribe GameEvent")
	for {
		sub, err := rtmnats.NatsConn.SubscribeSync(common.GlobalMessageChannel)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("NatsConn.SubscribeSync GameEvent Error")
			break
		}
		for {
			msg, err := sub.NextMsg(subTimeOut * time.Second)
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				} else {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("sub.NextMsg error")
					break
				}
			}
			e := common.NotifyEvent{}
			err = json.Unmarshal(msg.Data, &e)
			if err != nil {
				logrus.WithFields(logrus.Fields{}).Error("json.Unmarshal")
			} else {
				procGameEventMsg(&e)
			}
		}
	}
}

func RunNatsSubRTM() {
	logrus.Info("start nats Subscribe RTM")
	for {
		sub, err := rtmnats.NatsConn.SubscribeSync(common.RTMMessageChannel)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Error("NatsConn.SubscribeSync RTM Error")
			break
		}
		for {
			msg, err := sub.NextMsg(subTimeOut * time.Second)
			if err != nil {
				if err == nats.ErrTimeout {
					continue
				} else {
					logrus.WithFields(logrus.Fields{
						"error": err,
					}).Error("sub.NextMsg error")
					break
				}
			}
			e := common.RTMEvent{}
			err = json.Unmarshal(msg.Data, &e)
			if err != nil {
				logrus.WithFields(logrus.Fields{}).Error("json.Unmarshal")
			} else {
				err = procSubRTMMsg(e.SendUserID, e.MsgID, e.Msg)
				if err != nil {
				}
			}
		}
	}
}
