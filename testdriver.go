package testdriver

import (
	"flag"
	"fmt"
	"github.com/nshah/selenium"
	"regexp"
	"strings"
	"testing"
	"time"
)

var (
	webdriverRemoteUrl = flag.String(
		"webdriver.remote",
		"http://localhost:8080/wd/hub",
		"Remote webdriver URL.")
	webdriverQuit = flag.Bool(
		"webdriver.quit", true,
		"Determines if the browser will be quit at the end of a test.")
	webdriverProxy = flag.String(
		"webdriver.proxy",
		"",
		"Webdriver browser proxy.")
	webdriverImplicitWaitTimeout = flag.Duration(
		"webdriver.timeout.implicit-wait", 60*time.Second,
		"Webdriver implicit wait timeout.")
	webdriverAsyncScriptTimeout = flag.Duration(
		"webdriver.timeout.async-script", 60*time.Second,
		"Webdriver async script timeout.")
	browserSpec = flag.String(
		"webdriver.browsers",
		"firefox,chrome,iexplorer",
		"List of browser to run against.")
)

var browsers []string

func newRemote(browser string) (selenium.WebDriver, error) {
	caps := selenium.Capabilities{"browserName": browser}
	if *webdriverProxy != "" {
		proxy := make(map[string]string)
		proxy["proxyType"] = "MANUAL"
		proxy["httpProxy"] = *webdriverProxy
		proxy["sslProxy"] = *webdriverProxy
		caps["proxy"] = proxy
	}
	wd, err := selenium.NewRemote(caps, *webdriverRemoteUrl)
	if err != nil {
		return nil, fmt.Errorf("Can't start session %s for browser %s", err, browser)
	}
	err = wd.SetAsyncScriptTimeout(*webdriverAsyncScriptTimeout)
	if err != nil {
		return nil, fmt.Errorf("Can't set async script timeout %s", err)
	}
	err = wd.SetImplicitWaitTimeout(*webdriverImplicitWaitTimeout)
	if err != nil {
		return nil, fmt.Errorf("Can't set implicit wait timeout %s", err)
	}
	return wd, nil
}

func makeTestFunc(browser string, test func(*T)) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		wd, err := newRemote(browser)
		if err != nil {
			t.Fatalf("Failed to create remote: %s", err)
		}
		if *webdriverQuit {
			defer wd.Quit()
		}
		test(&T{
			Driver: wd,
			T:      t,
		})
	}
}

var matchPat string
var matchRe *regexp.Regexp

func matchString(pat, str string) (result bool, err error) {
	if matchRe == nil || matchPat != pat {
		matchPat = pat
		matchRe, err = regexp.Compile(matchPat)
		if err != nil {
			return
		}
	}
	return matchRe.MatchString(str), nil
}

func Main(tests map[string]func(*T)) {
	flag.Parse()

	browserList := strings.Split(*browserSpec, ",")
	for _, browser := range browserList {
		browsers = append(browsers, strings.TrimSpace(browser))
	}

	internalTests := make([]testing.InternalTest, 0, len(browsers)*len(tests))
	for name, test := range tests {
		for _, browser := range browsers {
			internalTests = append(internalTests, testing.InternalTest{
				Name: name + strings.Title(browser),
				F:    makeTestFunc(browser, test),
			})
		}
	}
	testing.Main(matchString, internalTests, nil, nil)
}
