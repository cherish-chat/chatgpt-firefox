package main

import (
	"github.com/cherish-chat/xxim-server/common/pb"
	"github.com/cherish-chat/xxim-server/common/utils"
	"github.com/cherish-chat/xxim-server/sdk/svc"
	"github.com/zeromicro/go-zero/core/logx"
	"google.golang.org/protobuf/proto"
	"nhooyr.io/websocket"
	"time"
)

type eventHandler struct {
	svcCtx  *svc.ServiceContext
	msgChan chan *pb.MsgData
}

func (h *eventHandler) BeforeClose(code websocket.StatusCode, reason string) {

}

func (h *eventHandler) AfterClose(code websocket.StatusCode, reason string) {
	// 重连
	time.Sleep(time.Second * 3)
	h.svcCtx.Client().ReConnect()
}

func (h *eventHandler) BeforeReConnect() {
}

func (h *eventHandler) AfterReConnect() {
}

func (h *eventHandler) BeforeConnect() {
}

func (h *eventHandler) AfterConnect() {
}

func (h *eventHandler) OnMessage(typ websocket.MessageType, message []byte) {
}

func (h *eventHandler) OnPushMsgDataList(body *pb.PushBody) {
	msgDataList := &pb.MsgDataList{}
	err := proto.Unmarshal(body.Data, msgDataList)
	if err != nil {
		logx.Errorf("unmarshal MsgDataList error: %s", err.Error())
		return
	}
	for _, msgData := range msgDataList.MsgDataList {
		if msgData.SenderId != h.svcCtx.Config.Client.UserConfig.UserId {
			if msgData.ContentType == 11 {
				logx.Infof("收到消息: %s", string(msgData.Content))
				h.msgChan <- msgData
			}
		}
	}
}

func (h *eventHandler) OnPushNoticeData(body *pb.PushBody) {
}

func (h *eventHandler) OnPushResponseBody(body *pb.PushBody) {
}

func (h *eventHandler) OnTimer() {
	h.sync()
}

func (h *eventHandler) sync() {
	client := h.svcCtx.Client()
	getFriendListResp := &pb.GetFriendListResp{}
	err := client.RequestX("/v1/relation/getFriendList", &pb.GetFriendListReq{
		Page: &pb.Page{Page: 1, Size: 100000},
		Opt:  pb.GetFriendListReq_OnlyId,
	}, getFriendListResp)
	if err != nil {
		logx.Errorf("getFriendList error: %s", err.Error())
		return
	}
	getGroupListResp := &pb.GetMyGroupListResp{}
	err = client.RequestX("/v1/group/getMyGroupList", &pb.GetMyGroupListReq{
		Page: &pb.Page{Page: 1, Size: 100000},
		Opt:  pb.GetMyGroupListReq_ONLY_ID,
	}, getGroupListResp)
	if err != nil {
		logx.Errorf("getGroupList error: %s", err.Error())
		return
	}
	var convIds []string
	for _, v := range getFriendListResp.Ids {
		convIds = append(convIds, pb.SingleConvId(h.svcCtx.Config.Client.UserConfig.UserId, v))
	}
	for _, v := range getGroupListResp.Ids {
		convIds = append(convIds, pb.GroupConvId(v))
	}
	convIds = utils.Set(convIds)
	if len(convIds) == 0 {
		return
	}
	resp := &pb.BatchGetConvSeqResp{}
	err = client.RequestX("/v1/msg/batchGetConvSeq", &pb.BatchGetConvSeqReq{
		ConvIdList: convIds,
		CommonReq:  nil,
	}, resp)
	if err != nil {
		logx.Errorf("batchGetConvSeq error: %s", err.Error())
		return
	}
	for convId, convSeq := range resp.ConvSeqMap {
		logx.Debugf("convId: %s, convSeq: %#v", convId, convSeq)
	}
}

func (h *eventHandler) sendTextMsg(convId string, text string) error {
	return h.svcCtx.Client().RequestX("/v1/msg/sendMsgList", &pb.SendMsgListReq{
		MsgDataList:  []*pb.MsgData{h.genTextMsg(convId, text)},
		DeliverAfter: nil,
		CommonReq:    nil,
	}, &pb.SendMsgListResp{})
}

func (h *eventHandler) genTextMsg(convId string, text string) *pb.MsgData {
	return &pb.MsgData{
		ClientMsgId: utils.GenId(),
		ClientTime:  utils.AnyToString(time.Now().UnixMilli()),
		SenderId:    h.svcCtx.Config.Client.UserConfig.UserId,
		SenderInfo:  []byte(`{"name":"昵称", "avatar":"头像"}`),
		ConvId:      convId,
		AtUsers:     nil,
		ContentType: 11,
		Content:     []byte(text),
		Seq:         "",
		Options: &pb.MsgData_Options{
			StorageForServer:  true,
			StorageForClient:  true,
			NeedDecrypt:       false,
			OfflinePush:       true,
			UpdateConvMsg:     true,
			UpdateUnreadCount: true,
		},
		OfflinePush: &pb.MsgData_OfflinePush{
			Title:   "昵称",
			Content: text,
			Payload: "",
		},
		Ext: nil,
	}
}

func newEventHandler(svcCtx *svc.ServiceContext) *eventHandler {
	return &eventHandler{svcCtx: svcCtx, msgChan: make(chan *pb.MsgData)}
}
