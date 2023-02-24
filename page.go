package chatgpt

import (
	"github.com/cherish-chat/chatgpt/config"
	"github.com/playwright-community/playwright-go"
	"github.com/sirupsen/logrus"
	"log"
)

func (h *Helper) MustGetPage(id string) playwright.Page {
	if page, ok := h.pageMap.Load(id); ok {
		return page.(playwright.Page)
	}
	page, err := h.NewPage(id)
	if err != nil {
		log.Fatalf("Error while creating new page: %v", err)
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
		logrus.Errorf("Error while navigating to google: %v", err)
		return nil, err
	}

	logrus.Info("Navigated to google successfully")
	response := page.WaitForResponse("https://chat.openai.com/backend-api/conversations*", playwright.PageWaitForResponseOptions{
		Timeout: playwright.Float(1000 * 60 * 60 * 24),
	})
	logrus.Info("Selector found successfully")
	// 打印
	{
		url := response.URL()
		body, _ := response.Text()
		logrus.Infof("Response from %s, body: %s", url, body)
	}
	h.pageMap.Store(id, page)
	return page, nil
}
