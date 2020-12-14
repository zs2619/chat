package rtm

import (
	"bytes"
	"encoding/binary"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/zs2619/kissnet-go"
	"github.com/zs2619/kissrtm/pb"
	"google.golang.org/protobuf/proto"
)

const (
	RoomChannel  = 1
	GuildChannel = 2
)

type UserRTM struct {
	Conn         kissnet.IConnection
	UserID       string
	RoomChannel  string
	GuildChannel string
	Channel      []string
}

func (this *UserRTM) SendLoginMsg() error {
	logrus.Debug("SendLoginMsg")
	pbMsg := &pb.S2C_LoginOK{UserID: this.UserID}
	m, _ := proto.Marshal(pbMsg)
	this.SendMsg(pb.S2C_MsgID_LoginOK, m)
	return nil
}
func (this *UserRTM) SendLoginFailed() error {
	pbMsg := &pb.S2C_LoginFail{}
	m, _ := proto.Marshal(pbMsg)
	this.SendMsg(pb.S2C_MsgID_LoginFail, m)
	return nil
}
func (this *UserRTM) SendJoinRoomOk() error {
	pbMsg := &pb.S2C_JoinRoomChanOK{}
	m, _ := proto.Marshal(pbMsg)
	this.SendMsg(pb.S2C_MsgID_JoinRoomChanOK, m)
	return nil
}
func (this *UserRTM) SendJoinRoomFailed() error {
	pbMsg := &pb.S2C_JoinRoomChanFailed{}
	m, _ := proto.Marshal(pbMsg)
	this.SendMsg(pb.S2C_MsgID_JoinRoomChanFailed, m)
	return nil
}

func (this *UserRTM) SendMsg(msgID pb.S2C_MsgID_MsgID, msg []byte) error {
	sendMsg := new(bytes.Buffer)
	binary.Write(sendMsg, binary.LittleEndian, uint16(msgID))
	sendMsg.Write(msg)
	this.Conn.SendMsg(sendMsg)
	return nil
}

func (this *UserRTM) RecvRTMMsg(msg *pb.C2S_RTMMsgSend) error {
	pbMsg := &pb.S2C_RTMMsgRecv{
		RtmType:    msg.RtmType,
		SendName:   msg.SendName,
		SendUserID: msg.SendUserID,
		ConvID:     msg.ConvID,
		Msg:        msg.Msg,
		SendTime:   time.Now().UTC().Unix(),
	}

	m, _ := proto.Marshal(pbMsg)
	this.SendMsg(pb.S2C_MsgID_RecvRTM, m)
	return nil
}

func (this *UserRTM) Init() {
	this.SendLoginMsg()
	this.loginNotify(this.UserID)

	logrus.WithFields(logrus.Fields{
		"userID": this.UserID,
	}).Info("UserRTM Init")
}

func (this *UserRTM) Dispose() {
	//退出channel
	logrus.Info("UserRTM Dispose" + this.UserID)
	if len(this.RoomChannel) != 0 {
		rtmChannelMgr.quitRTMChan(this, this.RoomChannel)
	}
	if len(this.GuildChannel) != 0 {
		rtmChannelMgr.quitRTMChan(this, this.GuildChannel)
	}
	this.RoomChannel = ""
	this.GuildChannel = ""
	for _, chID := range this.Channel {
		rtmChannelMgr.quitRTMChan(this, chID)
	}
	this.logoutnNotify(this.UserID)
}

func (this *UserRTM) loginNotify(uid string) {

}

func (this *UserRTM) logoutnNotify(uid string) {

}

type UserRTMMgr struct {
	userIDMap map[string]*UserRTM
	connMap   map[kissnet.IConnection]*UserRTM
	num       int64
	mutex     sync.RWMutex
}

var userRTMMgr *UserRTMMgr = &UserRTMMgr{
	userIDMap: make(map[string]*UserRTM),
	connMap:   make(map[kissnet.IConnection]*UserRTM),
	num:       int64(0),
}

func (this *UserRTMMgr) GetUserRTMByUserID(UserID string) *UserRTM {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if v, ok := this.userIDMap[UserID]; ok {
		return v
	}
	return nil
}
func (this *UserRTMMgr) GetUserRTMByConn(conn kissnet.IConnection) *UserRTM {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	if v, ok := this.connMap[conn]; ok {
		return v
	}
	return nil
}

func (this *UserRTMMgr) AddUserRTM(UserID string, conn kissnet.IConnection) *UserRTM {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	u := &UserRTM{
		Conn:   conn,
		UserID: UserID,
	}
	this.connMap[u.Conn] = u
	this.userIDMap[u.UserID] = u
	this.num++
	u.Init()
	return u
}

func (this *UserRTMMgr) DelUserRTM(conn kissnet.IConnection) {
	this.mutex.Lock()
	defer this.mutex.Unlock()
	u, ok := this.connMap[conn]
	if !ok {
		return
	}
	this.num--
	u.Dispose()

	delete(this.connMap, conn)
	delete(this.userIDMap, u.UserID)
}
func (this *UserRTMMgr) Close() {
	logrus.Info("UserRTMMgr Close")
	for k := range this.connMap {
		k.Close()
	}
}

func (this *UserRTMMgr) GetUserNum() int64 {
	this.mutex.RLock()
	defer this.mutex.RUnlock()
	return this.num
}
