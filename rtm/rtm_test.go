package rtm

import (
	"bytes"
	"testing"
	"time"

	"github.com/zs2619/kissnet-go"
	"github.com/zs2619/kissrtm/pb"
	"github.com/zs2619/kissrtm/test"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"google.golang.org/protobuf/proto"
)

type RTMTestSuite struct {
	test.BaseTestSuite
}

func TestRTMest(t *testing.T) {
	suite.Run(t, new(RTMTestSuite))
}

type dummyCoon struct {
	kissnet.IConnection
}

func (this *dummyCoon) start() {
}

func (this *dummyCoon) getID() int64 {
	return int64(0)
}
func (this *dummyCoon) setID(id int64) {
}
func (this *dummyCoon) SendMsg(msg *bytes.Buffer) error {
	logrus.Debug("SendMsg")
	return nil
}

var conn dummyCoon

func (t *RTMTestSuite) TestRTM() {
	user := &UserRTM{UserID: "UserID", Conn: &conn}

	rtmChannelMgr.joinRTMChan(user, "123")
	rtmChan := rtmChannelMgr.getRTMChan("123")
	t.NotNil(rtmChan)
	t.Equal(int64(1), rtmChan.count)
	rtmChannelMgr.quitRTMChan(user, "123")
	rtmChan = rtmChannelMgr.getRTMChan("123")
	t.Nil(rtmChan)

	//加入房间222
	pbMsg := &pb.C2S_JoinRoomChan{
		RoomID: "222",
	}
	msg, _ := proto.Marshal(pbMsg)
	err := JoinRoomChan(user, msg)
	t.Nil(err)
	t.Equal(user.RoomChannel, "222")

	rtmChan = rtmChannelMgr.getRTMChan("222")
	t.NotNil(rtmChan)
	t.Equal(rtmChan.count, int64(1))

	rtmChan = rtmChannelMgr.getRTMChan("333")
	t.Nil(rtmChan)
	//加入新房间333 频道前替换旧房间222
	pbMsg = &pb.C2S_JoinRoomChan{
		RoomID: "333",
	}
	msg, _ = proto.Marshal(pbMsg)
	err = JoinRoomChan(user, msg)
	t.Nil(err)
	t.Equal(user.RoomChannel, "333")

	rtmChan = rtmChannelMgr.getRTMChan("222")
	t.Nil(rtmChan)

	rtmChan = rtmChannelMgr.getRTMChan("333")
	t.NotNil(rtmChan)
	t.Equal(rtmChan.count, int64(1))

	rtmChannelMgr.quitRTMChan(user, "333")
	rtmChan = rtmChannelMgr.getRTMChan("333")
	t.Nil(rtmChan)
}

func (t *RTMTestSuite) TestRRTMg() {
	user := &UserRTM{UserID: "UserID", Conn: &conn}
	rtmChannelMgr.joinRTMChan(user, "123")

	user1 := &UserRTM{UserID: "UserID1", Conn: &conn}
	rtmChannelMgr.joinRTMChan(user1, "123")

	rtmChan := rtmChannelMgr.getRTMChan("123")
	t.NotNil(rtmChan)
	t.Equal(int64(2), rtmChan.count)

	pbMsg := &pb.C2S_RTMMsgSend{
		RtmType:    pb.RTMMsgType_Room,
		SendName:   "SendName",
		SendUserID: "UserID1",
		ConvID:     "123",
		Msg:        "hello world",
	}
	msg, _ := proto.Marshal(pbMsg)
	ccMsg := &RTMChannelMsg{
		msg:        bytes.NewBuffer(msg),
		sendUserID: "UserID1",
	}
	rtmChan.publish(ccMsg)

	SendRTM("UserID1", msg)

	time.Sleep(10 * time.Millisecond)
	rtmChannelMgr.quitRTMChan(user, "123")
	rtmChannelMgr.quitRTMChan(user1, "123")
}
