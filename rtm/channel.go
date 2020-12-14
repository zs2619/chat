package rtm

import (
	"bytes"
	"sync"
	"sync/atomic"

	"github.com/sirupsen/logrus"
	"github.com/zs2619/kissrtm/pb"
)

type RTMChannelMsg struct {
	msg        *bytes.Buffer
	sendUserID string
}

type RTMChannel struct {
	msgChan   chan *RTMChannelMsg
	subscribe sync.Map
	channelID string
	count     int64
}

func (this *RTMChannel) start() {
	logrus.WithFields(logrus.Fields{
		"channelID": this.channelID,
		"count":     this.count,
	}).Info("start RTMChannel")
	go this.run()
}

func (this *RTMChannel) run() {
	for ccMsg := range this.msgChan {
		if ccMsg == nil {
			break
		}
		i := 0
		this.subscribe.Range(func(k, v interface{}) bool {
			u := v.(*UserRTM)
			if ccMsg.sendUserID != u.UserID {
				u.SendMsg(pb.S2C_MsgID_RecvRTM, ccMsg.msg.Bytes())
			}
			i++
			return true
		})
		if i == 0 {
			break
		}
		//
	}
}

func (this *RTMChannel) publish(ccmsg *RTMChannelMsg) {
	this.msgChan <- ccmsg
}

func (this *RTMChannel) close() {
	logrus.WithFields(logrus.Fields{
		"channelID": this.channelID,
	}).Info("quit RTMChannel")
	this.msgChan <- nil
}

type RTMChannelMgr struct {
	mutex    sync.RWMutex
	rtmIDMap map[string]*RTMChannel
}

var rtmChannelMgr = &RTMChannelMgr{
	rtmIDMap: make(map[string]*RTMChannel),
}

func (this *RTMChannelMgr) getRTMChan(channelID string) *RTMChannel {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	cc, ok := this.rtmIDMap[channelID]
	if ok {
		return cc
	}
	return nil
}

func (this *RTMChannelMgr) joinRTMChan(user *UserRTM, channelID string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	cc, ok := this.rtmIDMap[channelID]
	if ok {
		cc.subscribe.Store(user.UserID, user)
	} else {
		cc = &RTMChannel{
			msgChan:   make(chan *RTMChannelMsg, 2048),
			channelID: channelID,
		}
		this.rtmIDMap[channelID] = cc
		cc.subscribe.Store(user.UserID, user)
		cc.start()
	}
	logrus.WithFields(logrus.Fields{
		"userID":    user.UserID,
		"channelID": channelID,
	}).Info("user join rtmChan")
	atomic.AddInt64(&cc.count, 1)
}

func (this *RTMChannelMgr) quitRTMChan(user *UserRTM, channelID string) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	cc, ok := this.rtmIDMap[channelID]
	if ok {
		cc.subscribe.Delete(user.UserID)
		atomic.AddInt64(&cc.count, -1)
		if cc.count <= 0 {
			//subscribe 为0 清除channel
			cc.close()
			delete(this.rtmIDMap, channelID)
		}
	}
	logrus.WithFields(logrus.Fields{
		"userID":    user.UserID,
		"channelID": channelID,
	}).Info("user quit rtmChan")
}

func (this *RTMChannelMgr) Close() {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	for k, v := range this.rtmIDMap {
		v.close()
		delete(this.rtmIDMap, k)
	}
}
