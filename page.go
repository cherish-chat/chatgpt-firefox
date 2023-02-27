package chatgpt

import (
	"github.com/cherish-chat/chatgpt/config"
	"github.com/playwright-community/playwright-go"
	"github.com/sirupsen/logrus"
	"time"
)

func (h *Helper) MustGetPage(id string) playwright.Page {
	if page, ok := h.pageMap.Load(id); ok {
		return page.(playwright.Page)
	}
	var page playwright.Page
	for {
		var err error
		page, err = h.NewPage(id)
		if err != nil {
			logrus.Errorf("Error while creating new page: %v", err)
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}
	return page
}

func (h *Helper) NewPage(id string) (playwright.Page, error) {
	page, err := h.browser.NewPage()
	if err != nil {
		logrus.Errorf("Error while creating new page: %v", err)
		return nil, err
	}
	logrus.Info("New page created successfully")
	{
		// 设置cookie
		cookies, err := config.LoadCookies("cookies.json")
		if err != nil {
			return nil, err
		}
		err = h.browser.AddCookies(cookies...)
		if err != nil {
			logrus.Errorf("Error while adding cookies: %v", err)
			return nil, err
		}
		logrus.Info("Cookies added successfully")
	}
	_, err = page.Goto("https://chat.openai.com/chat")
	if err != nil {
		page.Close()
		logrus.Errorf("Error while navigating to google: %v", err)
		return nil, err
	}

	logrus.Info("Navigated to openai successfully")
	h.pageMap.Store(id, page)
	return page, nil
}
