# chatgpt

使用chrome webdriver调用chatgpt 支持plus

## 使用方法

1. 架梯子访问 https://chat.openai.com/chat ，登录并能成功进入聊天界面
2. 安装chrome插件：EditThisCookie https://chrome.google.com/webstore/detail/editthiscookie/fngmhnnpilhplaeedifhccceomclgfbg
3. 导出当前页面的cookie，保存到 cookies.json
4. 参考example/main.go，编写自己的聊天机器人

```go
package main

import (
	"github.com/cherish-chat/chatgpt-firefox"
	"log"
)

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
						eh.sendTextMsg(msgData.ConvId, reply)
					}
				}
            }()
		}
	}
}

```