package main

import (
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
	"github.com/zs2619/kissnet-go"
	"github.com/zs2619/kissrtm/common"
	"github.com/zs2619/kissrtm/rtm"
	"github.com/zs2619/kissrtm/rtmnats"
)

var gAcceptor kissnet.IAcceptor

func main() {
	conf, err := common.LoadWebConfig("assets/config.json")
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("common.LoadWebConfig error")
		return
	}
	natsURI := os.Getenv("KISSRTM_NATSURI")
	if len(natsURI) == 0 {
		natsURI = conf.NatsURI
	}
	err = rtmnats.InitNats(conf.NatsURI)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Info("rtm.InitNats error")
		return
	}
	go rtm.RunNatsSubRTM()
	go rtm.RunNatsSubGameEvent()

	rtmPort := 0
	rtmPortStr := os.Getenv("KISSRTM_RTMSERVERPORT")
	if len(rtmPortStr) == 0 {
		rtmPort = conf.RTMServerPort
	} else {
		rtmPort, err = strconv.Atoi(rtmPortStr)
		if err != nil {
			logrus.WithFields(logrus.Fields{
				"error": err,
			}).Info("os.Getenv(KISSRTM_RTMSERVERPORT) error")
			return
		}
	}

	event := kissnet.NewNetEvent()
	logrus.Info("acceptor start")
	gAcceptor, err := kissnet.AcceptorFactory(
		conf.RTMServerType,
		rtmPort,
		rtm.RTMClientCB,
	)
	if err != nil {
		logrus.WithFields(logrus.Fields{
			"error": err,
		}).Fatal("AcceptorFactory error")
		return
	}

	gAcceptor.Run()
	event.EventLoop()
	gAcceptor.Close()
	logrus.Info("acceptor end")
}
