package rtm

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"

	"github.com/zs2619/kissrtm/common"
	"github.com/zs2619/kissrtm/rtmnats"

	"github.com/sirupsen/logrus"
	"github.com/zs2619/kissnet-go"
	"github.com/zs2619/kissrtm/pb"
	"google.golang.org/protobuf/proto"
)

type FuncMsg func(*UserRTM, []byte) error
type NatsFuncMsg func(string, []byte) error

var (
	msgMap    map[pb.C2S_MsgID_MsgID]FuncMsg     = make(map[pb.C2S_MsgID_MsgID]FuncMsg)
	msgSubMap map[pb.C2S_MsgID_MsgID]NatsFuncMsg = make(map[pb.C2S_MsgID_MsgID]NatsFuncMsg)
)

func init() {
	RegisterMsgMapFunc()
}

func RegisterMsgMapFunc() {
	msgMap[pb.C2S_MsgID_SendRTM] = PublishSendRTM
	msgMap[pb.C2S_MsgID_JoinRoomChan] = JoinRoomChan
	msgMap[pb.C2S_MsgID_QuitRoomChan] = QuitRoomChan
	///订阅广播消息
	msgSubMap[pb.C2S_MsgID_SendRTM] = SendRTM
}

func RTMClientCB(conn kissnet.IConnection, msg []byte) error {
	if msg == nil {
		//退出
		userRTMMgr.DelUserRTM(conn)
		return nil
	}
	if len(msg) < 2 {
		userRTMMgr.DelUserRTM(conn)
		return fmt.Errorf("msg len error")
	}
	msgID := binary.LittleEndian.Uint16(msg)
	if msgID == uint16(pb.C2S_MsgID_LoginRTM) {
		loginMsg := &pb.C2S_RTMMsgLogin{}
		err := proto.Unmarshal(msg[2:], loginMsg)
		if err != nil {
			return err
		}
		uid := loginMsg.UserSession

		u := userRTMMgr.GetUserRTMByUserID(uid)
		if u != nil {
			// 踢掉之前连接
			userRTMMgr.DelUserRTM(u.Conn)
		}
		u = userRTMMgr.AddUserRTM(uid, conn)

	} else {
		err := ProcMsg(conn, msgID, msg[2:])
		if err != nil {
		}
	}
	return nil
}
func ProcMsg(conn kissnet.IConnection, msgID uint16, msg []byte) error {
	f, ok := msgMap[pb.C2S_MsgID_MsgID(msgID)]
	if !ok {
		return fmt.Errorf("(%d) msgID nil", msgID)
	}
	u := userRTMMgr.GetUserRTMByConn(conn)
	if u == nil {
		return fmt.Errorf("(%s) (%d) ProcMsg GetUserRTM nil", u.UserID, msgID)
	}
	return f(u, msg)
}

func procGameEventMsg(e *common.NotifyEvent) {
	pbMsg := &pb.S2C_EventNotify{
		EventType: e.EventType,
		Msg:       e.Msg,
	}
	m, _ := proto.Marshal(pbMsg)
	if e.Broadcast {
		for _, v := range userRTMMgr.userIDMap {
			v.SendMsg(pb.S2C_MsgID_EventNotify, m)
		}
	} else {
		for _, v := range e.UserIDList {
			u := userRTMMgr.GetUserRTMByUserID(v)
			if u != nil {
				u.SendMsg(pb.S2C_MsgID_EventNotify, m)
				logrus.WithFields(logrus.Fields{
					"GameEvent": e,
				}).Info("procGameEventMsg")
			}
		}
	}
}
func publishNatsRTMMsg(userID string, msgID uint16, msg []byte) error {
	rtmMsg := &pb.C2S_RTMMsgSend{}
	err := proto.Unmarshal(msg, rtmMsg)
	if err != nil {
		return err
	}
	if len(rtmMsg.Msg) == 0 {
		return fmt.Errorf("msg nil (%s) (%d)", userID, msgID)
	}

	newMsg, err := proto.Marshal(rtmMsg)
	if err != nil {
		return err
	}
	msgJSON, err := json.Marshal(common.RTMEvent{SendUserID: userID, MsgID: msgID, Msg: newMsg})
	if err != nil {
		return err
	}

	err = rtmnats.NatsConn.Publish(common.RTMMessageChannel, msgJSON)
	return err
}

func procSubRTMMsg(sendUserID string, msgID uint16, msg []byte) error {
	f, ok := msgSubMap[pb.C2S_MsgID_MsgID(msgID)]
	if !ok {
		return fmt.Errorf("msgID nil (%s) (%d)", sendUserID, msgID)
	}

	logrus.WithFields(logrus.Fields{
		"userID": sendUserID,
		"msgID":  msgID,
	}).Info("procSubRTMMsg")

	return f(sendUserID, msg)
}

func PublishSendRTM(user *UserRTM, msg []byte) error {
	err := publishNatsRTMMsg(user.UserID, uint16(pb.C2S_MsgID_SendRTM), msg)
	if err != nil {
		return err
	}
	return nil
}
func SendRTM(sendUserID string, msg []byte) error {
	rtmMsg := &pb.C2S_RTMMsgSend{}
	err := proto.Unmarshal(msg, rtmMsg)
	if err != nil {
		return err
	}
	logrus.Info(rtmMsg)
	if rtmMsg.RtmType == pb.RTMMsgType_Private {
		u := userRTMMgr.GetUserRTMByUserID(rtmMsg.ConvID)
		if u != nil {
			u.RecvRTMMsg(rtmMsg)
		}
	} else if rtmMsg.RtmType == pb.RTMMsgType_Room {
		roomID := rtmMsg.ConvID
		if len(roomID) == 0 {
			return nil
		}
		logrus.Info(roomID)
		cc := rtmChannelMgr.getRTMChan(roomID)
		if cc != nil {
			ccMsg := &RTMChannelMsg{
				msg:        bytes.NewBuffer(msg),
				sendUserID: rtmMsg.SendUserID,
			}
			cc.publish(ccMsg)
		}
	}

	return nil
}

func JoinRoomChan(user *UserRTM, msg []byte) error {
	joinRoom := &pb.C2S_JoinRoomChan{}
	err := proto.Unmarshal(msg, joinRoom)
	if err != nil {
		user.SendJoinRoomFailed()
		return err
	}
	if len(joinRoom.RoomID) == 0 {
		user.SendJoinRoomFailed()
		return fmt.Errorf("(%s) len(joinRoom.RoomID) == 0", user.UserID)
	}
	if len(user.RoomChannel) != 0 {
		//退出房间
		rtmChannelMgr.quitRTMChan(user, user.RoomChannel)
		user.RoomChannel = ""
	}

	rtmChannelMgr.joinRTMChan(user, joinRoom.RoomID)
	user.RoomChannel = joinRoom.RoomID
	user.SendJoinRoomOk()
	return nil
}
func QuitRoomChan(user *UserRTM, msg []byte) error {
	quitRoom := &pb.C2S_JoinRoomChan{}
	err := proto.Unmarshal(msg, quitRoom)
	if err != nil {
		return err
	}
	if len(quitRoom.RoomID) == 0 {
		return fmt.Errorf("(%s) len(quitRoom.RoomID) == 0", user.UserID)
	}
	rtmChannelMgr.quitRTMChan(user, quitRoom.RoomID)
	user.RoomChannel = ""
	return nil
}
func JoinGuildChan(*UserRTM, []byte) error {
	return nil
}
func QuitGuildChan(*UserRTM, []byte) error {
	return nil
}
