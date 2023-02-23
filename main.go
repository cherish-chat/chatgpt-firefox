package main

import (
	"encoding/json"
	"github.com/atotto/clipboard"
	"github.com/playwright-community/playwright-go"
	"github.com/sirupsen/logrus"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"
	"time"
)

var pageMap sync.Map
var sendMsgLock sync.Mutex

func main() {
	browser, err := launchBrowser()
	if err != nil {
		log.Fatalf("Error while launching browser: %v", err)
	}
	go func() {
		page, err := newPage(browser, "1")
		if err != nil {
			log.Fatalf("Error while creating new page: %v", err)
		}
		file, _ := os.OpenFile("./main.go", os.O_RDONLY, 0666)
		defer file.Close()
		var code string
		for {
			buf := make([]byte, 1024)
			n, err := file.Read(buf)
			if err != nil {
				break
			}
			code += string(buf[:n])
		}
		text := "请详细说明下面代码的作用：\n" + code
		_, err = sendMsg(page, text)
		if err != nil {
			logrus.Errorf("Error while sending message: %v", err)
		}
		page.Close()
	}()
	select {}
}

type EditThisCookieItem struct {
	Domain         string  `json:"domain"`
	ExpirationDate float64 `json:"expirationDate"`
	HostOnly       bool    `json:"hostOnly"`
	HttpOnly       bool    `json:"httpOnly"`
	Name           string  `json:"name"`
	Path           string  `json:"path"`
	SameSite       string  `json:"sameSite"`
	Secure         bool    `json:"secure"`
	Session        bool    `json:"session"`
	StoreId        string  `json:"storeId"`
	Value          string  `json:"value"`
	Id             int     `json:"id"`
}

type ConversationStreamResponseItem struct {
	Message struct {
		Id     string `json:"id"`
		Author struct {
			Role     string      `json:"role"`
			Name     interface{} `json:"name"`
			Metadata struct {
			} `json:"metadata"`
		} `json:"author"`
		CreateTime interface{} `json:"create_time"`
		UpdateTime interface{} `json:"update_time"`
		Content    struct {
			ContentType string   `json:"content_type"`
			Parts       []string `json:"parts"`
		} `json:"content"`
		EndTurn  bool    `json:"end_turn"`
		Weight   float64 `json:"weight"`
		Metadata struct {
			MessageType   string `json:"message_type"`
			ModelSlug     string `json:"model_slug"`
			FinishDetails struct {
				Type string `json:"type"`
				Stop string `json:"stop"`
			} `json:"finish_details"`
		} `json:"metadata"`
		Recipient string `json:"recipient"`
	} `json:"message"`
	ConversationId string      `json:"conversation_id"`
	Error          interface{} `json:"error"`
}

func (i *ConversationStreamResponseItem) String() string {
	bytes, err := json.Marshal(i)
	if err != nil {
		return ""
	}
	return string(bytes)
}

func (i *ConversationStreamResponseItem) Text() string {
	parts := i.Message.Content.Parts
	return strings.Join(parts, "\n")
}

func loadCookies(filepath string) ([]playwright.BrowserContextAddCookiesOptionsCookies, error) {
	// 读取文件
	file, err := os.OpenFile(filepath, os.O_RDONLY, 0666)
	if err != nil {
		logrus.Errorf("Error while opening file: %v", err)
		return nil, err
	}
	defer file.Close()
	// 解析文件
	var cookies []EditThisCookieItem
	err = json.NewDecoder(file).Decode(&cookies)
	if err != nil {
		logrus.Errorf("Error while decoding file: %v", err)
		return nil, err
	}
	var result []playwright.BrowserContextAddCookiesOptionsCookies
	for _, cookie := range cookies {
		site := cookie.SameSite
		// site 首字母大写
		switch site {
		case "lax":
			site = "Lax"
		case "strict":
			site = "Strict"
		default:
			site = "None"
		}
		attribute := playwright.SameSiteAttribute(site)
		result = append(result, playwright.BrowserContextAddCookiesOptionsCookies{
			Name:  playwright.String(cookie.Name),
			Value: playwright.String(cookie.Value),
			//URL:      playwright.String(`https://chat.openai.com`),
			Domain:   playwright.String(cookie.Domain),
			Path:     playwright.String(cookie.Path),
			Expires:  playwright.Float(cookie.ExpirationDate),
			HttpOnly: playwright.Bool(cookie.HttpOnly),
			Secure:   playwright.Bool(cookie.Secure),
			SameSite: &attribute,
		})
	}
	return result, nil
}

func newPage(browser playwright.BrowserContext, id string) (playwright.Page, error) {
	page, err := browser.NewPage()
	if err != nil {
		logrus.Errorf("Error while creating new page: %v", err)
		return nil, err
	}
	logrus.Info("New page created successfully")
	{
		// 设置cookie
		cookies, err := loadCookies("cookies.json")
		if err != nil {
			return nil, err
		}
		err = browser.AddCookies(cookies...)
		if err != nil {
			logrus.Errorf("Error while adding cookies: %v", err)
			return nil, err
		}
		logrus.Info("Cookies added successfully")
	}
	// 监听响应
	//page.On("response", func(response playwright.Response) {
	//	url := response.URL()
	//	// 打印
	//	logrus.Infof("Response from %s", url)
	//	// 如果是json
	//	value, err := response.HeaderValue("content-type")
	//	if err != nil {
	//		logrus.Errorf("Error while getting header value: %v", err)
	//		return
	//	}
	//	if strings.Contains(value, "application/json") {
	//		// 获取body
	//		body, err := response.Text()
	//		if err != nil {
	//			logrus.Errorf("Error while getting response body: %v", err)
	//			return
	//		}
	//		logrus.Infof("Response body: %s", body)
	//	}
	//})
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
	pageMap.Store(id, page)
	return page, nil
}

