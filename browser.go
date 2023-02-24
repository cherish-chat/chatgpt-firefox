package chatgpt

import (
	"github.com/playwright-community/playwright-go"
	"github.com/sirupsen/logrus"
	"os"
)

func (h *Helper) LaunchBrowser() error {
	runOptions := &playwright.RunOptions{
		Browsers: []string{"firefox"},
		Verbose:  false,
	}
	err := playwright.Install(runOptions)
	if err != nil {
		logrus.Errorf("Error while installing playwright: %v", err)
		return err
	}
	logrus.Info("Playwright installed successfully")
	pw, err := playwright.Run()
	if err != nil {
		logrus.Errorf("Error while running playwright: %v", err)
		return err
	}
	logrus.Info("Playwright running successfully")
	args := []string{}
	os.RemoveAll("./cache")
	browser, err := pw.Firefox.LaunchPersistentContext("./cache", playwright.BrowserTypeLaunchPersistentContextOptions{
		Args:     args,
		Headless: playwright.Bool(false),
	})
	if err != nil {
		logrus.Errorf("Error while launching browser: %v", err)
		return err
	}
	logrus.Info("Browser launched successfully")
	h.browser = browser
	return nil
}
