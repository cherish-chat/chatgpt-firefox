package main

import (
	"github.com/cherish-chat/xxim-server/common/utils"
	"github.com/cherish-chat/xxim-server/sdk/config"
	"github.com/cherish-chat/xxim-server/sdk/conn"
	"github.com/cherish-chat/xxim-server/sdk/svc"
	"log"
	"strings"
)

func xxim() *eventHandler {
	conf := config.Config{
		Client: conn.Config{
			Addr: "wss://api.cherish.chat:443/ws",
			DeviceConfig: conn.DeviceConfig{
				PackageId:   utils.GenId(),
				Platform:    "macos",
				DeviceId:    utils.GenId(),
				DeviceModel: "macos golang",
				OsVersion:   "10.15.7",
				AppVersion:  "v1.0.0",
				Language:    "zh",
				NetworkUsed: "wifi",
				Ext:         nil,
			},
			UserConfig: conn.UserConfig{
				UserId:   "test123456",
				Password: utils.Md5("123456"),
				Token:    "",
				Ext:      nil,
			},
		},
	}

	svcCtx := svc.NewServiceContext(conf)
	eh := newEventHandler(svcCtx)

	svcCtx.SetEventHandler(eh)

	err := svcCtx.Client().Connect()
	if err != nil {
		log.Fatalf("connect error: %v", err)
	}
	return eh
}

func main() {
	helper := chatgpt.NewHelper("cookies.json")
	err := helper.LaunchBrowser()
	if err != nil {
		log.Fatalf("launch browser error: %v", err)
	}
	eh := xxim()
	for {
		select {
		case msgData := <-eh.msgChan:
			go func() {
				if helper.IsLocked() {
					eh.sendTextMsg(msgData.ConvId, "机器人正在忙，请稍后再试")
				} else {
					page := helper.MustGetPage(msgData.ConvId)
					reply, err := helper.SendMsg(page, string(msgData.Content))
					if err != nil {
						eh.sendTextMsg(msgData.ConvId, "机器人出错了，请稍后再试: "+err.Error())
					} else {
						if strings.TrimSpace(reply) == "" {
							eh.sendTextMsg(msgData.ConvId, "机器人没有回复")
							// 刷新页面
							helper.ClosePage(msgData.ConvId, page)
						} else {
							eh.sendTextMsg(msgData.ConvId, reply)
						}
					}
				}
			}()
		}
	}
}
