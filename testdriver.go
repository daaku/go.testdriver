package testdriver

import (
	"flag"
	"fmt"
	"github.com/nshah/selenium"
	"github.com/nshah/go.chromedriver"
	"regexp"
	"strings"
	"github.com/nshah/go.testdriver/testing"
	"time"
	"log"
	"os"
)

var (
	webdriverRemoteUrl = flag.String(
		"testdriver.remote",
		"http://localhost:8080/wd/hub",
		"Remote webdriver URL.")
	webdriverQuit = flag.Bool(
		"testdriver.quit", true,
		"Determines if the browser will be quit at the end of a test.")
	webdriverProxy = flag.String(
		"testdriver.proxy",
		"",
		"Webdriver browser proxy.")
	webdriverImplicitWaitTimeout = flag.Duration(
		"testdriver.timeout.implicit-wait", 60*time.Second,
		"Webdriver implicit wait timeout.")
	webdriverAsyncScriptTimeout = flag.Duration(
		"testdriver.timeout.async-script", 60*time.Second,
		"Webdriver async script timeout.")
	browserSpec = flag.String(
		"testdriver.browsers",
		"firefox,chrome,iexplorer",
		"List of browser to run against.")
	internalChromeMode = flag.Bool(
		"testdriver.internal-chrome",
		false,
		"Enable the internal chromedriver providing a self contained environment.")
)

var browsers []string

func newRemote(browser string) (selenium.WebDriver, error) {
	caps := selenium.Capabilities{
		"browserName": browser,
		"acceptSslCerts": true,
	}
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

	var internalChromeServer *chromedriver.Server
	var err error
	if *internalChromeMode {
		internalChromeServer, err = chromedriver.Start()
		*browserSpec = "chrome"
		if err != nil {
			log.Fatalf("Error starting internal chrome driver: %s", err)
		}
		*webdriverRemoteUrl = internalChromeServer.URL()
	}

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
	testOk := testing.Main(matchString, internalTests)
	if internalChromeServer != nil && *webdriverQuit {
		if testOk {
			internalChromeServer.StopOrFatal()
		} else {
			log.Print(
				"chromedriver was kept running to allow investigating failed tests.")
		}
	}
	if !testOk {
		os.Exit(1)
	}
}
