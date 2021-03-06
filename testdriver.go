// Package testdriver implements a library for writing WebDriver tests
// and provides a familar API for those used to the standard go
// "testing" package.
package testdriver

import (
	"flag"
	"fmt"
	"github.com/daaku/go.chromedriver"
	"github.com/daaku/go.flagconfig"
	"github.com/daaku/go.testdriver/testing"
	"github.com/tebeka/selenium"
	"os"
	"regexp"
	"strings"
	"time"
)

var (
	webdriverRemoteUrl = flag.String(
		"testdriver.remote",
		"http://localhost:8080/wd/hub",
		"Remote webdriver URL.")
	webdriverQuit = flag.Bool(
		"testdriver.quit", true,
		"Determines if the browser will be quit at the end of a test, "+
			"even if successful.")
	webdriverProxy = flag.String(
		"testdriver.proxy",
		"",
		"Webdriver browser proxy.")
	webdriverImplicitWaitTimeout = flag.Duration(
		"testdriver.timeout.implicit-wait", 120*time.Second,
		"Webdriver implicit wait timeout.")
	webdriverAsyncScriptTimeout = flag.Duration(
		"testdriver.timeout.async-script", 120*time.Second,
		"Webdriver async script timeout.")
	browserSpec = flag.String(
		"testdriver.browsers",
		"firefox,chrome,iexplorer",
		"List of browser to run against.")
	internalChromeMode = flag.Bool(
		"testdriver.internal-chrome",
		false,
		"Enable the internal chromedriver providing a self contained environment.")
	quitOnFail = flag.Bool(
		"testdriver.quit-on-fail",
		false,
		"Will kill the browser even if the test fails.")
)

var browsers []string

func newRemote(browser string) (selenium.WebDriver, func(), error) {
	caps := selenium.Capabilities{
		"browserName":    browser,
		"acceptSslCerts": true,
	}
	if *webdriverProxy != "" {
		proxy := make(map[string]string)
		proxy["proxyType"] = "MANUAL"
		proxy["httpProxy"] = *webdriverProxy
		proxy["sslProxy"] = *webdriverProxy
		caps["proxy"] = proxy
	}

	remoteUrl := *webdriverRemoteUrl
	var internalChromeServer *chromedriver.Server
	if *internalChromeMode {
		var err error
		internalChromeServer, err = chromedriver.Start()
		if err != nil {
			return nil, nil, fmt.Errorf(
				"Error starting internal chrome driver: %s", err)
		}
		remoteUrl = internalChromeServer.URL()
	}

	wd, err := selenium.NewRemote(caps, remoteUrl)
	if err != nil {
		return nil, nil, fmt.Errorf(
			"Can't start session %s for browser %s", err, browser)
	}
	err = wd.SetAsyncScriptTimeout(*webdriverAsyncScriptTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("Can't set async script timeout %s", err)
	}
	err = wd.SetImplicitWaitTimeout(*webdriverImplicitWaitTimeout)
	if err != nil {
		return nil, nil, fmt.Errorf("Can't set implicit wait timeout %s", err)
	}

	quit := func() {
		if *webdriverQuit {
			wd.Quit()
			if internalChromeServer != nil {
				internalChromeServer.StopOrFatal()
			}
		}
	}

	return wd, quit, nil
}

func makeTestFunc(browser string, test func(*T)) func(*testing.T) {
	return func(t *testing.T) {
		t.Parallel()
		wd, quit, err := newRemote(browser)
		if err != nil {
			t.Fatalf("Failed to create remote: %s", err)
		}
		defer func() {
			if !t.Failed() || *quitOnFail {
				quit()
			}
		}()
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
	flagconfig.Parse()

	if *internalChromeMode {
		browsers = []string{"chrome"}
	} else {
		browserList := strings.Split(*browserSpec, ",")
		for _, browser := range browserList {
			browsers = append(browsers, strings.TrimSpace(browser))
		}
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
	testOk := testing.Main(matchString, internalTests)
	if !testOk {
		os.Exit(1)
	}
}
