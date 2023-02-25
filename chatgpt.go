package chatgpt

import (
	"github.com/playwright-community/playwright-go"
	"sync"
)

type Helper struct {
	CookiePath  string
	browser     playwright.BrowserContext
	pageMap     sync.Map
	sendMsgLock sync.Mutex
}

func NewHelper(cookiePath string) *Helper {
	h := &Helper{CookiePath: cookiePath}
	return h
}

func (h *Helper) IsLocked() bool {
	locked := h.sendMsgLock.TryLock()
	if locked {
		h.sendMsgLock.Unlock()
	}
	return !locked
}

func (h *Helper) ClosePage(id string, page playwright.Page) {
	page.Close()
	h.pageMap.Delete(id)
}