func launchBrowser() (playwright.BrowserContext, error) {
	runOPtions := &playwright.RunOptions{
		Browsers: []string{"firefox"},
		Verbose:  false,
	}
	err := playwright.Install(runOPtions)
	if err != nil {
		logrus.Errorf("Error while installing playwright: %v", err)
		return nil, err
	}
	logrus.Info("Playwright installed successfully")
	pw, err := playwright.Run()
	if err != nil {
		logrus.Errorf("Error while running playwright: %v", err)
		return nil, err
	}
	logrus.Info("Playwright running successfully")
	args := []string{
		//"--no-sandbox",
		//"--disable-setuid-sandbox",
		//"--disable-infobars",
		//"--disable-dev-shm-usage",
		//"--disable-blink-features=AutomationControlled",
		//"--ignore-certificate-errors",
		//"--no-first-run",
		//"--no-service-autorun",
		//"--password-store=basic",
		//"--system-developer-mode",
		//// the following flags all try to reduce memory
		//// "--single-process",
		//"--mute-audio",
		//"--disable-default-apps",
		//"--no-zygote",
		//"--disable-accelerated-2d-canvas",
		//"--disable-web-security",
		//// "--disable-gpu"
		//// "--js-flags="--max-old-space-size=1024""
	}
	os.RemoveAll("/tmp/chatgpt")
	browser, err := pw.Firefox.LaunchPersistentContext("/tmp/chatgpt", playwright.BrowserTypeLaunchPersistentContextOptions{
		Args:     args,
		Headless: playwright.Bool(false),
	})
	if err != nil {
		logrus.Errorf("Error while launching browser: %v", err)
		return nil, err
	}
	logrus.Info("Browser launched successfully")
	return browser, nil
}

func sendMsg(page playwright.Page, inputText string) (string, error) {
	// 等待 //textarea
	textarea, err := page.WaitForSelector("//textarea", playwright.PageWaitForSelectorOptions{
		Timeout: playwright.Float(1000 * 60),
	})
	if err != nil {
		logrus.Errorf("Error while waiting for selector: %v", err)
		return "", err
	}
	// 输入
	//inputText := "你好，我是OpenAI生产的复读机，可以复读你说的话。"
	// 复制 inputText 粘贴到 textarea
	sendMsgLock.Lock()
	defer sendMsgLock.Unlock()
	err = clipboard.WriteAll(inputText)
	if err != nil {
		logrus.Errorf("Error while writing to clipboard: %v", err)
		return "", err
	}
	err = textarea.Focus()
	if err != nil {
		logrus.Errorf("Error while focusing textarea: %v", err)
		return "", err
	}
	err = textarea.Press(controlA())
	if err != nil {
		logrus.Errorf("Error while pressing control+a: %v", err)
		return "", err
	}
	err = textarea.Press(controlV())
	if err != nil {
		logrus.Errorf("Error while pressing control+v: %v", err)
		return "", err
	}
	go func() {
		for {
			time.Sleep(time.Second)
			// 回车
			// //textarea/../button/@disabled 是否有值 如果有说明此时不能回车
			selector, err := page.QuerySelector("//textarea/../button/@disabled")
			if err != nil {
				logrus.Errorf("Error while querying selector: %v", err)
				continue
			}
			if selector != nil {
				continue
			}
			err = textarea.Press("Enter")
			if err != nil {
				logrus.Errorf("Error while pressing enter: %v", err)
			}
			break
		}
	}()
	// 等待响应
	response := page.WaitForResponse("https://chat.openai.com/backend-api/conversation", playwright.PageWaitForResponseOptions{
		Timeout: playwright.Float(1000 * 60),
	})
	// 解析 text/event-stream
	{
		text, _ := response.Text()
		// 换行符分割，去掉 data:
		lines := strings.Split(text, "\n")
		var finalLine *ConversationStreamResponseItem
		for _, line := range lines {
			if strings.HasPrefix(line, "data:") {
				line = strings.TrimPrefix(line, "data:")
				// 解析json
				var data = &ConversationStreamResponseItem{}
				err := json.Unmarshal([]byte(line), data)
				if err != nil {
					continue
				}
				finalLine = data
			}
		}
		if finalLine != nil {
			inputText = finalLine.Text()
			logrus.Infof("AI: %s", finalLine.Text())
		} else {
			inputText = ""
		}
	}
	return inputText, nil
}

func controlA() string {
	// 判断系统
	if runtime.GOOS == "darwin" {
		return "Meta+A"
	}
	return "Control+A"
}

func controlV() string {
	// 判断系统
	if runtime.GOOS == "darwin" {
		return "Meta+V"
	}
	return "Control+V"
}
