package config

import (
	"encoding/json"
	"github.com/playwright-community/playwright-go"
	"github.com/sirupsen/logrus"
	"os"
)

func LoadCookies(filepath string) ([]playwright.BrowserContextAddCookiesOptionsCookies, error) {
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
